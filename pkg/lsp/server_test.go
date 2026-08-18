package lsp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/bovinemagnet/asciidoc-antora-ls/pkg/position"
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

func TestServer_Initialize_PositionEncoding(t *testing.T) {
	tests := []struct {
		name    string
		offered []position.Encoding
		want    position.Encoding
	}{
		{name: "defaults to UTF-16", want: position.UTF16},
		{name: "selects UTF-8", offered: []position.Encoding{position.UTF8, position.UTF16}, want: position.UTF8},
		{name: "honors client order", offered: []position.Encoding{position.UTF16, position.UTF8}, want: position.UTF16},
		{name: "skips unsupported encoding", offered: []position.Encoding{"utf-32", position.UTF8}, want: position.UTF8},
		{name: "falls back to UTF-16", offered: []position.Encoding{"utf-32"}, want: position.UTF16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer()
			params := &protocol.InitializeParams{ClientInfo: &protocol.ClientInfo{Name: "test-client"}}

			result, err := server.Initialize(context.Background(), params, tt.offered...)
			if err != nil {
				t.Fatalf("Initialize failed: %v", err)
			}
			if got := result.Capabilities.PositionEncoding; got != tt.want {
				t.Errorf("position encoding = %q, want %q", got, tt.want)
			}

			encoded, err := json.Marshal(result)
			if err != nil {
				t.Fatalf("marshal initialize result: %v", err)
			}
			if !strings.Contains(string(encoded), `"positionEncoding":"`+string(tt.want)+`"`) {
				t.Errorf("initialize result does not advertise %q: %s", tt.want, encoded)
			}
		})
	}
}

func TestClientPositionEncodings(t *testing.T) {
	params := json.RawMessage(`{"capabilities":{"general":{"positionEncodings":["utf-8","utf-16"]}}}`)

	encodings, err := clientPositionEncodings(params)
	if err != nil {
		t.Fatalf("clientPositionEncodings failed: %v", err)
	}
	if len(encodings) != 2 || encodings[0] != position.UTF8 || encodings[1] != position.UTF16 {
		t.Fatalf("unexpected position encodings: %v", encodings)
	}
}

func TestServer_Completion_UsesNegotiatedPositionEncoding(t *testing.T) {
	server := NewServer()
	ctx := context.Background()
	initializeParams := &protocol.InitializeParams{ClientInfo: &protocol.ClientInfo{Name: "test-client"}}
	if _, err := server.Initialize(ctx, initializeParams, position.UTF8); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	const documentURI = "file:///test.adoc"
	if err := server.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  documentURI,
			Text: "é xref:",
		},
	}); err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	result, err := server.Completion(ctx, &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
			Position:     protocol.Position{Character: 8},
		},
	})
	if err != nil {
		t.Fatalf("Completion failed: %v", err)
	}
	for _, item := range result.Items {
		if item.Label == "xref" {
			return
		}
	}
	t.Fatal("expected xref completion using negotiated UTF-8 position")
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
