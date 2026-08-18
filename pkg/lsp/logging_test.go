package lsp

import "testing"

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    LogLevel
		wantErr bool
	}{
		{name: "error", value: "error", want: LogLevelError},
		{name: "info", value: "info", want: LogLevelInfo},
		{name: "debug", value: "debug", want: LogLevelDebug},
		{name: "case insensitive", value: "DEBUG", want: LogLevelDebug},
		{name: "invalid", value: "trace", want: LogLevelInfo, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLogLevel(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseLogLevel(%q) error = %v, wantErr %t", tt.value, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseLogLevel(%q) = %d, want %d", tt.value, got, tt.want)
			}
		})
	}
}
