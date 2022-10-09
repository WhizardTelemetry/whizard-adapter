REPO ?= kubesphere
TAG ?= $(shell cat VERSION | tr -d " \t\n\r")

ADAPTER_IMG=${REPO}/whizard-adapter:${TAG}

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

all: build

##@ Development

generate: ## Generate swagger models.
	./hack/generate-swagger-models.sh

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...


##@ Build

build: ## Build binary.
	go build -o bin/adapter cmd/adapter.go

docker-build: ## Build docker image.
	go env -w GOPRIVATE=github.com/WhizardTelemetry
	go mod vendor
	docker build -t $(ADAPTER_IMG) -f Dockerfile .

##@ Deployment

