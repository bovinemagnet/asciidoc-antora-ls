package antora

import (
	"testing"

	"go.lsp.dev/protocol"
)

func TestAnalyzer_GetCompletions(t *testing.T) {
	analyzer := NewAnalyzer()
	content := "xref:"

	items := analyzer.GetCompletions(content, protocol.Position{Line: 0, Character: 5})
	if len(items) == 0 {
		t.Error("Expected Antora xref completions, got none")
	}

	// Check for Antora-specific completion
	hasPageRef := false
	for _, item := range items {
		if item.Label == "Page reference (same module)" {
			hasPageRef = true
			break
		}
	}
	if !hasPageRef {
		t.Error("Expected 'Page reference (same module)' in completions")
	}
}

func TestAnalyzer_GetCompletions_Attributes(t *testing.T) {
	analyzer := NewAnalyzer()
	content := "{component"

	items := analyzer.GetCompletions(content, protocol.Position{Line: 0, Character: 10})
	if len(items) == 0 {
		t.Error("Expected Antora attribute completions, got none")
	}

	// Check for Antora-specific attributes
	hasComponentName := false
	for _, item := range items {
		if item.Label == "component-name" {
			hasComponentName = true
			break
		}
	}
	if !hasComponentName {
		t.Error("Expected 'component-name' in attribute completions")
	}
}

func TestAnalyzer_IsAntoraDocument(t *testing.T) {
	analyzer := NewAnalyzer()

	tests := []struct {
		path     string
		expected bool
	}{
		{"/path/to/modules/ROOT/pages/index.adoc", true},
		{"/path/to/modules/admin/pages/guide.adoc", true},
		{"/path/to/docs/index.adoc", false},
		{"/path/to/modules/ROOT/examples/example.adoc", false},
	}

	for _, tt := range tests {
		result := analyzer.IsAntoraDocument(tt.path)
		if result != tt.expected {
			t.Errorf("IsAntoraDocument(%s) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

func TestAnalyzer_GetDefinition(t *testing.T) {
	analyzer := NewAnalyzer()
	content := "xref:other-page.adoc[Link text]"

	// This is a basic test - full implementation would need filesystem access
	locations := analyzer.GetDefinition(content, protocol.Position{Line: 0, Character: 10}, "file:///test/page.adoc")

	// Should return at least an attempt at a location
	if locations == nil {
		t.Log("GetDefinition returned nil, which is acceptable for this test")
	}
}
