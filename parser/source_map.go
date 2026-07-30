package parser

import "unicode/utf16"

// sourceMap converts byte offsets in the (UTF-8) lexer input into the
// position coordinates used by ScriptDom: UTF-16 code-unit offsets and
// 1-based line/column numbers (columns also count UTF-16 code units).
type sourceMap struct {
	// Indexed by byte offset (0..len(input)); continuation bytes of a
	// multi-byte rune carry the same values as the rune's first byte.
	u16  []int32
	line []int32
	col  []int32
}

func newSourceMap(input string) *sourceMap {
	n := len(input)
	m := &sourceMap{
		u16:  make([]int32, n+1),
		line: make([]int32, n+1),
		col:  make([]int32, n+1),
	}
	u16Off, line, col := int32(0), int32(1), int32(1)
	i := 0
	for i < n {
		r, size := decodeRuneAt(input, i)
		w := int32(1)
		if utf16.RuneLen(r) == 2 {
			w = 2
		}
		for j := 0; j < size; j++ {
			m.u16[i+j] = u16Off
			m.line[i+j] = line
			m.col[i+j] = col
		}
		u16Off += w
		if r == '\n' {
			line++
			col = 1
		} else {
			col += w
		}
		i += size
	}
	m.u16[n] = u16Off
	m.line[n] = line
	m.col[n] = col
	return m
}

// at returns the UTF-16 offset, line, and column for a byte offset.
func (m *sourceMap) at(byteOff int) (u16, line, col int) {
	if byteOff < 0 {
		byteOff = 0
	}
	if byteOff >= len(m.u16) {
		byteOff = len(m.u16) - 1
	}
	return int(m.u16[byteOff]), int(m.line[byteOff]), int(m.col[byteOff])
}

// u16Len returns the total input length in UTF-16 code units.
func (m *sourceMap) u16Len() int {
	return int(m.u16[len(m.u16)-1])
}
