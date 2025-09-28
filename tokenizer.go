package montoya

import (
	"fmt"
	"io"
)

type TokenType int

const (
	Whitespace   = iota // tab, space or carriage return
	CommentStart        // # or ;
	SectionStart        // [
	SectionEnd          // ]
	NewLine             // \n
	Equals              // =
	Quote               // "
	Other               // ?
)

type Token struct {
	Content byte
	Kind    TokenType
}

// convert tokensj
func convertToken(rawByte byte) TokenType {
	switch rawByte {
	case B_NEWLINE:
		return NewLine
	case B_BRACKET:
		return SectionStart
	case B_BRACKETCLOSE:
		return SectionEnd
	case B_EQUALS:
		return Equals
	case B_SEMICOLON,
		B_HASH:
		return CommentStart
	case B_QUOTE:
		return Quote
	// Whitespace cases
	case B_SPACE, // space
		B_TAB, // tab
		B_CR:  // carriage return
		return Whitespace
	default:
		return Other
	}
}

// Tokenize converts an input bytestream to a slice of Tokens
func Tokenize(input io.Reader) ([]Token, error) {
	var tokens []Token

	buf := make([]byte, 1) // slice of length 1

	for {
		_, err := input.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read input: %w", err)
		}
		token := Token{
			Content: buf[0],
			Kind:    convertToken(buf[0]),
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}
