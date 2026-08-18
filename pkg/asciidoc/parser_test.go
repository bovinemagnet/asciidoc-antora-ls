package asciidoc

import (
	"testing"

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
