#make

# Include .env file if it exists
-include .env

PKG_LIST := $(shell go list ./... | grep -v /vendor/)


.PHONY:
dep: ## Get the dependencies
	@go mod tidy

.PHONY:
vet: ## Run go vet
	@go vet ${PKG_LIST}

.PHONY:
fmt: ## Run gofmt
	@gofmt -w -l .

.PHONY:
test: ## Run unit tests
	@go test -short ${PKG_LIST}

.PHONY:
test-coverage: ## Run tests with coverage
	@mkdir -p reports
	@go test -short -coverprofile reports/cover.out ${PKG_LIST}
	@go tool cover -html reports/cover.out -o reports/cover.html

.PHONY: testacc
testacc: ## Run acceptance tests
	@TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

.PHONY:
generate-docs: dep ## Build the binary file
	@go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name solacebroker

.PHONY:
build: dep ## Build the binary file
	@go build -a -ldflags '-s -w -extldflags "-static"' -o terraform-provider-solacebroker

.PHONY:
install: dep ## Install the provider for dev use
	@go install -a

.PHONY:
clean: ## Remove previous build
	@rm -f reports/cover.html reports/cover.out terraform-provider-solacebroker

.PHONY:
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY:
generate-code: ## Generate latest code from SEMP API spec
	@if [ ! -d "broker-terraform-code-generator" ]; then \
		git clone https://github.com/SolaceDev/broker-terraform-code-generator.git; \
	fi
	@cd broker-terraform-code-generator && git pull ; \
	go mod tidy; \
	go install .; \
	ls ~/go/bin | grep broker-terraform-code-generator
	@cd internal/broker/generated; \
	rm ./*; \
	SEMP_V2_SWAGGER_CONFIG_EXTENDED_JSON="../../../ci/swagger_spec/$(shell ls ci/swagger_spec)" ~/go/bin/broker-terraform-code-generator software-provider all;
	@rm -rf broker-terraform-code-generator

.PHONY:
newbroker: ## Run a new broker container with a specified tag, usage: make newbroker [tag=<docker-tag>]
	$(eval tag := $(if $(tag),$(tag),$(DOCKER_TAG)))
	@echo "Running a new broker container with tag: $(tag)"
	@$(CONTAINER_ENGINE) kill solace >/dev/null 2>&1 || true ; $(CONTAINER_ENGINE) rm solace >/dev/null 2>&1 || true
	$(CONTAINER_ENGINE) run -d -p 8080:8080 -p 55554:55555 -p 8008:8008 -p 1883:1883 -p 8000:8000 -p 5672:5672 -p 9000:9000 -p 2222:2222 -p 1943:1943 --shm-size=1g --env username_admin_globalaccesslevel=$(BROKER_USERNAME) --env username_admin_password=$(BROKER_PASSWORD) --env system_scaling_maxconnectioncount="1000" --env system_scaling_maxkafkabridgecount="10" --name=solace docker.solacedev.ca/pubsubplus/solace-pubsub-standard:$(tag)

.PHONY:
fetch-swagger-spec: ## Fetch swagger spec from devserver, usage: make fetch-swagger-spec [schema=<vm|lm>] [version=<broker-version>]
	@if [ -z "$(DEVSERVER)" ] || [ -z "$(DEVSERVER_ACCOUNT)" ]; then \
		echo "Error: DEVSERVER and DEVSERVER_ACCOUNT must be set in .env file"; \
		exit 1; \
	fi
	$(eval schema := $(if $(schema),$(schema),vm))
	@if [ -z "$(version)" ]; then \
		echo "Error: version parameter is required. Usage: make fetch-swagger-spec version=<broker-version> [schema=<vm|lm>]"; \
		echo "Example: make fetch-swagger-spec version=10.11.1 schema=vm"; \
		exit 1; \
	fi
	@echo "Fetching swagger spec for version=$(version), schema=$(schema)"
	@echo "Checking remote path on $(DEVSERVER)..."
	@if ! ssh $(DEVSERVER_ACCOUNT)@$(DEVSERVER) "test -d /home/public/RND/loads/solcbr/$(version)/current/$(schema)_schema"; then \
		echo "Error: Directory /home/public/RND/loads/solcbr/$(version)/current/$(schema)_schema does not exist on $(DEVSERVER)"; \
		exit 1; \
	fi
	@echo "Resolving 'current' symlink to get load release version..."
	$(eval load_version := $(shell ssh $(DEVSERVER_ACCOUNT)@$(DEVSERVER) "readlink /home/public/RND/loads/solcbr/$(version)/current"))
	@if [ -z "$(load_version)" ]; then \
		echo "Error: Failed to resolve 'current' symlink"; \
		exit 1; \
	fi
	@echo "Load release version: $(load_version)"
	@echo "Cleaning ci/swagger_spec directory..."
	@rm -f ci/swagger_spec/*
	@echo "Copying swagger spec file from devserver..."
	@scp $(DEVSERVER_ACCOUNT)@$(DEVSERVER):/home/public/RND/loads/solcbr/$(version)/current/$(schema)_schema/semp-v2-swagger-config-extended.json ci/swagger_spec/semp-v2-swagger-config-extended.$(load_version).$(schema).json
	@echo "Successfully fetched: ci/swagger_spec/semp-v2-swagger-config-extended.$(load_version).$(schema).json"
	@ls -lh ci/swagger_spec/
