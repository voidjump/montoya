package montoya

import (
	"fmt"
	"io"
)

type TokenType int

const (
	Whitespace = iota //
	BracketOpen // [
	BracketClose // ]
	NewLine // \n
	Equals // =
	SemiColon // ;
	Hash // #
	Quote // "
	Other // ?
)

type Token struct {
	Content byte
	Kind TokenType
}

// convert tokensj
func convertToken(rawByte byte) TokenType {
	switch rawByte {
	case 0x0A: 
		return NewLine
	case 0x5B:
		return BracketOpen
	case 0x5D:
		return BracketClose
	case 0x3D:
		return Equals
	case 0x23:
		return Hash
	case 0x3B:
		return SemiColon
	case 0x22:
		return Quote
	// Whitespace cases
	case 0x20, // space
		0x09, // tab
		0x0D: // carriage return
		return Whitespace
	default:
		return Other
	}
}

// Tokenize converts an input bytestream to a slice of Tokens
func Tokenize(input io.Reader) ([]Token,error) {
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
			Content : buf[0],
			Kind : convertToken(buf[0]),
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}


