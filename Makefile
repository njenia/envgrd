.PHONY: build install clean test

# Get version from git tag, or use "dev" if no tag exists
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
# Remove 'v' prefix if present
VERSION := $(shell echo $(VERSION) | sed 's/^v//')

# Build flags
LDFLAGS := -s -w -X main.Version=$(VERSION)
BUILD_FLAGS := -ldflags="$(LDFLAGS)"

# Binary name
BINARY_NAME := envgrd

build:
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	CGO_ENABLED=1 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME) ./cmd/envgrd

install:
	@echo "Installing $(BINARY_NAME) version $(VERSION)..."
	CGO_ENABLED=1 go install $(BUILD_FLAGS) ./cmd/envgrd

clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe

test:
	go test ./...

# Build for all platforms (useful for testing)
build-all:
	@echo "Building for all platforms..."
	@echo "Linux amd64..."
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/envgrd
	@echo "Linux arm64..."
	GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/envgrd
	@echo "macOS amd64..."
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/envgrd
	@echo "macOS arm64..."
	GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/envgrd
	@echo "Windows amd64..."
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/envgrd
	@echo "Windows arm64..."
	GOOS=windows GOARCH=arm64 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-windows-arm64.exe ./cmd/envgrd

help:
	@echo "Available targets:"
	@echo "  build      - Build the binary with version from git tag"
	@echo "  install    - Install to GOPATH/bin"
	@echo "  clean      - Remove built binaries"
	@echo "  test       - Run tests"
	@echo "  build-all  - Build for all platforms"
	@echo ""
	@echo "Current version: $(VERSION)"

