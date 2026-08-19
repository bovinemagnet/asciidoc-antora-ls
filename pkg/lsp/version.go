package lsp

import (
	"fmt"
	"strings"
)

// Version is the language-server version. Release builds override it with
// -ldflags so the CLI and initialize response always report the same value.
var Version = "0.1.0"

// Commit is the source revision used for the build.
var Commit = "unknown"

// BuildDate is the source commit date used for the build.
var BuildDate = "unknown"

// VersionString returns the shared CLI and LSP build description.
func VersionString() string {
	return formatVersion(Version, Commit, BuildDate)
}

func formatVersion(version, commit, buildDate string) string {
	details := make([]string, 0, 2)
	if commit != "" && commit != "unknown" {
		details = append(details, "commit "+commit)
	}
	if buildDate != "" && buildDate != "unknown" {
		details = append(details, "built "+buildDate)
	}
	if len(details) == 0 {
		return version
	}
	return fmt.Sprintf("%s (%s)", version, strings.Join(details, ", "))
}
