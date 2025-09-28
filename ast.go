package montoya

import (
	"bytes"
	"io"
	"slices"
)

// An IniLine is any line type in an Inifile
type IniLine interface {
	io.Reader
	Reset()
	Next() IniLine
	Previous() IniLine
	SetPrev(node IniLine)
	SetNext(node IniLine)
}

// An IniNode is part of an IniLine
type IniNode interface {
	Content() []byte
}

type LineBase struct {
	ReadBuf []byte
	reader io.Reader

	prev IniLine
	next IniLine
}

// Reset reader state
func (l *LineBase) Reset() {
	l.ReadBuf = []byte{}
	l.reader = nil
}

func (l *LineBase) HasReader() bool {
	return l.reader != nil
}



func (l *LineBase) Read(p []byte) (n int , err error) {
	if l.reader == nil {
		l.reader = bytes.NewReader(l.ReadBuf)
	}
	return l.reader.Read(p)
}

func (l *LineBase) Next() IniLine{
	return l.next
}

func (l *LineBase) Previous() IniLine{
	return l.prev
}

func (l *LineBase) SetPrev(node IniLine) {
	l.prev = node 
}

func (l *LineBase) SetNext(node IniLine) {
	l.next = node 
}

type WhitespaceNode struct {
	content []byte
}

func (w *WhitespaceNode) Content() []byte {
	return w.content
}

type CommentNode struct {
	symbol byte
	content []byte
}

func (w *CommentNode) Content() []byte {
	return w.content
}

type HeaderNode struct {
	content []byte
}

func (w *HeaderNode) Content() []byte {
	return w.content
}

type KeyNode struct {
	content []byte
}

func (w *KeyNode) Content() []byte {
	return w.content
}

type ValueNode struct {
	content []byte
}

func (w ValueNode) Content() []byte {
	return w.content
}

type EmptyLine struct {
	LineBase

	Padding *WhitespaceNode
	Comment *CommentNode
}

func (l *EmptyLine) Read(p []byte) (n int, err error) {
	if !l.HasReader() {
		// Populate buffer
		l.ReadBuf = append(l.ReadBuf, l.Padding.content...)
		l.ReadBuf = append(l.ReadBuf, l.Comment.content...)
	}
	return l.LineBase.Read(p)
}

type SectionHeaderLine struct {
	LineBase

	Padding *WhitespaceNode
	Header *HeaderNode
	PostPad *WhitespaceNode
	Comment *CommentNode
}

func (l *SectionHeaderLine) Read(p []byte) (n int, err error) {
	if !l.HasReader() {
		// Populate buffer
		l.ReadBuf = append(l.ReadBuf, l.Padding.content...)
		l.ReadBuf = append(l.ReadBuf, l.Header.content...)
		l.ReadBuf = append(l.ReadBuf, l.PostPad.content...)
		l.ReadBuf = append(l.ReadBuf, l.Comment.content...)
	}
	return l.LineBase.Read(p)
}

type KeyValueLine struct {
	LineBase

	Padding *WhitespaceNode
	Key *KeyNode
	PostKeyPad *WhitespaceNode
	Value *ValueNode
	Comment *CommentNode
}

func (l *KeyValueLine) Read(p []byte) (n int, err error) {
	if !l.HasReader() {
		// Populate buffer
		l.ReadBuf = append(l.ReadBuf, l.Padding.content...)
		l.ReadBuf = append(l.ReadBuf, l.Key.content...)
		l.ReadBuf = append(l.ReadBuf, l.PostKeyPad.content...)
		l.ReadBuf = append(l.ReadBuf, B_EQUALS) // = 
		l.ReadBuf = append(l.ReadBuf, l.Value.content...)
		l.ReadBuf = append(l.ReadBuf, l.Comment.content...)
	}
	return l.LineBase.Read(p)
}

// isKeyByte checks if the input may be present in a Key
func isKeyByte(input byte) bool {
	switch {
	case input == B_BRACKET:
		return false
	case input == B_BRACKETCLOSE:
		return false
	case input <= B_US: // control characters (includes other whitespace)
		return false
	case input == B_SPACE:
		return false
	case input == B_EQUALS:
		// The parse order should take care of this, but including it just to be sure
		return false
	default:
		return true
	}
}

func isValueByte(input byte) bool {
	return input != B_NULL
}

// inQuotedString returns if `content` contains a currently unterminated quoted string
func inQuotedString(content []byte) bool {
	state := valueStringState(content)
	return state == VALUE_PARSE_QUOTED || state == VALUE_PARSE_QUOTED_BACKSLASH
}

// isClosedQuotedString returns if `content` contains a quoted string that has been closed
func isClosedQuotedString(content []byte) bool {
	state := valueStringState(content)
	return state == VALUE_PARSE_QUOTED_TERMINATED
}

// Check if all of content is whitespace
func allWhiteSpace(content []byte) bool {
	for _, b := range content {
		if TokenType(b) != Whitespace {
			return false
		}
	}
	return true
}

const VALUE_PARSE_WHITESPACE = 1
const VALUE_PARSE_QUOTED = 2
const VALUE_PARSE_QUOTED_BACKSLASH = 3
const VALUE_PARSE_QUOTED_TERMINATED = 4
const VALUE_PARSE_UNQUOTED = 9
const VALUE_PARSE_ERROR = 10

func valueStringState(content []byte) (state int) {
	state = VALUE_PARSE_WHITESPACE
	for _, token := range content {
		switch state {
		// We are currently still parsing whitespace
		case VALUE_PARSE_WHITESPACE:
			switch convertToken(token) {
				// Encounter another whitespace
				case Whitespace:
					continue
				// The string opens
				case Quote:
					state = VALUE_PARSE_QUOTED
				default:
					// The string opens unquoted, check if the token is valid
					if slices.Contains(validValueByteSetUnquoted, token) {
						state = VALUE_PARSE_UNQUOTED
					} else {
						// this character is not allowed
						return VALUE_PARSE_ERROR
					}
			}

		case VALUE_PARSE_QUOTED:
			// Start escaping on a backslash, terminate on a quote
			if token == B_BACKSLASH {
				state = VALUE_PARSE_QUOTED_BACKSLASH
				break
			}
			if token == B_QUOTE {
				state = VALUE_PARSE_QUOTED_TERMINATED
				break
			} 
			if slices.Contains(invalidValueByteSetQuoted, token) {
				return VALUE_PARSE_ERROR
			}
		
		case VALUE_PARSE_QUOTED_TERMINATED:
			// Only allow whitespace after a quoted string terminates
			if convertToken(token) != Whitespace {
				return VALUE_PARSE_ERROR
			}
		
	
		case VALUE_PARSE_QUOTED_BACKSLASH:
			// After a backslash basically anything is allowed except invalid bytes
			if slices.Contains(invalidValueByteSetQuoted, token) {
				return VALUE_PARSE_ERROR

			}
			state = VALUE_PARSE_QUOTED
			
		case VALUE_PARSE_UNQUOTED:
			// Check for invalid bytes
			if slices.Contains(invalidValueByteSetUnquoted, token) {
				return VALUE_PARSE_ERROR
			}
		default:
			panic("unknown state")
		}

	}
	return
}


// isExtraQuoteLegal returns if `content` would still be legal after adding a quote
//
// This means concretely either:
// The string is empty
// The string has only whitespace 
// The string is a non-terminated quoted string
// The last content character is an escaping backslash that is itself not escaped
func isExtraQuoteLegal(content []byte) bool {
	state := valueStringState(content)
	return state != VALUE_PARSE_QUOTED_TERMINATED && 
		   state != VALUE_PARSE_UNQUOTED &&
		   state != VALUE_PARSE_ERROR
}
