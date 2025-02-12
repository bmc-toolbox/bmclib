EXAMPLES_DIR:=./examples
GO_EXAMPLES_GO_MAINS:=$(wildcard $(EXAMPLES_DIR)/*/main.go)
EXAMPLES:=$(basename $(notdir $(patsubst %/,%,$(dir $(GO_EXAMPLES_GO_MAINS)))))
EXAMPLES_BIN_DIR:=$(EXAMPLES_DIR)/bin

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | cut -d":" -f2,3 | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test
test: ## Run unit tests
	go test -v -covermode=atomic -race ./...

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

.PHONY: build-examples
build-examples: $(EXAMPLES)
	@echo "Done!"

.PHONY:$(EXAMPLES)
.SECONDEXPANSION:
$(EXAMPLES):
	@mkdir -p $(EXAMPLES_BIN_DIR)
	@echo "Building example: $@"
	@go build -o ./examples/bin/$@ ./examples/$@/main.go

.PHONY: clean
clean:
	rm -rf $(EXAMPLES_BIN_DIR)

-include lint.mk