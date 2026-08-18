package lsp

import (
	"context"
	"testing"

	"go.lsp.dev/protocol"
)

func TestNewServer(t *testing.T) {
	server := NewServer()
	if server == nil {
		t.Fatal("NewServer() returned nil")
	}
	if server.documents == nil {
		t.Error("Server documents map is nil")
	}
	if server.parser == nil {
		t.Error("Server parser is nil")
	}
	if server.antora == nil {
		t.Error("Server antora analyzer is nil")
	}
}

func TestServer_Initialize(t *testing.T) {
	server := NewServer()
	ctx := context.Background()

	params := &protocol.InitializeParams{
		ClientInfo: &protocol.ClientInfo{
			Name:    "test-client",
			Version: "1.0.0",
		},
	}

	result, err := server.Initialize(ctx, params)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if result == nil {
		t.Fatal("Initialize result is nil")
	}

	if result.ServerInfo == nil {
		t.Error("ServerInfo is nil")
	} else {
		if result.ServerInfo.Name != "asciidoc-antora-ls" {
			t.Errorf("Expected server name 'asciidoc-antora-ls', got '%s'", result.ServerInfo.Name)
		}
	}

	// Check capabilities
	if result.Capabilities.HoverProvider != true {
		t.Error("HoverProvider should be true")
	}
	if result.Capabilities.CompletionProvider == nil {
		t.Error("CompletionProvider should not be nil")
	}
	if result.Capabilities.DefinitionProvider != true {
		t.Error("DefinitionProvider should be true")
	}
}

func TestServer_DidOpen(t *testing.T) {
	server := NewServer()
	ctx := context.Background()

	params := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///test.adoc",
			Version: 1,
			Text:    "= Test Document\n\nContent here.",
		},
	}

	err := server.DidOpen(ctx, params)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Check that document was stored
	server.mu.RLock()
	doc, exists := server.documents["file:///test.adoc"]
	server.mu.RUnlock()

	if !exists {
		t.Fatal("Document was not stored")
	}
	if doc.Content != params.TextDocument.Text {
		t.Error("Document content does not match")
	}
	if doc.Version != 1 {
		t.Errorf("Expected version 1, got %d", doc.Version)
	}
}

func TestServer_DidChange(t *testing.T) {
	server := NewServer()
	ctx := context.Background()

	// First open a document
	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///test.adoc",
			Version: 1,
			Text:    "Initial content",
		},
	}
	if err := server.DidOpen(ctx, openParams); err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Now change it
	changeParams := &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: "file:///test.adoc",
			},
			Version: 2,
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{
			{
				Text: "Changed content",
			},
		},
	}

	err := server.DidChange(ctx, changeParams)
	if err != nil {
		t.Fatalf("DidChange failed: %v", err)
	}

	// Check that document was updated
	server.mu.RLock()
	doc := server.documents["file:///test.adoc"]
	server.mu.RUnlock()

	if doc.Content != "Changed content" {
		t.Errorf("Expected 'Changed content', got '%s'", doc.Content)
	}
	if doc.Version != 2 {
		t.Errorf("Expected version 2, got %d", doc.Version)
	}
}

func TestServer_DidClose(t *testing.T) {
	server := NewServer()
	ctx := context.Background()

	// First open a document
	openParams := &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:     "file:///test.adoc",
			Version: 1,
			Text:    "Test content",
		},
	}
	if err := server.DidOpen(ctx, openParams); err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Now close it
	closeParams := &protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: "file:///test.adoc",
		},
	}

	err := server.DidClose(ctx, closeParams)
	if err != nil {
		t.Fatalf("DidClose failed: %v", err)
	}

	// Check that document was removed
	server.mu.RLock()
	_, exists := server.documents["file:///test.adoc"]
	server.mu.RUnlock()

	if exists {
		t.Error("Document should have been removed")
	}
}
