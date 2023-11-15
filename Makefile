#GOLANGCI_LINT_VERSION := "v1.55.2" # Optional configuration to pinpoint golangci-lint version.

build:
	goreleaser build --snapshot --clean

## Run tests
test: test-unit
