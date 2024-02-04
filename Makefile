.PHONY: help
help: ## Display help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: lint
lint: ## Run linter
	golangci-lint run -v ./...

.PHONY:statictest
statictest: ## Run statictest
	go vet -vettool=$$(which statictest) ./...

.PHONY:clean
clean: ## Delete old binaries
	-rm -f ./cmd/gophermart/gophermart

.PHONY:build
build: ## Prepare binaries
	go build -C ./cmd/gophermart/ -o gophermart

.PHONY: test
test: build ## Run tests
	gophermarttest -test.v -test.run=^TestGophermart$ \
       	-gophermart-binary-path=cmd/gophermart/gophermart \
       	-gophermart-host=localhost \
       	-gophermart-port=8080 \
       	-gophermart-database-uri="postgresql://postgres:postgres@postgres/praktikum?sslmode=disable" \
       	-accrual-binary-path=cmd/accrual/accrual_darwin_arm64 \
       	-accrual-host=localhost \
       	-accrual-port=$$(random unused-port) \
       	-accrual-database-uri="postgresql://postgres:postgres@postgres/praktikum?sslmode=disable"
