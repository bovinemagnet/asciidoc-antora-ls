package asciidoc

import (
	"testing"

	"github.com/bovinemagnet/asciidoc-antora-ls/pkg/position"
	"go.lsp.dev/protocol"
)

func TestParser_GetHoverInfo_Heading(t *testing.T) {
	parser := NewParser()
	content := "= Main Title\n\nSome content\n\n== Section Heading"

	// Test main title
	info := parser.GetHoverInfo(content, protocol.Position{Line: 0, Character: 0})
	if info == "" {
		t.Error("Expected hover info for main title, got empty string")
	}
	if info != "**Heading Level 1**\n\nMain Title" {
		t.Errorf("Expected heading level 1 info, got: %s", info)
	}

	// Test section heading
	info = parser.GetHoverInfo(content, protocol.Position{Line: 4, Character: 0})
	if info == "" {
		t.Error("Expected hover info for section heading, got empty string")
	}
}

func TestParser_GetHoverInfo_Attribute(t *testing.T) {
	parser := NewParser()
	content := ":toc: left\n:author: John Doe"

	info := parser.GetHoverInfo(content, protocol.Position{Line: 0, Character: 0})
	if info == "" {
		t.Error("Expected hover info for attribute, got empty string")
	}
	if info != "**Document Attribute**\n\nName: `toc`\n\nValue: `left`" {
		t.Errorf("Expected attribute info, got: %s", info)
	}
}

func TestParser_GetCompletions(t *testing.T) {
	parser := NewParser()
	content := ":toc"

	items := parser.GetCompletions(content, protocol.Position{Line: 0, Character: 4})
	if len(items) == 0 {
		t.Error("Expected attribute completions, got none")
	}

	// Check that we have common attributes
	hasAuthor := false
	for _, item := range items {
		if item.Label == "author" {
			hasAuthor = true
			break
		}
	}
	if !hasAuthor {
		t.Error("Expected 'author' in attribute completions")
	}
}

func TestParser_GetDocumentSymbols(t *testing.T) {
	parser := NewParser()
	content := "= Main Title\n\n== Section 1\n\n=== Subsection 1.1\n\n== Section 2"

	symbols := parser.GetDocumentSymbols(content)
	if len(symbols) != 4 {
		t.Errorf("Expected 4 symbols, got %d", len(symbols))
	}

	if symbols[0].Name != "Main Title" {
		t.Errorf("Expected first symbol to be 'Main Title', got '%s'", symbols[0].Name)
	}
}

func TestParser_GetCompletions_BlockDelimiters(t *testing.T) {
	parser := NewParser()
	content := ""

	items := parser.GetCompletions(content, protocol.Position{Line: 0, Character: 0})
	if len(items) == 0 {
		t.Error("Expected block delimiter completions, got none")
	}
}

func TestParser_GetCompletions_PositionEncoding(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		encoding  position.Encoding
		character uint32
	}{
		{name: "UTF-16 Latin", content: "é xref:", encoding: position.UTF16, character: 7},
		{name: "UTF-16 CJK", content: "界 xref:", encoding: position.UTF16, character: 7},
		{name: "UTF-16 astral", content: "😀 xref:", encoding: position.UTF16, character: 8},
		{name: "UTF-8 Latin", content: "é xref:", encoding: position.UTF8, character: 8},
		{name: "out of range", content: "é xref:", encoding: position.UTF16, character: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			parser.SetPositionEncoding(tt.encoding)

			items := parser.GetCompletions(tt.content, protocol.Position{Character: tt.character})
			if !hasCompletion(items, "xref") {
				t.Errorf("expected xref completion for %q at character %d", tt.content, tt.character)
			}
		})
	}
}

func TestParser_GetDocumentSymbols_PositionEncoding(t *testing.T) {
	parser := NewParser()
	parser.SetPositionEncoding(position.UTF16)

	symbols := parser.GetDocumentSymbols("= Hé😀")
	if len(symbols) != 1 {
		t.Fatalf("expected one symbol, got %d", len(symbols))
	}
	if got := symbols[0].Range.End.Character; got != 6 {
		t.Errorf("UTF-16 range end = %d, want 6", got)
	}

	parser.SetPositionEncoding(position.UTF8)
	symbols = parser.GetDocumentSymbols("= Hé😀")
	if got := symbols[0].Range.End.Character; got != 9 {
		t.Errorf("UTF-8 range end = %d, want 9", got)
	}
}

func hasCompletion(items []protocol.CompletionItem, label string) bool {
	for _, item := range items {
		if item.Label == label {
			return true
		}
	}
	return false
}
