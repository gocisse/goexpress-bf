BINARY_NAME=goexpress-api
MAIN_PACKAGE=./main.go

.PHONY: build run clean test coverage help setup-db

## build: Build the GoExpress application
build:
	@echo "🔨 Building GoExpress application..."
	go build -o $(BINARY_NAME) $(MAIN_PACKAGE)

## run: Run the GoExpress application
run:
	@echo "🚀 Running GoExpress application..."
	go run $(MAIN_PACKAGE)

## clean: Clean build files
clean:
	@echo "🧹 Cleaning..."
	go clean
	rm -f $(BINARY_NAME)

## test: Run tests
test:
	@echo "🧪 Running tests..."
	go test -v ./tests/...

## coverage: Run tests with coverage
coverage:
	@echo "📊 Running tests with coverage..."
	go test -coverprofile=coverage.out ./tests/...
	go tool cover -html=coverage.out

## deps: Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy

## setup-db: Setup GoExpress database
setup-db:
	@echo "🗄️ Setting up GoExpress database..."
	chmod +x setup-database.sh
	./setup-database.sh

## test-api: Test all API endpoints
test-api:
	@echo "🧪 Testing GoExpress API endpoints..."
	chmod +x test-endpoints.sh
	./test-endpoints.sh

## swagger: Generate Swagger documentation
swagger:
	@echo "📚 Generating Swagger documentation..."
	swag init -g main.go

## lint: Run linter
lint:
	@echo "🔍 Running linter..."
	golangci-lint run

## dev: Run in development mode with auto-reload
dev:
	@echo "🔄 Running in development mode..."
	air

## prod: Build and run in production mode
prod: build
	@echo "🚀 Running GoExpress in production mode..."
	./$(BINARY_NAME)

## install-tools: Install development tools
install-tools:
	@echo "🛠️ Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

## logs: Show application logs (if running as service)
logs:
	@echo "📋 Showing GoExpress logs..."
	sudo journalctl -u goexpress -f

## status: Check GoExpress service status
status:
	@echo "📊 Checking GoExpress service status..."
	sudo systemctl status goexpress

## restart: Restart GoExpress service
restart:
	@echo "🔄 Restarting GoExpress service..."
	sudo systemctl restart goexpress

## help: Show help
help:
	@echo "GoExpress Delivery Management API"
	@echo "================================="
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p\' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
