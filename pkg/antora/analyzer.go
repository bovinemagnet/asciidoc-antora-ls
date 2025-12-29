package antora

import (
	"path/filepath"
	"regexp"
	"strings"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

// Analyzer handles Antora-specific features
type Analyzer struct {
	pageXrefRegex      *regexp.Regexp
	componentXrefRegex *regexp.Regexp
	attributeRegex     *regexp.Regexp
}

// NewAnalyzer creates a new Antora analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		pageXrefRegex:      regexp.MustCompile(`xref:([^:]+\.adoc|[^:]+::[^:]+\.adoc)\[([^\]]*)\]`),
		componentXrefRegex: regexp.MustCompile(`xref:(\w+):(\w+):([^\[]+)\[([^\]]*)\]`),
		attributeRegex:     regexp.MustCompile(`\{([a-zA-Z0-9_-]+)\}`),
	}
}

// GetCompletions returns Antora-specific completion items
func (a *Analyzer) GetCompletions(content string, pos protocol.Position) []protocol.CompletionItem {
	lines := strings.Split(content, "\n")
	if int(pos.Line) >= len(lines) {
		return []protocol.CompletionItem{}
	}

	line := lines[pos.Line]
	prefix := line[:pos.Character]

	var items []protocol.CompletionItem

	// Antora page attributes
	if strings.Contains(prefix, "{") {
		items = append(items, getAntoraAttributeCompletions()...)
	}

	// Antora xref formats
	if strings.Contains(prefix, "xref:") {
		items = append(items, getAntoraXrefCompletions()...)
	}

	return items
}

// GetDefinition returns the definition location for Antora cross-references
func (a *Analyzer) GetDefinition(content string, pos protocol.Position, docURI uri.URI) []protocol.Location {
	lines := strings.Split(content, "\n")
	if int(pos.Line) >= len(lines) {
		return nil
	}

	line := lines[pos.Line]

	// Try to find an xref at the cursor position
	if matches := a.pageXrefRegex.FindStringSubmatch(line); matches != nil {
		targetPath := matches[1]

		// Resolve the target path relative to the current document
		docPath := docURI.Filename()
		docDir := filepath.Dir(docPath)

		// Handle Antora page references
		var targetFile string
		if strings.Contains(targetPath, "::") {
			// Component reference format: component::page.adoc
			parts := strings.Split(targetPath, "::")
			if len(parts) == 2 {
				// This would need actual component resolution in a real implementation
				targetFile = filepath.Join(docDir, "..", parts[0], "pages", parts[1])
			}
		} else {
			// Simple page reference
			targetFile = filepath.Join(docDir, targetPath)
		}

		if targetFile != "" {
			location := protocol.Location{
				URI: uri.File(targetFile),
				Range: protocol.Range{
					Start: protocol.Position{Line: 0, Character: 0},
					End:   protocol.Position{Line: 0, Character: 0},
				},
			}
			return []protocol.Location{location}
		}
	}

	return nil
}

// IsAntoraDocument checks if a document is part of an Antora component
func (a *Analyzer) IsAntoraDocument(docPath string) bool {
	// Check if the document is in a typical Antora structure
	// e.g., modules/*/pages/*.adoc
	return strings.Contains(docPath, "/modules/") &&
		strings.Contains(docPath, "/pages/") &&
		strings.HasSuffix(docPath, ".adoc")
}

func getAntoraAttributeCompletions() []protocol.CompletionItem {
	attrs := []string{
		"component-name", "component-title", "component-version",
		"module", "page-component-name", "page-component-title",
		"page-component-version", "page-module", "page-relative-src-path",
		"page-origin-type", "page-origin-url", "page-origin-refname",
		"page-origin-start-path", "site-title", "site-url",
	}

	items := make([]protocol.CompletionItem, len(attrs))
	for i, attr := range attrs {
		items[i] = protocol.CompletionItem{
			Label:      attr,
			Kind:       protocol.CompletionItemKindVariable,
			Detail:     "Antora attribute",
			InsertText: attr,
		}
	}
	return items
}

func getAntoraXrefCompletions() []protocol.CompletionItem {
	return []protocol.CompletionItem{
		{
			Label:      "Page reference (same module)",
			Kind:       protocol.CompletionItemKindReference,
			Detail:     "xref:page.adoc[]",
			InsertText: "xref:$1.adoc[$2]",
		},
		{
			Label:      "Page reference (other module)",
			Kind:       protocol.CompletionItemKindReference,
			Detail:     "xref:module:page.adoc[]",
			InsertText: "xref:$1:$2.adoc[$3]",
		},
		{
			Label:      "Page reference (other component)",
			Kind:       protocol.CompletionItemKindReference,
			Detail:     "xref:version@component:module:page.adoc[]",
			InsertText: "xref:$1@$2:$3:$4.adoc[$5]",
		},
		{
			Label:      "Section reference",
			Kind:       protocol.CompletionItemKindReference,
			Detail:     "xref:page.adoc#section[]",
			InsertText: "xref:$1.adoc#$2[$3]",
		},
	}
}
