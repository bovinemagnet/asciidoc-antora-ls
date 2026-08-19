package lsp

import "testing"

func TestFormatVersion(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		commit    string
		buildDate string
		want      string
	}{
		{
			name:    "development defaults",
			version: "0.1.0",
			commit:  "unknown",
			want:    "0.1.0",
		},
		{
			name:      "release metadata",
			version:   "1.2.3",
			commit:    "abc1234",
			buildDate: "2026-08-18T02:21:53Z",
			want:      "1.2.3 (commit abc1234, built 2026-08-18T02:21:53Z)",
		},
		{
			name:      "date without commit",
			version:   "1.2.3",
			commit:    "unknown",
			buildDate: "2026-08-18T02:21:53Z",
			want:      "1.2.3 (built 2026-08-18T02:21:53Z)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatVersion(tt.version, tt.commit, tt.buildDate); got != tt.want {
				t.Errorf("formatVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}
