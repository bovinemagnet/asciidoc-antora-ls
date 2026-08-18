.PHONY: build test clean install lint

# Build the language server
build:
	go build -o bin/asciidoc-antora-ls ./cmd/asciidoc-antora-ls

# Install the language server
install:
	go install ./cmd/asciidoc-antora-ls

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint the code
lint:
	go vet ./...
	go fmt ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Run the language server
run: build
	./bin/asciidoc-antora-ls

# Format code
fmt:
	go fmt ./...

# Check for common errors
vet:
	go vet ./...
