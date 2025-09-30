package montoya

import (
	"bytes"
	"io"
	"slices"
)

// An IniLine is any line type in an Inifile
type IniLine interface {
	io.Reader
	// Reset resets the reader state so the line may be re-read
	Reset()
	// Terminated tests if all nodes on the line were properly terminated
	Terminated() bool
	// Next yields the next line
	Next() IniLine
	// Next yields the previous line
	Previous() IniLine
	// SetNext sets the next line
	SetNext(node IniLine)
	// SetPrev sets the previous line
	SetPrev(node IniLine)
}

// An IniNode is part of an IniLine
type IniNode interface {
	Content() []byte
}

// LineBase is the base struct for iniLines and forms a node in a doubly linked list
type LineBase struct {
	// ReadBuf is a buffer populated when the line is being read
	ReadBuf []byte
	// reader contains the reader state when the line is being read
	reader io.Reader

	// prev is a link to the previous iniLine
	prev IniLine
	// next is a link to the next iniLine
	next IniLine
}

// Reset resets reader state
func (l *LineBase) Reset() {
	l.ReadBuf = []byte{}
	l.reader = nil
}

// HasReader returns whether the line is currently being read
func (l *LineBase) HasReader() bool {
	return l.reader != nil
}

// Read implements io.Reader for `LineBase`
func (l *LineBase) Read(p []byte) (n int, err error) {
	if l.reader == nil {
		l.reader = bytes.NewReader(l.ReadBuf)
	}
	return l.reader.Read(p)
}

// Next returns the next IniLine
func (l *LineBase) Next() IniLine {
	return l.next
}

// Previous returns the previous IniLine
func (l *LineBase) Previous() IniLine {
	return l.prev
}

// SetPrev sets the previous IniLine node
func (l *LineBase) SetPrev(node IniLine) {
	l.prev = node
}

// SetNext sets the next InitLine node
func (l *LineBase) SetNext(node IniLine) {
	l.next = node
}

// WhitespaceNode is whitespace in an IniLine
type WhitespaceNode struct {
	content []byte
}

// Content returns the node's content
func (w *WhitespaceNode) Content() []byte {
	return w.content
}

// CommentNode is a comment in an IniLine
type CommentNode struct {
	// symbol is the symbol indicating the comment from  `commentStartBytes`
	symbol byte
	// content contains the comment content, without the comment start symbol
	content []byte
}

// Content returns the node's content
func (w *CommentNode) Content() []byte {
	return w.content
}

// HeaderNode is a header in an IniLine denoting a section
type HeaderNode struct {
	// content contains the header name, without brackets
	content []byte
}

// Content returns the node's content
func (w *HeaderNode) Content() []byte {
	return w.content
}

// KeyNode contains a key in a KeyValueLine
type KeyNode struct {
	// content contains the key content
	content []byte
}

// Content returns the node's content
func (w *KeyNode) Content() []byte {
	return w.content
}

// ValueNode is a Value inside a KeyValueLine
type ValueNode struct {
	// content contains the value
	content []byte
}

// Content returns the node's content
func (w ValueNode) Content() []byte {
	return w.content
}

// EmptyLine is an iniLine containing an optional comment
type EmptyLine struct {
	LineBase

	Padding *WhitespaceNode
	Comment *CommentNode
}

// Read implements io.Reader for `EmptyLine`
func (l *EmptyLine) Read(p []byte) (n int, err error) {
	if !l.HasReader() {
		// Populate buffer
		l.ReadBuf = append(l.ReadBuf, l.Padding.content...)
		l.ReadBuf = append(l.ReadBuf, l.Comment.content...)
	}
	return l.LineBase.Read(p)
}

// Terminated is always true for an EmptyLine
func (l *EmptyLine) Terminated() bool {
	return true
}

// Section is an iniLine containing a section header and optional comment
type SectionHeaderLine struct {
	LineBase

	Padding *WhitespaceNode
	Header  *HeaderNode
	PostPad *WhitespaceNode
	Comment *CommentNode
}

// Read implements io.Reader for `SectionHeaderLine`
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

// Terminated indicates whether a SectionHeaderLine was properly terminated
func (l *SectionHeaderLine) Terminated() bool {
	// the parser adds a PostPad node when the section header terminates
	return l.PostPad != nil
}

// KeyValueLine is an iniLine containing a key, value and optional comment
type KeyValueLine struct {
	LineBase

	Padding    *WhitespaceNode
	Key        *KeyNode
	PostKeyPad *WhitespaceNode
	Value      *ValueNode
	Comment    *CommentNode
}

// Read implements io.Reader for `KeyValueLine`
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

// Terminated indicates whether a KeyValueLine was properly terminated
func (l *KeyValueLine) Terminated() bool {
	if l.Value == nil {
		// parser never saw B_EQUALS
		return false
	}

	state := valueStringState(l.Value.content)
	return (state == VALUE_PARSE_WHITESPACE || state == VALUE_PARSE_QUOTED_TERMINATED || state == VALUE_PARSE_UNQUOTED)
}

// isKeyByte checks if the input may be present in a Key
func isKeyByte(input byte) bool {
	return slices.Contains(validKeyByteSet, input)
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

const VALUE_PARSE_WHITESPACE = 1        // The value consists entirely of whitespace
const VALUE_PARSE_QUOTED = 2            // The value is an open, quoted string
const VALUE_PARSE_QUOTED_BACKSLASH = 3  // The value is an open, quoted string and the last byte is an escape backslash
const VALUE_PARSE_QUOTED_TERMINATED = 4 // The Value is a terminated quoted string
const VALUE_PARSE_UNQUOTED = 9          // The value is an unquoted string
const VALUE_PARSE_ERROR = 10            // The value contains illegal bytes

// valuestringState uses a state machine to determine the state of the currently parsed value string
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
