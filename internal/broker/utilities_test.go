package broker

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestClient(t *testing.T) {
	matrix := []struct {
		ParamUsername    string
		ParamPassword    string
		ParamBearertoken string
		ParamURL         string
		EnvUsername      string
		EnvPassword      string
		EnvBearertoken   string
		EnvURL           string
		Expected         string
	}{
		// Original test cases - URL always provided in config
		{"testuser", "testpassword", "testbearertoken", "https://example.com", "", "", "", "", "Cannot use Bearer token with basic authentication credentials"},
		{"", "testpassword", "testbearertoken", "https://example.com", "", "", "", "", "Cannot use Bearer token with basic authentication credentials"},
		{"testuser", "", "testbearertoken", "https://example.com", "", "", "", "", "Cannot use Bearer token with basic authentication credentials"},
		{"", "", "", "https://example.com", "testuser", "testpassword", "testbearertoken", "", "Cannot use Bearer token with basic authentication credentials"},
		{"", "", "", "https://example.com", "testuser", "", "testbearertoken", "", "Cannot use Bearer token with basic authentication credentials"},
		{"", "", "", "https://example.com", "", "testpassword", "testbearertoken", "", "Cannot use Bearer token with basic authentication credentials"},
		{"", "", "", "https://example.com", "", "", "", "", "Bearer token or basic authentication credentials must be provided"},
		{"testuser", "testpassword", "", "https://example.com", "", "", "", "", ""},
		{"", "", "testbearertoken", "https://example.com", "", "", "", "", ""},
		{"", "", "testbearertoken", "https://example.com", "", "", "testbearertoken", "", ""},
		{"testuser", "testpassword", "", "https://example.com", "", "", "testbearertoken", "", ""},
		{"testuser", "testpassword", "", "https://example.com", "testuser", "testpassword", "testbearertoken", "", ""},
		{"", "", "testbearertoken", "https://example.com", "testuser", "testpassword", "", "", ""},
		{"", "", "testbearertoken", "https://example.com", "testuser", "testpassword", "testbearertoken", "", ""},
		{"", "", "", "https://example.com", "", "", "testbearertoken", "", ""},
		{"", "", "", "https://example.com", "testuser", "testpassword", "", "", ""},
		{"testuser", "", "", "https://example.com", "", "", "", "", "Both username and password must be provided for basic authentication and cannot mix params and env vars"},
		{"", "testpassword", "", "https://example.com", "", "", "", "", "Both username and password must be provided for basic authentication and cannot mix params and env vars"},
		{"testuser", "", "", "https://example.com", "", "testpassword", "", "", "Both username and password must be provided for basic authentication and cannot mix params and env vars"},
		{"", "testpassword", "", "https://example.com", "testuser", "", "", "", "Both username and password must be provided for basic authentication and cannot mix params and env vars"},

		// New test cases for URL environment variable
		{"testuser", "testpassword", "", "", "", "", "", "https://env.example.com", ""},                           // URL from env var
		{"testuser", "testpassword", "", "https://config.example.com", "", "", "", "https://env.example.com", ""}, // Config URL takes precedence
		{"testuser", "testpassword", "", "", "", "", "", "", "Missing required provider attribute"},               // No URL provided
		{"", "", "", "", "", "", "testbearertoken", "https://env.example.com", ""},                                // Bearer token with URL from env
	}

	// Iterate over the test matrix
	for testNr, test := range matrix {
		// Set the environment variables
		os.Setenv("SOLACEBROKER_USERNAME", test.EnvUsername)
		os.Setenv("SOLACEBROKER_PASSWORD", test.EnvPassword)
		os.Setenv("SOLACEBROKER_BEARER_TOKEN", test.EnvBearertoken)
		os.Setenv("SOLACEBROKER_URL", test.EnvURL)

		// Create a providerData struct from the test matrix
		var username, password, bearertoken, url types.String
		if test.ParamUsername != "" {
			username = types.StringValue(test.ParamUsername)
		} else {
			username = types.StringNull()
		}
		if test.ParamPassword != "" {
			password = types.StringValue(test.ParamPassword)
		} else {
			password = types.StringNull()
		}
		if test.ParamBearertoken != "" {
			bearertoken = types.StringValue(test.ParamBearertoken)
		} else {
			bearertoken = types.StringNull()
		}
		if test.ParamURL != "" {
			url = types.StringValue(test.ParamURL)
		} else {
			url = types.StringNull()
		}
		providerData := &providerData{
			Username:    username,
			Password:    password,
			BearerToken: bearertoken,
			Url:         url,
		}
		_, diag := client(providerData)
		// Check if the actual value is equal to the expected value
		if diag != nil {
			summary := diag.Summary()
			if test.Expected != summary {
				t.Errorf("Test %d: expected %v but got %v", testNr, test.Expected, summary)
			}
		} else {
			if test.Expected != "" {
				t.Errorf("Test %d: expected %v but got nil diag", testNr, test.Expected)
			}
		}
	}
}
