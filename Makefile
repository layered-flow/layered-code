.PHONY: all build test clean download-ripgrep build-all

# Default target
all: download-ripgrep build

# Download ripgrep binaries for all platforms
download-ripgrep:
	@echo "Downloading ripgrep binaries..."
	@./third-party/ripgrep/download.sh

# Build for current platform
build: download-ripgrep
	go build -mod=mod -o layered-code ./cmd/layered_code/

# Build for all platforms
build-all: download-ripgrep
	@echo "Building for all platforms..."
	# macOS Intel
	GOOS=darwin GOARCH=amd64 go build -mod=mod -o dist/layered-code-darwin-amd64 ./cmd/layered_code/
	# macOS Apple Silicon
	GOOS=darwin GOARCH=arm64 go build -mod=mod -o dist/layered-code-darwin-arm64 ./cmd/layered_code/
	# Linux x64
	GOOS=linux GOARCH=amd64 go build -mod=mod -o dist/layered-code-linux-amd64 ./cmd/layered_code/
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build -mod=mod -o dist/layered-code-linux-arm64 ./cmd/layered_code/
	# Windows x64
	GOOS=windows GOARCH=amd64 go build -mod=mod -o dist/layered-code-windows-amd64.exe ./cmd/layered_code/

# Run tests
test:
	go test -mod=mod ./...

# Clean build artifacts
clean:
	rm -f layered-code
	rm -rf dist/
	rm -rf third-party/ripgrep/*/

# Install locally
install: build
	cp layered-code /usr/local/bin/