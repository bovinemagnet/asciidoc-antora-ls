package asciidoc

import (
	"fmt"
	"regexp"
	"strings"

	"go.lsp.dev/protocol"
)

// Parser handles AsciiDoc document parsing
type Parser struct {
	headingRegex   *regexp.Regexp
	attributeRegex *regexp.Regexp
	includeRegex   *regexp.Regexp
	xrefRegex      *regexp.Regexp
}

// NewParser creates a new AsciiDoc parser
func NewParser() *Parser {
	return &Parser{
		headingRegex:   regexp.MustCompile(`^(=+)\s+(.+)$`),
		attributeRegex: regexp.MustCompile(`^:([a-zA-Z0-9_-]+):\s*(.*)$`),
		includeRegex:   regexp.MustCompile(`^include::([^\[]+)\[([^\]]*)\]$`),
		xrefRegex:      regexp.MustCompile(`xref:([^\[]+)\[([^\]]*)\]`),
	}
}

// GetHoverInfo returns hover information for the given position
func (p *Parser) GetHoverInfo(content string, pos protocol.Position) string {
	lines := strings.Split(content, "\n")
	if int(pos.Line) >= len(lines) {
		return ""
	}

	line := lines[pos.Line]

	// Check for heading
	if matches := p.headingRegex.FindStringSubmatch(line); matches != nil {
		level := len(matches[1])
		return formatHeadingInfo(level, matches[2])
	}

	// Check for attribute
	if matches := p.attributeRegex.FindStringSubmatch(line); matches != nil {
		return formatAttributeInfo(matches[1], matches[2])
	}

	// Check for include
	if matches := p.includeRegex.FindStringSubmatch(line); matches != nil {
		return formatIncludeInfo(matches[1])
	}

	return ""
}

// GetCompletions returns completion items for the given position
func (p *Parser) GetCompletions(content string, pos protocol.Position) []protocol.CompletionItem {
	lines := strings.Split(content, "\n")
	if int(pos.Line) >= len(lines) {
		return []protocol.CompletionItem{}
	}

	line := lines[pos.Line]
	prefix := line[:pos.Character]

	var items []protocol.CompletionItem

	// Attribute completions
	if strings.HasPrefix(strings.TrimSpace(prefix), ":") {
		items = append(items, getAttributeCompletions()...)
	}

	// Block delimiters
	if strings.TrimSpace(prefix) == "" || strings.HasPrefix(prefix, "-") || strings.HasPrefix(prefix, "=") {
		items = append(items, getBlockCompletions()...)
	}

	// Inline macros
	if strings.Contains(prefix, "xref:") || strings.Contains(prefix, "include::") {
		items = append(items, getMacroCompletions()...)
	}

	return items
}

// GetDocumentSymbols returns document symbols (headings, sections)
func (p *Parser) GetDocumentSymbols(content string) []protocol.DocumentSymbol {
	lines := strings.Split(content, "\n")
	symbols := []protocol.DocumentSymbol{}

	for i, line := range lines {
		if matches := p.headingRegex.FindStringSubmatch(line); matches != nil {
			level := len(matches[1])
			text := matches[2]

			symbol := protocol.DocumentSymbol{
				Name: text,
				Kind: protocol.SymbolKindNamespace,
				Range: protocol.Range{
					Start: protocol.Position{Line: uint32(i), Character: 0},
					End:   protocol.Position{Line: uint32(i), Character: uint32(len(line))},
				},
				SelectionRange: protocol.Range{
					Start: protocol.Position{Line: uint32(i), Character: uint32(level + 1)},
					End:   protocol.Position{Line: uint32(i), Character: uint32(len(line))},
				},
			}

			symbols = append(symbols, symbol)
		}
	}

	return symbols
}

func formatHeadingInfo(level int, text string) string {
	return fmt.Sprintf("**Heading Level %d**\n\n%s", level, text)
}

func formatAttributeInfo(name, value string) string {
	if value != "" {
		return fmt.Sprintf("**Document Attribute**\n\nName: `%s`\n\nValue: `%s`", name, value)
	}
	return fmt.Sprintf("**Document Attribute**\n\nName: `%s`", name)
}

func formatIncludeInfo(path string) string {
	return fmt.Sprintf("**Include Directive**\n\nPath: `%s`", path)
}

func getAttributeCompletions() []protocol.CompletionItem {
	attrs := []string{
		"toc", "toc-title", "doctype", "author", "email",
		"revdate", "revnumber", "revremark", "description",
		"keywords", "lang", "encoding", "icons", "iconsdir",
		"imagesdir", "stylesheet", "stylesdir", "linkcss",
		"sectanchors", "sectlinks", "sectnums", "partnums",
	}

	items := make([]protocol.CompletionItem, len(attrs))
	for i, attr := range attrs {
		items[i] = protocol.CompletionItem{
			Label:      attr,
			Kind:       protocol.CompletionItemKindProperty,
			Detail:     "AsciiDoc attribute",
			InsertText: attr + ": ",
		}
	}
	return items
}

func getBlockCompletions() []protocol.CompletionItem {
	return []protocol.CompletionItem{
		{
			Label:      "Example Block",
			Kind:       protocol.CompletionItemKindSnippet,
			Detail:     "====",
			InsertText: "====\n$0\n====",
		},
		{
			Label:      "Sidebar Block",
			Kind:       protocol.CompletionItemKindSnippet,
			Detail:     "****",
			InsertText: "****\n$0\n****",
		},
		{
			Label:      "Listing Block",
			Kind:       protocol.CompletionItemKindSnippet,
			Detail:     "----",
			InsertText: "----\n$0\n----",
		},
		{
			Label:      "Literal Block",
			Kind:       protocol.CompletionItemKindSnippet,
			Detail:     "....",
			InsertText: "....\n$0\n....",
		},
	}
}

func getMacroCompletions() []protocol.CompletionItem {
	return []protocol.CompletionItem{
		{
			Label:      "xref",
			Kind:       protocol.CompletionItemKindFunction,
			Detail:     "Cross reference",
			InsertText: "xref:$1[$2]",
		},
		{
			Label:      "include",
			Kind:       protocol.CompletionItemKindFunction,
			Detail:     "Include file",
			InsertText: "include::$1[]",
		},
	}
}
