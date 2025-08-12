PROJECT_NAME := digestabot
BINARY := $(PROJECT_NAME)ctl
PKG := "github.com/chainguard-dev/platform-examples/$(PROJECT_NAME)"
VERSION := $(shell if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then git describe --exact-match --tags HEAD 2>/dev/null || echo "dev-$(shell git rev-parse --short HEAD)"; else echo "dev"; fi)
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

.PHONY: deps lint test coverage

deps: ## Get dependencies
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest

lint: deps ## Lint the files
	go vet
	gocyclo -over 15 -ignore "generated" ./

test: lint ## Run unittests
	go test -v ./...

coverage: ## Create test coverage report
	go test -cover ./...
	go test ./... -coverprofile=cover.out && go tool cover -html=cover.out -o coverage.html

build: ## Builds the binary on the current platform
	go build -a -ldflags "-w -X '$(PKG)/cmd.Version=$(VERSION)'" -o $(BINARY)

docs: build ## Builds the cli documentation
	mkdir -p docs
	./digestabotctl docs

clean: ## Reset everything
	docker run --rm -v ./output:/out alpine rm -rf /out/*
	git clean -fd
	git clean -fx
	git reset --hard

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
