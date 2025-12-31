SERVICE := holdings
GIT_VERSION := $(shell git describe --tags --dirty --match "${SERVICE}/*" 2>/dev/null || echo dev)
GIT_REVISION := $(shell git rev-parse --short HEAD 2>/dev/null || echo local)
SERVICE_VERSION := $(GIT_VERSION:$(SERVICE)/%=%)
DOCKER_BUILD_IMAGE := monay/$(SERVICE)
DOCKER_BUILD_VERSION := $(SERVICE_VERSION)
BUF_VERSION := 1.47.2
CGO_ENABLED ?= 0
GO_BUILD_SRC := ./cmd
GO_BUILD_TARGET := ./dist/$(SERVICE)
GO_BUILD_MODULE := github.com/levisegal/monay/services/$(SERVICE)/version
GO_BUILD_TAGS ?=
GO_BUILD_FLAGS := -trimpath
GO_BUILD_LDFLAG_DEFS = -X $(GO_BUILD_MODULE).Version=$(SERVICE_VERSION)
GO_BUILD_LDFLAG_DEFS += -X $(GO_BUILD_MODULE).Revision=$(GIT_REVISION)
GO_BUILD_LDFLAGS := "-s -w $(GO_BUILD_LDFLAG_DEFS)"

# C cross-compiler for CGO cross-compiling linux/arm64
CC_LINUX_ARM64 ?= aarch64-linux-gnu-gcc

# ------------------------------------------------------------------------------

.PHONY: version test mkdist dist clean

version:
	@echo $(SERVICE_VERSION)

test:
	@echo "Running tests"
	@go test -coverprofile=.coverage.out $(shell go list ./... | grep -Ev '/(gen|mocks|database/sql)/')
	@go tool cover -html=.coverage.out -o .coverage.html
	@rm -f .coverage.out

dist: mkdist build

mkdist:
	@mkdir -p dist
	@rm -f dist/*

clean:
	@rm -f dist/*

# --- go build -----------------------------------------------------------------

.PHONY: build build.linux-arm64 build.linux-amd64 build.darwin-arm64

build: mkdist build.darwin-arm64

build.linux-arm64:
	@CC=$(CC_LINUX_ARM64) CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -tags $(GO_BUILD_TAGS) $(GO_BUILD_FLAGS) -ldflags $(GO_BUILD_LDFLAGS) -o $(GO_BUILD_TARGET)-arm64 $(GO_BUILD_SRC)

build.linux-amd64:
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags $(GO_BUILD_TAGS) $(GO_BUILD_FLAGS) -ldflags $(GO_BUILD_LDFLAGS) -o $(GO_BUILD_TARGET)-amd64 $(GO_BUILD_SRC)

build.darwin-arm64:
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=arm64 go build -tags $(GO_BUILD_TAGS) $(GO_BUILD_FLAGS) -ldflags $(GO_BUILD_LDFLAGS) -o $(GO_BUILD_TARGET)-darwin-arm64 $(GO_BUILD_SRC)

# --- docker -------------------------------------------------------------------

.PHONY: docker.build docker.push docker.buildx.create

docker.build:
	@docker build -f build/Dockerfile -t $(DOCKER_BUILD_IMAGE):$(DOCKER_BUILD_VERSION) .

docker.push: docker.buildx.create
	@docker buildx build \
		--builder mybuilder \
		--platform=linux/arm64,linux/amd64 \
		-f build/Dockerfile \
		-t $(DOCKER_BUILD_IMAGE):$(DOCKER_BUILD_VERSION) \
		--push \
		.

docker.buildx.create:
	@docker buildx use mybuilder || \
	 docker buildx create \
	 	--name mybuilder \
		--driver docker-container \
		--bootstrap \
		--use

# --- tools --------------------------------------------------------------------

.PHONY: proto.generate run

proto.generate:
	@if [ -z $(shell buf --version | grep $(BUF_VERSION)) ]; then \
		echo "buf version is not $(BUF_VERSION)"; \
		exit 1; \
	fi
	@if [ -n "$(CI)" ]; then \
		REPLY="y"; \
	else \
		read -p "Do you want to refresh all the *.go files in the gen/api directory? [y/N] " -n 1 -r; \
		echo; \
	fi; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo "Deleting all the *.go files in the gen/api directory"; \
		find gen/api/ -name '*.go' -exec rm -f {} + 2>/dev/null || true; \
	fi
	@echo "Generating the *.go files in the gen/api directory"
	@buf generate ../../api

run:
	@go run $(GO_BUILD_SRC) server
