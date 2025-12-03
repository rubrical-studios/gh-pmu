.PHONY: build test lint clean install

# Build variables
BINARY_NAME=gh-pmu
VERSION?=dev
LDFLAGS=-ldflags "-X github.com/scooter-indie/gh-pmu/cmd.version=$(VERSION)"

# Go commands
GO=go
GOTEST=$(GO) test
GOBUILD=$(GO) build

# Build the binary
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

# Run tests
test:
	$(GOTEST) -v -race ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Install to GOPATH/bin
install:
	$(GO) install $(LDFLAGS) .

# Install as gh extension
install-extension: build
	gh extension install .

# Run all checks (test + lint)
check: test lint

# Format code
fmt:
	$(GO) fmt ./...

# Tidy dependencies
tidy:
	$(GO) mod tidy
