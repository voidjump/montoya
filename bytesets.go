package montoya

import "slices"

const B_NULL byte = 0x00
const B_NEWLINE byte = 0x0A
const B_SPACE byte = 0x20
const B_TAB byte = 0x09
const B_CR byte = 0x0D
const B_BRACKET byte = 0x5B
const B_BRACKETCLOSE byte = 0x5D
const B_HASH byte = 0x23
const B_SEMICOLON byte = 0x3B
const B_EQUALS byte = 0x3D
const B_QUOTE byte = 0x22
const B_BACKSLASH byte = 0x5C
const B_US byte = 0x1F

var commentStartBytes = []byte{
	B_HASH,
	B_SEMICOLON,
}

var validWhitespaceByteSet = []byte{
	B_SPACE,
	B_TAB,
	B_CR,
}

// var invalidWhitespaceByteset = invertByteSet(validWhitespaceByteSet)

// Sections may not contain nulls, newlines, brackets, or comment start symbols
var invalidSectionByteSet = []byte{
	B_NULL,
	B_NEWLINE,
	B_BRACKET,
	B_BRACKETCLOSE,
	B_HASH,
	B_SEMICOLON,
}
var validSectionByteSet = invertByteSet(invalidSectionByteSet)

// Comment may not contain nulls or newlines
var invalidCommentByteSet = []byte{
	B_NULL,
	B_NEWLINE,
}
var validCommentByteSet = invertByteSet(invalidCommentByteSet)

// Value not inside quotes may not contain quotes or comment symbols
var invalidValueByteSetUnquoted = []byte{
	B_NULL,
	B_NEWLINE,
	B_HASH,
	B_SEMICOLON,
	B_QUOTE,
}
var validValueByteSetUnquoted = invertByteSet(invalidValueByteSetUnquoted)

// Values inside quotes may contain comment starts and quotes, if they are escaped
var invalidValueByteSetQuoted = []byte{
	B_NULL,
	B_NEWLINE,
}
var validValueByteSetQuoted = invertByteSet(invalidValueByteSetQuoted)

// invertBytes returns a set of all possible byte values excluding those in `invalid`
func invertByteSet(invalid []byte) (inverted []byte) {
	for i := range 256 {
		b := byte(i)
		if slices.Contains(invalid, b) {
			continue
		}
		inverted = append(inverted, b)
	}
	return
}
