// Command asciidoc-antora-ls runs the AsciiDoc and Antora language server.
package main

import (
	"context"
	"log"
	"os"

	"github.com/bovinemagnet/asciidoc-antora-ls/pkg/lsp"
)

func main() {
	log.SetOutput(os.Stderr)

	server := lsp.NewServer()
	if err := server.Run(context.Background(), os.Stdin, os.Stdout); err != nil {
		log.Printf("language server failed: %v", err)
		os.Exit(1)
	}
}
