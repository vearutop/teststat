#GOLANGCI_LINT_VERSION := "v1.55.2" # Optional configuration to pinpoint golangci-lint version.

build:
	goreleaser build --snapshot --clean

## Run tests
tests: tests-cov-gen
	$(eval total=$(shell go tool cover -func=./coverage.out | grep total | grep -Eo '[0-9]+\.[0-9]+'))
	@echo "Current test coverage : $(total) %"

tests-cov-gen:
	@echo "  >  Running tests and generating coverage output ..."
	@go test ./... -coverprofile coverage.out -covermode count

tests-cov-report: tests-cov-gen
	go tool cover -html coverage.out -o coverage.html
