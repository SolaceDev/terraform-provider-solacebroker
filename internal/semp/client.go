// terraform-provider-solacebroker
//
// Copyright 2023 Solace Corporation. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package semp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	ErrResourceNotFound = errors.New("Resource not found")
	ErrBadRequest       = errors.New("Bad request")
	ErrAPIUnreachable   = errors.New("SEMP API unreachable")
)

var cookieJar, _ = cookiejar.New(nil)

var firstRequest = true

type retryableTransport struct {
	transport http.RoundTripper
}

const RetryCount = 6

type Client struct {
	*http.Client
	url                string
	username           string
	password           string
	bearerToken        string
	retries            uint
	retryMinInterval   time.Duration
	retryMaxInterval   time.Duration
	requestMinInterval time.Duration
	rateLimiter        <-chan time.Time
}

var Cookies = map[string]*http.Cookie{}

type Option func(*Client)

func BasicAuth(username, password string) Option {
	return func(client *Client) {
		client.username = username
		client.password = password
	}
}

func BearerToken(bearerToken string) Option {
	return func(client *Client) {
		client.bearerToken = bearerToken
	}
}

func Retries(numRetries uint, retryMinInterval, retryMaxInterval time.Duration) Option {
	return func(client *Client) {
		client.retries = numRetries
		client.retryMinInterval = retryMinInterval
		client.retryMaxInterval = retryMaxInterval
	}
}

func RequestLimits(requestTimeoutDuration, requestMinInterval time.Duration) Option {
	return func(client *Client) {
		client.Client.Timeout = requestTimeoutDuration
		client.requestMinInterval = requestMinInterval
	}
}

func backoff(retries int) time.Duration {
	return time.Duration(math.Pow(2, float64(retries))) * time.Second
}

func shouldRetry(err error, resp *http.Response) bool {
	if err != nil {
		return true
	}

	if resp.StatusCode == http.StatusBadGateway ||
		resp.StatusCode == http.StatusServiceUnavailable ||
		resp.StatusCode == http.StatusGatewayTimeout {
		return true
	}

	return false
}

func drainBody(resp *http.Response) {
	if resp.Body != nil {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func (t *retryableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request body
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Send the request
	resp, err := t.transport.RoundTrip(req)

	// Retry logic
	retries := 0
	for shouldRetry(err, resp) && retries < RetryCount {
		// Wait for the specified backoff period
		tm := backoff(retries)
		fmt.Println("Retrying after sleeping for " + tm.String())
		time.Sleep(tm)

		// We're going to retry, consume any response to reuse the connection.
		if resp != nil {
			drainBody(resp)
		}

		// Clone the request body again
		if req.Body != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Retry the request
		resp, err = t.transport.RoundTrip(req)

		retries++
	}

	// Return the response
	return resp, err
}

func NewClient(url string, insecure_skip_verify bool, cookiejar http.CookieJar, options ...Option) *Client {
	transport := &retryableTransport{
		transport: http.DefaultTransport,
	}
	// transport := http.DefaultTransport.(*http.Transport)
	// transport.transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: insecure_skip_verify}
	httpClient := http.DefaultClient
	httpClient.Transport = transport
	httpClient.Jar = cookiejar
	httpClient.Timeout = time.Second * 120
	client := &Client{
		Client:           httpClient,
		url:              url,
		retries:          3,
		retryMinInterval: time.Second,
		retryMaxInterval: time.Second * 10,
	}
	for _, o := range options {
		o(client)
	}
	if client.requestMinInterval > 0 {
		client.rateLimiter = time.NewTicker(client.requestMinInterval).C
	} else {
		ch := make(chan time.Time)
		// closing the channel will make receiving from the channel non-blocking (the value received will be the
		//  zero value)
		close(ch)
		client.rateLimiter = ch
	}

	return client
}

func (c *Client) RequestWithBody(ctx context.Context, method, url string, body any) (map[string]any, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, method, c.url+url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	dumpData(ctx, fmt.Sprintf("%v to %v", request.Method, request.URL), data)
	rawBody, err := c.doRequest(ctx, request)
	if err != nil {
		return nil, err
	}
	return parseResponseAsObject(ctx, request, rawBody)
}

func (c *Client) doRequest(ctx context.Context, request *http.Request) ([]byte, error) {
	if !firstRequest {
		// the value doesn't matter, it is waiting for the value that matters
		<-c.rateLimiter
	} else {
		// only skip rate limiter for the first request
		firstRequest = false
	}
	if request.Method != http.MethodGet {
		request.Header.Set("Content-Type", "application/json")
	}
	// Prefer OAuth even if Basic Auth credentials provided
	if c.bearerToken != "" {
		// TODO: add log
		request.Header.Set("Authorization", "Bearer "+c.bearerToken)
	} else if c.username != "" {
		request.SetBasicAuth(c.username, c.password)
	} else {
		return nil, fmt.Errorf("either username or bearer token must be provided to access the broker")
	}
	attemptsRemaining := c.retries + 1
	retryWait := c.retryMinInterval
	var response *http.Response
	var err error

	// https://gosamples.dev/connection-reset-by-peer/
	// Also, remove unnecessary wait after attempts remaining elapsed
	// https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
	// https://medium.com/@kdthedeveloper/golang-http-retries-fbf7abacbe27

loop:
	for attemptsRemaining != 0 {
		response, err = c.Do(request)
		if err != nil {
			response = nil // make sure response is nil
		} else {
			switch response.StatusCode {
			case http.StatusOK:
				break loop
			case http.StatusBadRequest:
				break loop
			case http.StatusTooManyRequests:
				// ignore the too many requests body and any errors that happen while reading it
				_, _ = io.ReadAll(response.Body)
				// just continue
			default:
				// ignore errors while reading the error response body
				body, _ := io.ReadAll(response.Body)
				return nil, fmt.Errorf("unexpected status %v (%v) during %v to %v, body:\n%s", response.StatusCode, response.Status, request.Method, request.URL, body)
			}
		}
		tflog.Debug(ctx, fmt.Sprintf("===== Request failed, retrying in %v. Attempts remaining: %v =====", retryWait, attemptsRemaining))
		time.Sleep(retryWait)
		retryWait *= 2
		if retryWait > c.retryMaxInterval {
			retryWait = c.retryMaxInterval
		}
		attemptsRemaining--
	}
	if response == nil {
		return nil, err
	}
	rawBody, _ := io.ReadAll(response.Body)
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusBadRequest {
		return nil, fmt.Errorf("could not perform request: status %v (%v) during %v to %v, response body:\n%s", response.StatusCode, response.Status, request.Method, request.URL, rawBody)
	}
	if _, err := io.Copy(io.Discard, response.Body); err != nil {
		return nil, fmt.Errorf("could not perform request: status %v (%v) during %v to %v, response body:\n%s", response.StatusCode, response.Status, request.Method, request.URL, rawBody)
	}
	defer response.Body.Close()
	return rawBody, nil
}

func parseResponseAsObject(ctx context.Context, request *http.Request, dataResponse []byte) (map[string]any, error) {
	data := map[string]any{}
	err := json.Unmarshal(dataResponse, &data)
	if err != nil {
		return nil, fmt.Errorf("could not parse response body from %v to %v, response body was:\n%s", request.Method, request.URL, dataResponse)
	}
	dumpData(ctx, "response", dataResponse)
	rawData, ok := data["data"]
	if ok {
		// Valid data
		data, _ = rawData.(map[string]any)
		return data, nil
	} else {
		// Analize response metadata details
		rawData, ok = data["meta"]
		if ok {
			data, _ = rawData.(map[string]any)
			if data["responseCode"].(float64) == http.StatusOK {
				// this is valid response for delete
				return nil, nil
			}
			description := data["error"].(map[string]interface{})["description"].(string)
			status := data["error"].(map[string]interface{})["status"].(string)
			if status == "NOT_FOUND" {
				// resource not found is a special type we want to return
				return nil, fmt.Errorf("request failed from %v to %v, %v, %v, %w", request.Method, request.URL, description, status, ErrResourceNotFound)
			}
			tflog.Error(ctx, fmt.Sprintf("SEMP request returned %v, %v", description, status))
			return nil, fmt.Errorf("request failed for %v using %v, %v, %v", request.URL, request.Method, description, status)
		}
	}
	return nil, fmt.Errorf("could not parse response details from %v to %v, response body was:\n%s", request.Method, request.URL, dataResponse)
}

func parseResponseForGenerator(c *Client, ctx context.Context, basePath string, method string, request *http.Request, dataResponse []byte, appendToResult []map[string]any) ([]map[string]any, error) {
	data := map[string]any{}
	err := json.Unmarshal(dataResponse, &data)
	if err != nil {
		return nil, fmt.Errorf("could not parse response body from %v to %v, response body was:\n%s", request.Method, request.URL, dataResponse)
	}
	responseData := []map[string]any{}
	dumpData(ctx, "response", dataResponse)
	rawData, ok := data["data"]
	if ok {
		switch rawData.(type) {
		case []interface{}:
			responseDataRaw, _ := rawData.([]interface{})
			for _, t := range responseDataRaw {
				responseData = append(responseData, t.(map[string]any))
			}
		case map[string]interface{}:
			responseDataRaw, _ := rawData.(map[string]any)
			responseData = append(responseData, responseDataRaw)
		}
		metaData, hasMeta := data["meta"]
		appendToResult = append(appendToResult, responseData...)
		if hasMeta {
			pageData, hasPaging := metaData.(map[string]any)["paging"]
			if hasPaging {
				nextPage := fmt.Sprint(pageData.(map[string]any)["nextPageUri"])
				nextPageUrl := strings.Split(nextPage, basePath)
				print("..")
				return c.RequestWithoutBodyForGenerator(ctx, basePath, method, nextPageUrl[1], appendToResult)
			}
		}
		return appendToResult, nil
	} else {
		rawData, ok = data["meta"]
		if ok {
			data, _ = rawData.(map[string]any)
			responseData = append(responseData, data)
			errorCode, errorCodeExist := data["responseCode"]
			if errorCodeExist && fmt.Sprint(errorCode) == "400" {
				return responseData, ErrBadRequest
			}
			return responseData, ErrResourceNotFound
		}
	}
	return nil, nil
}

func (c *Client) RequestWithoutBody(ctx context.Context, method, url string) (map[string]interface{}, error) {
	request, err := http.NewRequestWithContext(ctx, method, c.url+url, nil)
	if err != nil {
		return nil, err
	}
	tflog.Debug(ctx, fmt.Sprintf("===== %v to %v =====", request.Method, request.URL))
	rawBody, err := c.doRequest(ctx, request)
	if err != nil {
		return nil, err
	}
	return parseResponseAsObject(ctx, request, rawBody)
}

func (c *Client) RequestWithoutBodyForGenerator(ctx context.Context, basePath string, method string, url string, appendToResult []map[string]any) ([]map[string]interface{}, error) {
	request, err := http.NewRequestWithContext(ctx, method, c.url+url, nil)
	if err != nil {
		return nil, err
	}
	rawBody, err := c.doRequest(ctx, request)
	if err != nil {
		return nil, err
	}
	return parseResponseForGenerator(c, ctx, basePath, method, request, rawBody, appendToResult)
}

func dumpData(ctx context.Context, tag string, data []byte) {
	var in any
	_ = json.Unmarshal(data, &in)
	out, _ := json.MarshalIndent(in, "", "\t")
	tflog.Debug(ctx, fmt.Sprintf("===== %v =====\n%s\n", tag, out))
}
