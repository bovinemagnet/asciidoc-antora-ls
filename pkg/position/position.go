// Package position converts Language Server Protocol character offsets.
package position

import "unicode/utf8"

// Encoding identifies the code units used for LSP character offsets.
type Encoding string

const (
	// UTF8 counts UTF-8 bytes.
	UTF8 Encoding = "utf-8"
	// UTF16 counts UTF-16 code units and is the LSP default.
	UTF16 Encoding = "utf-16"
)

// ByteOffset converts an LSP character offset to a safe byte offset in line.
// Out-of-range offsets clamp to the end of the line. A UTF-16 offset inside a
// surrogate pair clamps to the beginning of that rune.
func ByteOffset(line string, character uint32, encoding Encoding) int {
	if character == 0 || line == "" {
		return 0
	}

	if encoding == UTF8 {
		if uint64(character) >= uint64(len(line)) {
			return len(line)
		}
		return int(character)
	}

	var codeUnits uint32
	for byteOffset := 0; byteOffset < len(line); {
		r, size := utf8.DecodeRuneInString(line[byteOffset:])
		width := uint32(1)
		if r > 0xffff {
			width = 2
		}

		if character < codeUnits+width {
			return byteOffset
		}

		codeUnits += width
		byteOffset += size
		if character == codeUnits {
			return byteOffset
		}
	}

	return len(line)
}

// CharacterOffset converts a byte offset to the selected LSP character units.
// Byte offsets outside the line clamp to the nearest line boundary.
func CharacterOffset(line string, byteOffset int, encoding Encoding) uint32 {
	if byteOffset <= 0 || line == "" {
		return 0
	}
	if byteOffset > len(line) {
		byteOffset = len(line)
	}
	if encoding == UTF8 {
		return uint32(byteOffset)
	}

	var codeUnits uint32
	for offset := 0; offset < byteOffset; {
		r, size := utf8.DecodeRuneInString(line[offset:])
		if offset+size > byteOffset {
			break
		}
		if r > 0xffff {
			codeUnits += 2
		} else {
			codeUnits++
		}
		offset += size
	}
	return codeUnits
}
