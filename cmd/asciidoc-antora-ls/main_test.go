package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bovinemagnet/asciidoc-antora-ls/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
	"go.lsp.dev/protocol"
)

func TestRunVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if exitCode := run([]string{"--version"}, strings.NewReader(""), &stdout, &stderr); exitCode != 0 {
		t.Fatalf("run --version exit code = %d, stderr = %q", exitCode, stderr.String())
	}
	if got := strings.TrimSpace(stdout.String()); got != lsp.Version {
		t.Fatalf("version output = %q, want %q", got, lsp.Version)
	}
	if stderr.Len() != 0 {
		t.Fatalf("version wrote to stderr: %q", stderr.String())
	}
}

func TestRunStdioWritesOnlyJSONRPCFrames(t *testing.T) {
	stdout, stderr := runCLISession(t, []string{"--stdio"})

	assertJSONRPCFrames(t, stdout, 2)
	if strings.Contains(stderr, "Document changed:") {
		t.Fatalf("default log level emitted per-change message: %s", stderr)
	}
}

func TestRunLogFileAndDebugLevel(t *testing.T) {
	logPath := t.TempDir() + "/server.log"
	stdout, stderr := runCLISession(t, []string{"--stdio", "--log-file", logPath, "--log-level", "debug"})

	assertJSONRPCFrames(t, stdout, 2)
	if stderr != "" {
		t.Fatalf("logging was not redirected from stderr: %q", stderr)
	}
	contents, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.Contains(string(contents), "Document changed:") {
		t.Fatalf("debug log does not contain document change: %s", contents)
	}
}

func runCLISession(t *testing.T, args []string) ([]byte, string) {
	t.Helper()

	serverSide, clientSide := net.Pipe()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := make(chan int, 1)
	go func() {
		exitCode <- run(args, serverSide, io.MultiWriter(&stdout, serverSide), &stderr)
	}()

	client := jsonrpc2.NewConn(
		ctx,
		jsonrpc2.NewBufferedStream(clientSide, jsonrpc2.VSCodeObjectCodec{}),
		nil,
	)
	t.Cleanup(func() {
		if err := client.Close(); err != nil && !errors.Is(err, jsonrpc2.ErrClosed) {
			t.Errorf("close JSON-RPC client: %v", err)
		}
		if err := serverSide.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			t.Errorf("close server pipe: %v", err)
		}
	})

	var initializeResult lsp.InitializeResult
	if err := client.Call(ctx, "initialize", map[string]any{
		"processId":    0,
		"clientInfo":   map[string]string{"name": "cli-test"},
		"capabilities": map[string]any{},
	}, &initializeResult); err != nil {
		t.Fatalf("initialize request: %v", err)
	}
	if initializeResult.ServerInfo == nil || initializeResult.ServerInfo.Version != lsp.Version {
		t.Fatalf("server version = %#v, want %q", initializeResult.ServerInfo, lsp.Version)
	}
	if err := client.Notify(ctx, "initialized", protocol.InitializedParams{}); err != nil {
		t.Fatalf("initialized notification: %v", err)
	}

	const documentURI = "file:///cli-test.adoc"
	if err := client.Notify(ctx, "textDocument/didOpen", protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{URI: documentURI, Version: 1, Text: "= CLI Test"},
	}); err != nil {
		t.Fatalf("didOpen notification: %v", err)
	}
	if err := client.Notify(ctx, "textDocument/didChange", protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{URI: documentURI},
			Version:                2,
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{{Text: "= Changed"}},
	}); err != nil {
		t.Fatalf("didChange notification: %v", err)
	}

	var shutdownResult any
	if err := client.Call(ctx, "shutdown", nil, &shutdownResult); err != nil {
		t.Fatalf("shutdown request: %v", err)
	}
	if err := client.Notify(ctx, "exit", struct{}{}); err != nil {
		t.Fatalf("exit notification: %v", err)
	}

	select {
	case code := <-exitCode:
		if code != 0 {
			t.Fatalf("run exit code = %d, stderr = %q", code, stderr.String())
		}
	case <-ctx.Done():
		t.Fatal("CLI did not stop after exit notification")
	}

	return append([]byte(nil), stdout.Bytes()...), stderr.String()
}

func assertJSONRPCFrames(t *testing.T, output []byte, want int) {
	t.Helper()

	const separator = "\r\n\r\n"
	frameCount := 0
	for len(output) > 0 {
		headerEnd := bytes.Index(output, []byte(separator))
		if headerEnd < 0 {
			t.Fatalf("stdout contains non-JSON-RPC trailing data: %q", output)
		}
		header := string(output[:headerEnd])
		if !strings.HasPrefix(header, "Content-Length: ") {
			t.Fatalf("stdout frame has invalid header: %q", header)
		}
		length, err := strconv.Atoi(strings.TrimPrefix(header, "Content-Length: "))
		if err != nil || length < 0 {
			t.Fatalf("stdout frame has invalid content length %q: %v", header, err)
		}
		bodyStart := headerEnd + len(separator)
		if len(output) < bodyStart+length {
			t.Fatalf("stdout frame body is truncated: need %d bytes, have %d", length, len(output)-bodyStart)
		}
		body := output[bodyStart : bodyStart+length]
		var message map[string]any
		if err := json.Unmarshal(body, &message); err != nil {
			t.Fatalf("stdout frame is not JSON: %v (%q)", err, body)
		}
		if got := fmt.Sprint(message["jsonrpc"]); got != "2.0" {
			t.Fatalf("stdout frame JSON-RPC version = %q", got)
		}
		frameCount++
		output = output[bodyStart+length:]
	}
	if frameCount != want {
		t.Fatalf("stdout frame count = %d, want %d", frameCount, want)
	}
}
