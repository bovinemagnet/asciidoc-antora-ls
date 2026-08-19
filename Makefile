GO_FILES := $(shell find . -type f -name '*.go')
VERSION ?= 0.1.0
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null)
BUILD_DATE ?= $(shell git show -s --format=%cI HEAD 2>/dev/null)
LDFLAGS := -s -w \
	-X github.com/bovinemagnet/asciidoc-antora-ls/pkg/lsp.Version=$(VERSION) \
	-X github.com/bovinemagnet/asciidoc-antora-ls/pkg/lsp.Commit=$(COMMIT) \
	-X github.com/bovinemagnet/asciidoc-antora-ls/pkg/lsp.BuildDate=$(BUILD_DATE)

.PHONY: build install test test-coverage lint clean run fmt fmt-check vet

# Build the language server
build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/asciidoc-antora-ls ./cmd/asciidoc-antora-ls

# Install the language server
install:
	go install -trimpath -ldflags "$(LDFLAGS)" ./cmd/asciidoc-antora-ls

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint the code
lint: fmt-check vet

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Run the language server
run: build
	./bin/asciidoc-antora-ls

# Format code
fmt:
	gofmt -s -w $(GO_FILES)

# Check formatting without changing files
fmt-check:
	@unformatted="$$(gofmt -s -l $(GO_FILES))"; \
	if [ -n "$$unformatted" ]; then \
		echo "Code is not formatted. Run 'make fmt'"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

# Check for common errors
vet:
	go vet ./...
