package position

import "testing"

func TestByteOffset_UTF16(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		character uint32
		want      int
	}{
		{name: "ascii", line: "abc", character: 2, want: 2},
		{name: "two byte latin", line: "éx", character: 1, want: 2},
		{name: "three byte CJK", line: "界x", character: 1, want: 3},
		{name: "surrogate pair", line: "😀x", character: 2, want: 4},
		{name: "inside surrogate pair", line: "😀x", character: 1, want: 0},
		{name: "out of range", line: "éx", character: 100, want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ByteOffset(tt.line, tt.character, UTF16); got != tt.want {
				t.Errorf("ByteOffset(%q, %d, UTF16) = %d, want %d", tt.line, tt.character, got, tt.want)
			}
		})
	}
}

func TestByteOffset_UTF8(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		character uint32
		want      int
	}{
		{name: "ascii", line: "abc", character: 2, want: 2},
		{name: "two byte latin", line: "éx", character: 2, want: 2},
		{name: "three byte CJK", line: "界x", character: 3, want: 3},
		{name: "astral rune", line: "😀x", character: 4, want: 4},
		{name: "out of range", line: "😀x", character: 100, want: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ByteOffset(tt.line, tt.character, UTF8); got != tt.want {
				t.Errorf("ByteOffset(%q, %d, UTF8) = %d, want %d", tt.line, tt.character, got, tt.want)
			}
		})
	}
}

func TestCharacterOffset(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		encoding Encoding
		want     uint32
	}{
		{name: "ASCII UTF-16", line: "abc", encoding: UTF16, want: 3},
		{name: "Latin UTF-16", line: "é", encoding: UTF16, want: 1},
		{name: "CJK UTF-16", line: "界", encoding: UTF16, want: 1},
		{name: "astral UTF-16", line: "😀", encoding: UTF16, want: 2},
		{name: "UTF-8 bytes", line: "é界😀", encoding: UTF8, want: 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CharacterOffset(tt.line, len(tt.line), tt.encoding); got != tt.want {
				t.Errorf("CharacterOffset(%q, len, %q) = %d, want %d", tt.line, tt.encoding, got, tt.want)
			}
		})
	}
}
