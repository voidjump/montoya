package montoya

import (
	"bytes"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test all specified types are tokenized correctly
func TestSpecifiedTokens(t *testing.T) {

	file := []byte {
		B_NEWLINE, B_BRACKET, B_BRACKETCLOSE, B_EQUALS, B_HASH, B_SEMICOLON, B_QUOTE, B_SPACE, B_TAB, B_CR,
	}

	reader := bytes.NewReader(file)

	tokenStream, err := Tokenize(reader)

	assert.NoError(t, err)

	expected := []Token{
		{
			Content:B_NEWLINE,
			Kind:NewLine,
		},
		{
			Content:B_BRACKET,
			Kind:SectionStart,
		},
		{
			Content:B_BRACKETCLOSE,
			Kind:SectionEnd,
		},
		{
			Content:B_EQUALS,
			Kind:Equals,
		},
		{
			Content:B_HASH,
			Kind:CommentStart,
		},
		{
			Content:B_SEMICOLON,
			Kind:CommentStart,
		},
		{
			Content:B_QUOTE,
			Kind:Quote,
		},
		{
			Content:B_SPACE,
			Kind:Whitespace,
		},
		{
			Content:B_TAB,
			Kind:Whitespace,
		},
		{
			Content:B_CR,
			Kind:Whitespace,
		},
	}

	assert.ElementsMatch(t, expected, tokenStream)
}


// Check that all Unspecified tokens are tokenized as type 'Other'
func TestOtherTokens(t *testing.T) {
	specified := []byte {
		B_NEWLINE, B_BRACKET, B_BRACKETCLOSE, B_EQUALS, B_HASH, B_SEMICOLON, B_QUOTE, B_SPACE, B_TAB, B_CR,
	}

	for rawByte := 0x00; rawByte < 256; rawByte++ {
		castByte := byte(rawByte)
		// do something with x
		if slices.Contains(specified, castByte) {
			continue
		}

		file := []byte{castByte}
		reader := bytes.NewReader(file)

		tokenStream, err := Tokenize(reader)

		assert.NoError(t, err)
		assert.Equal(t, tokenStream[0].Content, castByte)
		assert.Equal(t, tokenStream[0].Kind, TokenType(Other))
	}
}