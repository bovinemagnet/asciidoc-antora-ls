package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/bovinemagnet/asciidoc-antora-ls/pkg/antora"
	"github.com/bovinemagnet/asciidoc-antora-ls/pkg/asciidoc"
	"github.com/bovinemagnet/asciidoc-antora-ls/pkg/position"
	"github.com/sourcegraph/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

// Server implements the LSP server for AsciiDoc and Antora
type Server struct {
	mu               sync.RWMutex
	documents        map[string]*Document
	parser           *asciidoc.Parser
	antora           *antora.Analyzer
	positionEncoding position.Encoding
}

// Document represents an open document in the editor
type Document struct {
	URI     string
	Content string
	Version int32
}

// ServerCapabilities adds LSP 3.17 position encoding support to the protocol
// dependency's older capability model.
type ServerCapabilities struct {
	protocol.ServerCapabilities
	PositionEncoding position.Encoding `json:"positionEncoding"`
}

// InitializeResult is the server's response to an initialize request.
type InitializeResult struct {
	Capabilities ServerCapabilities   `json:"capabilities"`
	ServerInfo   *protocol.ServerInfo `json:"serverInfo,omitempty"`
}

// NewServer creates a new LSP server instance
func NewServer() *Server {
	parser := asciidoc.NewParser()
	antoraAnalyzer := antora.NewAnalyzer()
	return &Server{
		documents:        make(map[string]*Document),
		parser:           parser,
		antora:           antoraAnalyzer,
		positionEncoding: position.UTF16,
	}
}

// Initialize handles the initialize request
func (s *Server) Initialize(ctx context.Context, params *protocol.InitializeParams, offeredEncodings ...position.Encoding) (*InitializeResult, error) {
	log.Printf("Initialize request from client: %s", params.ClientInfo.Name)

	s.positionEncoding = selectPositionEncoding(offeredEncodings)
	s.parser.SetPositionEncoding(s.positionEncoding)
	s.antora.SetPositionEncoding(s.positionEncoding)

	return &InitializeResult{
		Capabilities: ServerCapabilities{
			PositionEncoding: s.positionEncoding,
			ServerCapabilities: protocol.ServerCapabilities{
				TextDocumentSync: protocol.TextDocumentSyncOptions{
					OpenClose: true,
					Change:    protocol.TextDocumentSyncKindFull,
					Save: &protocol.SaveOptions{
						IncludeText: false,
					},
				},
				HoverProvider: true,
				CompletionProvider: &protocol.CompletionOptions{
					TriggerCharacters: []string{":", "{", "["},
				},
				DefinitionProvider:     true,
				DocumentSymbolProvider: true,
			},
		},
		ServerInfo: &protocol.ServerInfo{
			Name:    "asciidoc-antora-ls",
			Version: "0.1.0",
		},
	}, nil
}

// Initialized handles the initialized notification
func (s *Server) Initialized(ctx context.Context, params *protocol.InitializedParams) error {
	log.Println("Server initialized")
	return nil
}

// Shutdown handles the shutdown request
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Server shutdown requested")
	return nil
}

// Exit handles the exit notification
func (s *Server) Exit(ctx context.Context) error {
	log.Println("Server exit")
	return nil
}

// DidOpen handles document open notifications
func (s *Server) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	docURI := string(params.TextDocument.URI)
	log.Printf("Document opened: %s", docURI)

	s.documents[docURI] = &Document{
		URI:     docURI,
		Content: params.TextDocument.Text,
		Version: params.TextDocument.Version,
	}

	return nil
}

// DidChange handles document change notifications
func (s *Server) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	docURI := string(params.TextDocument.URI)
	doc, exists := s.documents[docURI]
	if !exists {
		return fmt.Errorf("document not found: %s", docURI)
	}

	for _, change := range params.ContentChanges {
		doc.Content = change.Text
	}
	doc.Version = params.TextDocument.Version

	log.Printf("Document changed: %s (version %d)", docURI, doc.Version)
	return nil
}

// DidSave handles document save notifications
func (s *Server) DidSave(ctx context.Context, params *protocol.DidSaveTextDocumentParams) error {
	docURI := string(params.TextDocument.URI)
	log.Printf("Document saved: %s", docURI)
	return nil
}

// DidClose handles document close notifications
func (s *Server) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	docURI := string(params.TextDocument.URI)
	delete(s.documents, docURI)
	log.Printf("Document closed: %s", docURI)
	return nil
}

// Hover provides hover information
func (s *Server) Hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	docURI := string(params.TextDocument.URI)
	doc, exists := s.documents[docURI]
	if !exists {
		return nil, nil
	}

	info := s.parser.GetHoverInfo(doc.Content, params.Position)
	if info == "" {
		return nil, nil
	}

	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.Markdown,
			Value: info,
		},
	}, nil
}

// Completion provides completion suggestions
func (s *Server) Completion(ctx context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	docURI := string(params.TextDocument.URI)
	doc, exists := s.documents[docURI]
	if !exists {
		return &protocol.CompletionList{Items: []protocol.CompletionItem{}}, nil
	}

	items := s.parser.GetCompletions(doc.Content, params.Position)

	// Add Antora-specific completions
	antoraItems := s.antora.GetCompletions(doc.Content, params.Position)
	items = append(items, antoraItems...)

	return &protocol.CompletionList{
		IsIncomplete: false,
		Items:        items,
	}, nil
}

// Definition provides go-to-definition functionality
func (s *Server) Definition(ctx context.Context, params *protocol.DefinitionParams) ([]protocol.Location, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	docURI := string(params.TextDocument.URI)
	doc, exists := s.documents[docURI]
	if !exists {
		return nil, nil
	}

	locations := s.antora.GetDefinition(doc.Content, params.Position, uri.URI(docURI))
	return locations, nil
}

// DocumentSymbol provides document symbols
func (s *Server) DocumentSymbol(ctx context.Context, params *protocol.DocumentSymbolParams) ([]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	docURI := string(params.TextDocument.URI)
	doc, exists := s.documents[docURI]
	if !exists {
		return nil, nil
	}

	symbols := s.parser.GetDocumentSymbols(doc.Content)

	// Convert to interface{} slice
	result := make([]interface{}, len(symbols))
	for i, sym := range symbols {
		result[i] = sym
	}

	return result, nil
}

// Run starts the LSP server using the provided reader and writer
func (s *Server) Run(ctx context.Context, reader interface{ Read([]byte) (int, error) }, writer interface{ Write([]byte) (int, error) }) error {
	return s.runJSONRPC(ctx, reader, writer)
}

// runJSONRPC runs the JSON-RPC server
func (s *Server) runJSONRPC(ctx context.Context, reader interface{ Read([]byte) (int, error) }, writer interface{ Write([]byte) (int, error) }) error {
	stream := jsonrpc2.NewBufferedStream(&rwc{reader, writer}, jsonrpc2.VSCodeObjectCodec{})
	conn := jsonrpc2.NewConn(ctx, stream, s)
	<-conn.DisconnectNotify()
	return nil
}

// Handle implements jsonrpc2.Handler
func (s *Server) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	reply := func(result interface{}) {
		if err := conn.Reply(ctx, req.ID, result); err != nil {
			log.Printf("Failed to reply to %s: %v", req.Method, err)
		}
	}
	replyWithError := func(responseErr *jsonrpc2.Error) {
		if err := conn.ReplyWithError(ctx, req.ID, responseErr); err != nil {
			log.Printf("Failed to send %s error response: %v", req.Method, err)
		}
	}

	switch req.Method {
	case "initialize":
		var params protocol.InitializeParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			replyWithError(&jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams, Message: err.Error()})
			return
		}
		positionEncodings, err := clientPositionEncodings(*req.Params)
		if err != nil {
			replyWithError(&jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams, Message: err.Error()})
			return
		}
		result, err := s.Initialize(ctx, &params, positionEncodings...)
		if err != nil {
			replyWithError(&jsonrpc2.Error{Code: jsonrpc2.CodeInternalError, Message: err.Error()})
			return
		}
		reply(result)

	case "initialized":
		var params protocol.InitializedParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			replyWithError(&jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams, Message: err.Error()})
			return
		}
		if err := s.Initialized(ctx, &params); err != nil {
			log.Printf("Failed to handle initialized notification: %v", err)
		}

	case "shutdown":
		if err := s.Shutdown(ctx); err != nil {
			replyWithError(&jsonrpc2.Error{Code: jsonrpc2.CodeInternalError, Message: err.Error()})
			return
		}
		reply(nil)

	case "exit":
		if err := s.Exit(ctx); err != nil {
			log.Printf("Failed to handle exit notification: %v", err)
		}

	case "textDocument/didOpen":
		var params protocol.DidOpenTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return
		}
		if err := s.DidOpen(ctx, &params); err != nil {
			log.Printf("Failed to handle didOpen notification: %v", err)
		}

	case "textDocument/didChange":
		var params protocol.DidChangeTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return
		}
		if err := s.DidChange(ctx, &params); err != nil {
			log.Printf("Failed to handle didChange notification: %v", err)
		}

	case "textDocument/didSave":
		var params protocol.DidSaveTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return
		}
		if err := s.DidSave(ctx, &params); err != nil {
			log.Printf("Failed to handle didSave notification: %v", err)
		}

	case "textDocument/didClose":
		var params protocol.DidCloseTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return
		}
		if err := s.DidClose(ctx, &params); err != nil {
			log.Printf("Failed to handle didClose notification: %v", err)
		}

	case "textDocument/hover":
		var params protocol.HoverParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			replyWithError(&jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams, Message: err.Error()})
			return
		}
		result, err := s.Hover(ctx, &params)
		if err != nil {
			replyWithError(&jsonrpc2.Error{Code: jsonrpc2.CodeInternalError, Message: err.Error()})
			return
		}
		reply(result)

	case "textDocument/completion":
		var params protocol.CompletionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			replyWithError(&jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams, Message: err.Error()})
			return
		}
		result, err := s.Completion(ctx, &params)
		if err != nil {
			replyWithError(&jsonrpc2.Error{Code: jsonrpc2.CodeInternalError, Message: err.Error()})
			return
		}
		reply(result)

	case "textDocument/definition":
		var params protocol.DefinitionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			replyWithError(&jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams, Message: err.Error()})
			return
		}
		result, err := s.Definition(ctx, &params)
		if err != nil {
			replyWithError(&jsonrpc2.Error{Code: jsonrpc2.CodeInternalError, Message: err.Error()})
			return
		}
		reply(result)

	case "textDocument/documentSymbol":
		var params protocol.DocumentSymbolParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			replyWithError(&jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams, Message: err.Error()})
			return
		}
		result, err := s.DocumentSymbol(ctx, &params)
		if err != nil {
			replyWithError(&jsonrpc2.Error{Code: jsonrpc2.CodeInternalError, Message: err.Error()})
			return
		}
		reply(result)

	default:
		if req.Notif {
			// Ignore unknown notifications
			return
		}
		replyWithError(&jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: "method not found"})
	}
}

func selectPositionEncoding(offered []position.Encoding) position.Encoding {
	for _, encoding := range offered {
		if encoding == position.UTF8 || encoding == position.UTF16 {
			return encoding
		}
	}
	return position.UTF16
}

func clientPositionEncodings(params json.RawMessage) ([]position.Encoding, error) {
	var capabilities struct {
		Capabilities struct {
			General struct {
				PositionEncodings []position.Encoding `json:"positionEncodings"`
			} `json:"general"`
		} `json:"capabilities"`
	}
	if err := json.Unmarshal(params, &capabilities); err != nil {
		return nil, err
	}
	return capabilities.Capabilities.General.PositionEncodings, nil
}

// rwc wraps a reader and writer
type rwc struct {
	reader interface{ Read([]byte) (int, error) }
	writer interface{ Write([]byte) (int, error) }
}

func (c *rwc) Read(p []byte) (int, error) {
	return c.reader.Read(p)
}

func (c *rwc) Write(p []byte) (int, error) {
	return c.writer.Write(p)
}

func (c *rwc) Close() error {
	return nil
}
