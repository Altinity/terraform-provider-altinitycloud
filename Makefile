PROVIDER_BIN:=$(shell basename `git rev-parse --show-toplevel`)
PROVIDER_NAME:=$(shell echo $(PROVIDER_BIN) | sed 's/terraform-provider-//g')
PROVIDER_DIRECTORY:=~/.terraform.d/plugins/local/altinity/${PROVIDER_NAME}
DEFAULT_GRAPHQL_SCHEMA_FILE:=internal/sdk/client/graphql.schema
DEFAULT_GRAPHQL_SCHEMA_URL:=https://anywhere.altinity.cloud/api/v1/graphql.schema

GRAPHQL_SCHEMA_URL:=$(or $(GRAPHQL_SCHEMA_URL),$(DEFAULT_GRAPHQL_SCHEMA_URL))
GRAPHQL_SCHEMA_FILE:=$(or $(GRAPHQL_SCHEMA_FILE),$(DEFAULT_GRAPHQL_SCHEMA_FILE))

VERSION:=0.0.1
OS:=darwin
ARCH:=amd64
LOCAL_DIRECTOY:=local

ifeq ($(shell uname -s), Linux)
  OS := linux
endif
# ifeq ($(shell uname -m), arm64)
#   ARCH=arm64
# endif

default: help

.PHONY: testacc
testacc:
	ALTINITYCLOUD_TEST_ENV_PREFIX="altinity" \
	ALTINITYCLOUD_API_TOKEN="eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhbHRpbml0eS5jbG91ZC5sb2NhbGhvc3QiLCJpYXQiOjE3MDk4MzgyMzUsInN1YiI6Im5hY2hvQGFsdGluaXR5LmNvbSIsImM6YWxpYXMiOiJhbHRpbml0eS10Zi10ZXN0In0.IeW1lA_3ONcrhgxGaUktLWK3LvgHLvelIGdYI2ZpA-YfigjffN2YpLQPDrLrORNTgAIPpdakfxAhhTyDOR8YBA" \
	ALTINITYCLOUD_API_URL="https://internal.altinity.cloud.localhost:7443" \
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

.PHONY: build
build:
	go build -o ${PROVIDER_BIN}

.PHONY: local
local: build
	chmod +x ${PROVIDER_BIN}
	mkdir -p ${PROVIDER_DIRECTORY}/${VERSION}/${OS}_${ARCH}/
	mv ${PROVIDER_BIN} ${PROVIDER_DIRECTORY}/${VERSION}/${OS}_${ARCH}/${PROVIDER_BIN}_v${VERSION}
	mkdir -p $(LOCAL_DIRECTOY)
	if [ ! -f $(LOCAL_DIRECTOY)/versions.tf ]; then echo 'terraform {\n  required_providers {\n    ${PROVIDER_NAME} = {\n      source  = "local/altinity/${PROVIDER_NAME}"\n      version = "${VERSION}"\n    }\n  }\n}' > $(LOCAL_DIRECTOY)/versions.tf; fi
	cd ${LOCAL_DIRECTOY} && rm -f .terraform.lock.hcl .terrform
	cd ${LOCAL_DIRECTOY} && TF_LOG=TRACE terraform init -upgrade

.PHONY: bump
bump:
	@if [ -z "$(type)" ]; then \
		echo "Error: 'type' not specified. Use 'make bump type=major', 'make bump type=minor', or 'make bump type=patch'."; \
		exit 1; \
	fi; \
	LATEST_VERSION=$$(git describe --tags `git rev-list --tags --max-count=1` | sed 's/^v//'); \
	MAJOR=$$(echo $$LATEST_VERSION | cut -d. -f1); \
	MINOR=$$(echo $$LATEST_VERSION | cut -d. -f2); \
	PATCH=$$(echo $$LATEST_VERSION | cut -d. -f3); \
	if [ "$(type)" = "major" ]; then \
		NEW_VERSION=v$$((MAJOR + 1)).0.0; \
	elif [ "$(type)" = "minor" ]; then \
		NEW_VERSION=v$$MAJOR.$$((MINOR + 1)).0; \
	elif [ "$(type)" = "patch" ]; then \
		NEW_VERSION=v$$MAJOR.$$MINOR.$$((PATCH + 1)); \
	else \
		echo "Invalid type: $(type). Use 'major', 'minor', or 'patch'."; \
		exit 1; \
	fi; \
	echo "New version: $$NEW_VERSION"; \
	git tag $$NEW_VERSION; \
	echo "New version tagged: $$NEW_VERSION"


.PHONY: sync
sync:
	@$(eval LATEST_VERSION=$(shell git describe --tags `git rev-list --tags --max-count=1` | sed 's/^v//'))
	@echo "Current Version: $(LATEST_VERSION)"
	@# Determine OS type
	@$(eval OS_TYPE=$(shell uname))
	@# Adjust sed command based on OS
	@if [ $(OS_TYPE) = "Darwin" ]; then \
		find ./docs -name "*.md" -exec sed -i '' "s/%%VERSION%%/$(LATEST_VERSION)/g" {} +; \
	else \
		find ./docs -name "*.md" -exec sed -i "s/%%VERSION%%/$(LATEST_VERSION)/g" {} +; \
	fi
	@echo "Updated Terraform files in './docs' directory to version $(LATEST_VERSION)"

.PHONY: docs
docs:
ifeq ($(OS), darwin)
	GOOS=darwin GOARCH=amd64 go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
else
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
endif

.PHONY: sdk
sdk:
	@echo "Fetching GraphQL schema from ${GRAPHQL_SCHEMA_URL} to ${GRAPHQL_SCHEMA_FILE}"
	curl -o ${GRAPHQL_SCHEMA_FILE} ${GRAPHQL_SCHEMA_URL}
	cd internal/sdk/client && go run github.com/Yamashou/gqlgenc

.PHONY: gen
gen: sdk docs

.PHONY: fmt
fmt:
	go fmt ./...
	terraform fmt -recursive

.PHONY: help
help:
	@echo "Available commands:"
	@echo
	@echo "build             - Build the provider binary. This compiles the provider's Go code into a binary executable."
	@echo "bump              - Bump version tags in Git. Use 'make bump type=[major|minor|patch]' to create a new version tag."
	@echo "docs              - Generate provider documentation. This uses terraform-plugin-docs to create documentation for the provider."
	@echo "fmt               - Format Terraform and Go code. This ensures that the code follows standard formatting conventions."
	@echo "gen               - Run SDK generation, version sync, and docs generation. This is a combined command that runs sdk, sync, and docs commands."
	@echo "local             - Build the provider and set up the local directory for testing. This is useful for local development and testing."
	@echo "sdk               - Re-sync the SDK client and models. This pulls the latest GraphQL schema and regenerates the client code."
	@echo "testacc           - Run acceptance tests. These are integration tests that use the Terraform binary to test real infrastructure."
	@echo "sync              - Fetch and update the current version in the 'example' directory. This syncs the version used in examples with the latest git tag."
	@echo "tool              - Run Go tools. This is a placeholder for any Go-based tools you might want to run as part of the build."
