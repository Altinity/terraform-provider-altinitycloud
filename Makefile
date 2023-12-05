PROVIDER_BIN:=$(shell basename `git rev-parse --show-toplevel`)
PROVIDER_NAME:=$(shell echo $(PROVIDER_BIN) | sed 's/terraform-provider-//g')
PROVIDER_DIRECTORY:=~/.terraform.d/plugins/local/altinity/${PROVIDER_NAME}
GRAPHQL_SCHEMA_URL?=https://anywhere.altinity.cloud/api/v1/graphql.schema
GRAPHQL_SCHEMA_FILE:=internal/sdk/client/graphql.schema

VERSION:=0.0.1
OS:=darwin
ARCH:=amd64
LOCAL_DIRECTOY:=local

ifeq ($(shell uname -s), Linux)
  OS := linux
endif
ifeq ($(shell uname -m), arm64)
  ARCH=arm64
endif

default: help

.PHONY: testacc
testacc:
	ALTINITYCLOUD_TEST_ENV_PREFIX="altinity" \
	ALTINITYCLOUD_API_TOKEN="eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJhbHRpbml0eS5jbG91ZC5sb2NhbGhvc3QiLCJpYXQiOjE2OTY5MzU1NTcsInN1YiI6Im5hY2hvQGFsdGluaXR5LmNvbSIsImM6YWxpYXMiOiJ0ZXN0In0.qAWOoui01smS-PQZ2j3a_H6HTX8JzThY-IidA5KspJYEzRoNYsJXf87kooXsqGMl74AtEpENgZrc0oe7hxXHDw" \
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

.PHONY: gen
gen: sdk docs

.PHONY: docs
docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

.PHONY: sdk
sdk:
	curl -o ${GRAPHQL_SCHEMA_FILE} ${GRAPHQL_SCHEMA_URL}
	cd internal/sdk/client && go run github.com/Yamashou/gqlgenc

.PHONY: fmt
fmt:
	go fmt ./...
	terraform fmt -recursive

.PHONY: help
help:
	@echo "Available commands:"
	@echo
	@echo "build             - Build the provider binary."
	@echo "docs              - Generate provider documentation"
	@echo "fmt               - Format Terraform and Go code."
	@echo "local             - Build the provider and and setup the local directoy for testing."
	@echo "sdk               - Re-sync sdk client and models"
	@echo "testacc           - Run acceptance tests (altinity internal usage)"
	@echo "tool              - Run go tools."
