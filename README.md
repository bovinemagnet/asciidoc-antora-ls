# asciidoc-antora-ls

A Go-based Language Server Protocol (LSP) implementation for AsciiDoc with Antora support.

## Features

This language server provides IDE-like features for AsciiDoc files, with special support for Antora documentation projects:

### AsciiDoc Features
- **Document Synchronization**: Real-time document tracking and updates
- **Hover Information**: Contextual information for headings, attributes, and includes
- **Code Completion**: Intelligent completions for:
  - Document attributes (`:toc:`, `:author:`, etc.)
  - Block delimiters (example blocks, sidebar blocks, listing blocks)
  - Inline macros (`xref:`, `include::`)
- **Document Symbols**: Navigate document structure through headings
- **Go to Definition**: Jump to included files and cross-referenced pages

### Antora Features
- **Antora Attribute Completion**: Component and page-specific attributes
- **Cross-reference Support**: Navigate between Antora pages
- **Multi-format Xrefs**: Support for:
  - Same-module references: `xref:page.adoc[]`
  - Cross-module references: `xref:module:page.adoc[]`
  - Cross-component references: `xref:version@component:module:page.adoc[]`
  - Section references: `xref:page.adoc#section[]`

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/bovinemagnet/asciidoc-antora-ls.git
cd asciidoc-antora-ls

# Build the binary
make build

# Install to $GOPATH/bin
make install
```

### Binary Release

Download the latest release from the [releases page](https://github.com/bovinemagnet/asciidoc-antora-ls/releases).

## Usage

The language server communicates via JSON-RPC over stdin/stdout, following the Language Server Protocol specification.

### Running the Server

```bash
./bin/asciidoc-antora-ls
```

### Editor Integration

#### VS Code

Create or update `.vscode/settings.json`:

```json
{
  "asciidoc.languageServer": {
    "command": "/path/to/asciidoc-antora-ls",
    "args": []
  }
}
```

#### Neovim

Using `nvim-lspconfig`:

```lua
local lspconfig = require('lspconfig')
local configs = require('lspconfig.configs')

if not configs.asciidoc_antora then
  configs.asciidoc_antora = {
    default_config = {
      cmd = {'/path/to/asciidoc-antora-ls'},
      filetypes = {'asciidoc'},
      root_dir = lspconfig.util.root_pattern('antora.yml', '.git'),
      settings = {},
    },
  }
end

lspconfig.asciidoc_antora.setup{}
```

#### Emacs

Using `lsp-mode`:

```elisp
(require 'lsp-mode)

(add-to-list 'lsp-language-id-configuration '(asciidoc-mode . "asciidoc"))

(lsp-register-client
 (make-lsp-client :new-connection (lsp-stdio-connection "/path/to/asciidoc-antora-ls")
                  :major-modes '(asciidoc-mode)
                  :server-id 'asciidoc-antora-ls))
```

## Development

### Building

```bash
# Build the binary
make build

# Build and run
make run
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

### Linting

```bash
# Run go vet
make vet

# Format code
make fmt

# Run both
make lint
```

## Project Structure

```
.
├── cmd/
│   └── asciidoc-antora-ls/    # Main entry point
│       └── main.go
├── pkg/
│   ├── asciidoc/              # AsciiDoc parsing and features
│   │   ├── parser.go
│   │   └── parser_test.go
│   ├── antora/                # Antora-specific features
│   │   ├── analyzer.go
│   │   └── analyzer_test.go
│   └── lsp/                   # LSP server implementation
│       ├── server.go
│       └── server_test.go
├── Makefile                   # Build automation
├── go.mod                     # Go module definition
└── README.md                  # This file
```

## LSP Capabilities

The server implements the following LSP capabilities:

- `textDocument/didOpen` - Document opened notification
- `textDocument/didChange` - Document changed notification
- `textDocument/didSave` - Document saved notification
- `textDocument/didClose` - Document closed notification
- `textDocument/hover` - Hover information provider
- `textDocument/completion` - Completion provider
- `textDocument/definition` - Go to definition provider
- `textDocument/documentSymbol` - Document symbols provider

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Built using the [sourcegraph/jsonrpc2](https://github.com/sourcegraph/jsonrpc2) library
- LSP protocol types from [go.lsp.dev/protocol](https://pkg.go.dev/go.lsp.dev/protocol)
- Inspired by the [Language Server Protocol](https://microsoft.github.io/language-server-protocol/) specification
