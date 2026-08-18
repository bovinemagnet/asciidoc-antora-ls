package lsp

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/jsonrpc2"
	"go.lsp.dev/protocol"
)

func TestServer_JSONRPCSession(t *testing.T) {
	fixture, err := os.ReadFile("../../examples/sample.adoc")
	if err != nil {
		t.Fatalf("read sample document: %v", err)
	}
	content := string(fixture)
	lines := strings.Split(content, "\n")

	var logs bytes.Buffer
	server := NewServer(WithLogger(log.New(&logs, "", 0)))
	serverSide, clientSide := net.Pipe()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	serverDone := make(chan error, 1)
	go func() {
		serverDone <- server.Run(ctx, serverSide, serverSide)
	}()

	client := jsonrpc2.NewConn(
		ctx,
		jsonrpc2.NewBufferedStream(clientSide, jsonrpc2.VSCodeObjectCodec{}),
		nil,
		jsonrpc2.SetLogger(log.New(&logs, "", 0)),
	)
	t.Cleanup(func() {
		if err := client.Close(); err != nil && !errors.Is(err, jsonrpc2.ErrClosed) {
			t.Errorf("close JSON-RPC client: %v", err)
		}
		if err := serverSide.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			t.Errorf("close server pipe: %v", err)
		}
	})

	var initializeResult InitializeResult
	if err := client.Call(ctx, "initialize", map[string]any{
		"processId":  0,
		"clientInfo": map[string]string{"name": "e2e-test"},
		"capabilities": map[string]any{
			"general": map[string]any{"positionEncodings": []string{"utf-8", "utf-16"}},
		},
	}, &initializeResult); err != nil {
		t.Fatalf("initialize request: %v", err)
	}
	if initializeResult.ServerInfo == nil || initializeResult.ServerInfo.Version != Version {
		t.Fatalf("server version = %#v, want %q", initializeResult.ServerInfo, Version)
	}
	capabilities := initializeResult.Capabilities
	if capabilities.HoverProvider != true || capabilities.CompletionProvider == nil || capabilities.DefinitionProvider != true || capabilities.DocumentSymbolProvider != true {
		t.Fatalf("missing advertised capabilities: %#v", capabilities)
	}

	if err := client.Notify(ctx, "initialized", protocol.InitializedParams{}); err != nil {
		t.Fatalf("initialized notification: %v", err)
	}

	const documentURI = "file:///docs/sample.adoc"
	if err := client.Notify(ctx, "textDocument/didOpen", protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        documentURI,
			LanguageID: "asciidoc",
			Version:    1,
			Text:       content,
		},
	}); err != nil {
		t.Fatalf("didOpen notification: %v", err)
	}

	var hover protocol.Hover
	if err := client.Call(ctx, "textDocument/hover", protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
			Position:     protocol.Position{Line: 0, Character: 2},
		},
	}, &hover); err != nil {
		t.Fatalf("hover request: %v", err)
	}
	if hover.Contents.Kind != protocol.Markdown || !strings.Contains(hover.Contents.Value, "Heading Level 1") {
		t.Fatalf("unexpected hover result: %#v", hover)
	}

	completionLine := lineContaining(t, lines, "try typing `xref:")
	var completions protocol.CompletionList
	if err := client.Call(ctx, "textDocument/completion", protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
			Position: protocol.Position{
				Line:      uint32(completionLine),
				Character: uint32(len(lines[completionLine])),
			},
		},
	}, &completions); err != nil {
		t.Fatalf("completion request: %v", err)
	}
	if len(completions.Items) == 0 {
		t.Fatal("completion request returned no items")
	}

	definitionLine := lineContaining(t, lines, "xref:other-page.adoc")
	var definitions []protocol.Location
	if err := client.Call(ctx, "textDocument/definition", protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
			Position: protocol.Position{
				Line:      uint32(definitionLine),
				Character: uint32(strings.Index(lines[definitionLine], "xref:") + len("xref:")),
			},
		},
	}, &definitions); err != nil {
		t.Fatalf("definition request: %v", err)
	}
	if len(definitions) != 1 {
		t.Fatalf("definition count = %d, want 1", len(definitions))
	}

	var symbols []protocol.DocumentSymbol
	if err := client.Call(ctx, "textDocument/documentSymbol", protocol.DocumentSymbolParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
	}, &symbols); err != nil {
		t.Fatalf("documentSymbol request: %v", err)
	}
	if len(symbols) == 0 || symbols[0].Name != "Example AsciiDoc Document" {
		t.Fatalf("unexpected document symbols: %#v", symbols)
	}

	if err := client.Notify(ctx, "textDocument/didChange", protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: documentURI},
			Version:                2,
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{{Text: content}},
	}); err != nil {
		t.Fatalf("didChange notification: %v", err)
	}
	if err := client.Notify(ctx, "textDocument/didSave", protocol.DidSaveTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
	}); err != nil {
		t.Fatalf("didSave notification: %v", err)
	}

	var ignoredResult any
	err = client.Call(ctx, "unknown/request", struct{}{}, &ignoredResult)
	var responseErr *jsonrpc2.Error
	if !errors.As(err, &responseErr) || responseErr.Code != jsonrpc2.CodeMethodNotFound {
		t.Fatalf("unknown request error = %v, want method-not-found", err)
	}
	if err := client.Notify(ctx, "unknown/notification", struct{}{}); err != nil {
		t.Fatalf("unknown notification: %v", err)
	}
	if err := client.Call(ctx, "textDocument/hover", protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
			Position:     protocol.Position{Line: 0, Character: 2},
		},
	}, &hover); err != nil {
		t.Fatalf("request after unknown notification: %v", err)
	}

	if err := client.Notify(ctx, "textDocument/didClose", protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: documentURI},
	}); err != nil {
		t.Fatalf("didClose notification: %v", err)
	}

	var shutdownResult any
	if err := client.Call(ctx, "shutdown", nil, &shutdownResult); err != nil {
		t.Fatalf("shutdown request: %v", err)
	}
	if err := client.Notify(ctx, "exit", struct{}{}); err != nil {
		t.Fatalf("exit notification: %v", err)
	}

	select {
	case err := <-serverDone:
		if err != nil {
			t.Fatalf("server stopped with error: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("server did not stop after exit notification")
	}

	if strings.Contains(logs.String(), "Document changed:") {
		t.Fatalf("default logs contain per-change message: %s", logs.String())
	}
}

func lineContaining(t *testing.T, lines []string, fragment string) int {
	t.Helper()
	for index, line := range lines {
		if strings.Contains(line, fragment) {
			return index
		}
	}
	t.Fatalf("sample document does not contain %q", fragment)
	return -1
}
