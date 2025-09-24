package montoya

import (
	"bytes"
	"io"
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
	Value *ValueNode
	Comment *CommentNode
}

func (l *KeyValueLine) Read(p []byte) (n int, err error) {
	if !l.HasReader() {
		// Populate buffer
		l.ReadBuf = append(l.ReadBuf, l.Padding.content...)
		l.ReadBuf = append(l.ReadBuf, l.Key.content...)
		l.ReadBuf = append(l.ReadBuf, l.Value.content...)
		l.ReadBuf = append(l.ReadBuf, l.Comment.content...)
	}
	return l.LineBase.Read(p)
}

// isKeyByte checks if the input may be present in a Key
func isKeyByte(input byte) bool {
	switch {
	case input == '[':
		return false
	case input == ']':
		return false
	case input <= 0x1F: // control characters
		return false
	default:
		return true
	}
}

func isValueByte(input byte) bool {
	return input != 0x00
}

// inQuotedString returns if `content` contains a currently unterminated quoted string
func inQuotedString(content []byte) bool {
	whitespace := true
	inString := false
	escape := false
	for _, contentByte := range content {
		// skip al initial whitespace
		if whitespace {
			if TokenType(contentByte) != Whitespace {
				whitespace = false
			}
			continue
		}
		// Only continue if the first non-whitespace character is a "
		if !inString {
			if contentByte == '"' {
				inString = true
				continue
			}
			return false
		}
		if !escape {
			if contentByte == '\\' {
				escape = true
				continue
			}
			if contentByte == '"' {
				return false
			}
			continue
		} 		
		// this character is escaped, and then we look at the next
		escape = false
	}
	return true 
}

// isClosedQuotedString returns if `content` contains a quoted string that has been closed
func isClosedQuotedString(content []byte) bool {
	whitespace := true
	inString := false
	escape := false
	for _, contentByte := range content {
		// skip al initial whitespace
		if whitespace {
			if TokenType(contentByte) != Whitespace {
				whitespace = false
			}
			continue
		}
		// Only continue if the first non-whitespace character is a "
		if !inString {
			if contentByte == '"' {
				inString = true
				continue
			}
			return false
		}
		if !escape {
			if contentByte == '\\' {
				escape = true
				continue
			}
			if contentByte == '"' {
				return true 
			}
			continue
		} 		
		// this character is escaped, and then we look at the next
		escape = false
	}
	return false 
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

// isExtraQuoteLegal returns if `content` would still be legal after adding a quote
//
// This means concretely either:
// The string is empty
// The string has only whitespace 
// The string is a non-terminated quoted string
// The last content character is an escaping backslash that is itself not escaped
func isExtraQuoteLegal(content []byte) bool {
	if len(content) == 0 {
		return true
	}
	if allWhiteSpace(content) {
		return true
	}
	if inQuotedString(content) {
		return true
	}

	// check if content ends with unescaped backslash
	i := len(content) - 1
    count := 0
    for i >= 0 && content[i] == '\\' {
        count++
        i--
    }
	return count % 2 == 1 
}
