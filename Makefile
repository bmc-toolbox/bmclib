help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | cut -d":" -f2,3 | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test
test: ## Run unit tests
	go test -v -covermode=count ./...

.PHONY: cover
cover: ## Run unit tests with coverage report
	go test -coverprofile=coverage.txt ./...
	go tool cover -func=coverage.txt

.PHONY: all-tests
all-tests: test cover ## run all tests

.PHONY: all-checks
all-checks: lint ## run all formatters
	go mod tidy
	go vet ./...

-include lint.mk