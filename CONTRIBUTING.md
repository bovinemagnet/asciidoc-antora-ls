# Contributing to asciidoc-antora-ls

Thank you for your interest in contributing to asciidoc-antora-ls! This document provides guidelines and instructions for contributing.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally
3. Create a new branch for your changes
4. Make your changes
5. Test your changes
6. Submit a pull request

## Development Setup

### Prerequisites

- Go 1.21 or later
- Make (optional but recommended)
- Git

### Building the Project

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/asciidoc-antora-ls.git
cd asciidoc-antora-ls

# Install dependencies
go mod download

# Build the project
make build

# Or without make
go build -o bin/asciidoc-antora-ls ./cmd/asciidoc-antora-ls
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Or without make
go test -v ./...
```

### Code Quality

Before submitting a pull request, please ensure your code passes all checks:

```bash
# Format code
make fmt

# Run static analysis
make vet

# Run both
make lint
```

## Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` to format your code
- Write clear, descriptive variable and function names
- Add comments for exported functions and types
- Keep functions focused and concise

## Testing Guidelines

- Write tests for new functionality
- Maintain or improve code coverage
- Test edge cases and error conditions
- Use table-driven tests where appropriate

Example test structure:

```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"basic case", "input", "output"},
        {"edge case", "", ""},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Feature(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## Pull Request Process

1. Update the README.md with details of significant changes
2. Add tests for new functionality
3. Ensure all tests pass
4. Update documentation as needed
5. Create a pull request with a clear description of the changes

### Pull Request Checklist

- [ ] Code follows the project's style guidelines
- [ ] All tests pass
- [ ] New tests added for new functionality
- [ ] Documentation updated
- [ ] Commit messages are clear and descriptive
- [ ] No merge conflicts

## Commit Messages

Write clear, concise commit messages:

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests when relevant

Example:
```
Add hover support for AsciiDoc includes

- Parse include directives
- Display file path in hover tooltip
- Add tests for include hover

Fixes #123
```

## Project Structure

```
.
├── cmd/
│   └── asciidoc-antora-ls/    # Main application entry point
├── pkg/
│   ├── asciidoc/              # AsciiDoc parsing and features
│   ├── antora/                # Antora-specific functionality
│   └── lsp/                   # LSP server implementation
├── examples/                  # Example files
└── .github/                   # GitHub workflows and configs
```

## Adding New Features

When adding a new LSP feature:

1. Add the handler method to `pkg/lsp/server.go`
2. Implement the feature logic in the appropriate package (`asciidoc` or `antora`)
3. Add tests for the new functionality
4. Update the README.md with the new capability
5. Add the method to the `Handle` function in `server.go`

## Reporting Bugs

When reporting bugs, please include:

- Your operating system and version
- Go version (`go version`)
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Any error messages or logs

## Feature Requests

Feature requests are welcome! Please open an issue with:

- A clear description of the feature
- Use cases for the feature
- Any relevant examples from other language servers
- Why this feature would be beneficial

## Questions?

If you have questions, feel free to:

- Open an issue for discussion
- Check existing issues and pull requests
- Review the README.md for documentation

## License

By contributing to asciidoc-antora-ls, you agree that your contributions will be licensed under the MIT License.
