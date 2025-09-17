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
		0x0A, 0x5B, 0x5D, 0x3D, 0x23, 0x3B, 0x22, 0x20, 0x09, 0x0D,
	}

	reader := bytes.NewReader(file)

	tokenStream, err := Tokenize(reader)

	assert.NoError(t, err)

	expected := []Token{
		{
			Content:0x0A,
			Kind:NewLine,
		},
		{
			Content:0x5B,
			Kind:BracketOpen,
		},
		{
			Content:0x5D,
			Kind:BracketClose,
		},
		{
			Content:0x3D,
			Kind:Equals,
		},
		{
			Content:0x23,
			Kind:Hash,
		},
		{
			Content:0x3B,
			Kind:SemiColon,
		},
		{
			Content:0x22,
			Kind:Quote,
		},
		{
			Content:0x20,
			Kind:Whitespace,
		},
		{
			Content:0x09,
			Kind:Whitespace,
		},
		{
			Content:0x0D,
			Kind:Whitespace,
		},
	}

	assert.ElementsMatch(t, expected, tokenStream)
}


// Check that all Unspecified tokens are tokenized as type 'Other'
func TestOtherTokens(t *testing.T) {
	specified := []byte {
		0x0A, 0x5B, 0x5D, 0x3D, 0x23, 0x3B, 0x22, 0x20, 0x09, 0x0D,
	}

	for rawByte := 0x00; rawByte < 0xff; rawByte++ {
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