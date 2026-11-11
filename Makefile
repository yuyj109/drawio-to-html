GO ?= go
GO_VERSION := 1.27.4
GOLANGCI_VERSION := v2.15.0
GOVULNCHECK_VERSION := v1.9.0

.PHONY: build install fmt fmt-check lint golangci-lint govulncheck test check clean check-go

check-go:
	@test "$$($(GO) env GOVERSION)" = "go$(GO_VERSION)" || (echo 'Go $(GO_VERSION) is required'; exit 1)

build: check-go
	$(GO) build -trimpath -o bin/drawio-to-html .

install: check-go
	$(GO) install .

fmt: check-go
	$(GO) fmt ./...

fmt-check: check-go
	@test -z "$$(gofmt -l main.go cmd internal)" || (echo 'Run make fmt'; exit 1)

lint: fmt-check golangci-lint govulncheck

golangci-lint: check-go
	$(GO) run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION) run

govulncheck: check-go
	$(GO) run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...

test: check-go
	$(GO) test -race ./...

check: lint test build

clean:
	rm -rf bin
