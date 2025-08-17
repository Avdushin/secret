# Project settings
NAME = secret
# VERSION = $(shell git describe --tags --always --dirty)
VERSION = 0.1.0
BUILD_DIR = bin
GO_PACKAGE = ./cmd/$(NAME)

# Build flags
LDFLAGS = -ldflags="-s -w -X main.version=$(VERSION)"
GOBUILD = CGO_ENABLED=0 go build $(LDFLAGS)

# Platforms to build for
PLATFORMS = linux windows darwin
ARCHITECTURES = amd64 arm64

# Default target
.PHONY: default
default: build

# Run the application
.PHONY: run
run:
	go run $(GO_PACKAGE)

# Build for current platform
.PHONY: build
build:
	$(GOBUILD) -o $(BUILD_DIR)/$(NAME) $(GO_PACKAGE)

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

# Install dependencies
.PHONY: deps
deps:
	go mod download

# Build for all platforms
.PHONY: release
release: clean deps
	$(foreach GOOS, $(PLATFORMS), \
		$(foreach GOARCH, $(ARCHITECTURES), \
			$(shell export GOOS=$(GOOS); export GOARCH=$(GOARCH); \
				if [ "$(GOOS)" = "windows" ]; then \
					$(GOBUILD) -o $(BUILD_DIR)/$(NAME)-$(GOOS)-$(GOARCH).exe $(GO_PACKAGE); \
				else \
					$(GOBUILD) -o $(BUILD_DIR)/$(NAME)-$(GOOS)-$(GOARCH) $(GO_PACKAGE); \
				fi; \
			) \
		) \
	)

# Build for Linux
.PHONY: linux
linux: clean deps
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(NAME)-linux-amd64 $(GO_PACKAGE)
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(NAME)-linux-arm64 $(GO_PACKAGE)

# Build for Windows
.PHONY: windows
windows: clean deps
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(NAME)-windows-amd64.exe $(GO_PACKAGE)
	GOOS=windows GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(NAME)-windows-arm64.exe $(GO_PACKAGE)

# Build for macOS
.PHONY: darwin
darwin: clean deps
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(NAME)-darwin-amd64 $(GO_PACKAGE)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(NAME)-darwin-arm64 $(GO_PACKAGE)

# Build and install locally
.PHONY: install
install: build
	cp $(BUILD_DIR)/$(NAME) /usr/local/bin/$(NAME)

# Run tests
.PHONY: test
test:
	go test ./...

# Run tests with coverage
.PHONY: test-cover
test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Format code
.PHONY: fmt
fmt:
	go fmt ./...

# Check code quality
.PHONY: lint
lint:
	golangci-lint run