package montoya

import (
	"math/rand"
)

// fuzzyChoice returns a randomly chosen byte from the input
func fuzzyChoice(choices []byte) byte {
	return choices[rand.Intn(len(choices))]
}

// fuzzWhiteSpace creates a bytestring of whitespace for testing
func fuzzWhiteSpace(numChars int) (whitespace []byte) {
	for range numChars {
		whitespace = append(whitespace, fuzzyChoice(validWhitespaceByteSet))
	}
	return whitespace
}

// fuzzComment creates a fuzzy comment
func fuzzComment() (comment []byte) {

	comment = append(comment, fuzzyChoice(commentStartBytes))
	n := rand.Intn(100)
	for i := 0; i < n; i++ {
		comment = append(comment, fuzzyChoice(validCommentByteSet))
	}
	return
}

// fuzzSection creates a fuzzy section name
//
// If `includeBrackets` is set, adds brackets
func fuzzSection(includeBrackets bool) (section []byte) {
	if includeBrackets {
		section = append(section, B_BRACKET)
	}

	n := rand.Intn(100)
	for range n {
		section = append(section, fuzzyChoice(validSectionByteSet))
	}

	if includeBrackets {
		section = append(section, B_BRACKETCLOSE)
	}
	return
}

// fuzzValue creates a fuzzy value
func fuzzValue(useQuotes bool) (value []byte) {
	if useQuotes {
		value = append(value, B_QUOTE)
	}
	n := rand.Intn(100)
	for range n {
		if useQuotes {
			newByte := fuzzyChoice(validValueByteSetQuoted)
			if newByte == B_QUOTE {
				value = append(value, B_BACKSLASH)
				value = append(value, B_QUOTE)
			} else {
				value = append(value, newByte)
			}
		} else {
			value = append(value, fuzzyChoice(validValueByteSetUnquoted))
		}
	}
	if useQuotes {
		value = append(value, B_QUOTE)
	}

	return
}

func fuzzKey() (key []byte) {
	for range 100 {
		key = append(key, fuzzyChoice(validKeyByteSet))
	}
	return key
}
