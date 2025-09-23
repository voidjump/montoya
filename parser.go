package montoya

import (
	"bytes"
	"fmt"
	"io"
)


const (
	// LineStart means we are starting to parse a new line
	LineStart = 0
	// ParseKV means we are currently parsing a key-value pair
	ParseKV = 1
	// ParseComment means we encountered a comment
	ParseComment = 2
	// ParseHeader means we are starting a header
	ParseHeaderStart = 3
	// ParseHeaderContent means we are currently parsing the header key
	ParseSectionName = 4
	// ParseSectionEnd means we are currently parsing the section end
	ParseHeaderEnd = 5
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

// iniParser state
type iniParser struct {
	// file is the IniFile representation being parsed into
	file *IniFile

	// input is the data source to be consumed
	input io.Reader

	// lineNo and colNo keep track of the current parser position
	lineNo, colNo int
	// currentNode is the current node being parsed
	currentNode IniNode
	// currentLine and previousLine hold references to the line object in the file
	currentLine, previousLine IniLine
	// tokenType is the current token type
	tokenType TokenType
	// currentByte is the raw byte value of the current token
	currentByte byte
}

// Parse consumes the input and returns a parsed IniFile 
func Parse(input io.Reader) (*IniFile, error) {
	parser := &iniParser{
		input: input,
		file: &IniFile{},
	}
	return parser.parse()
}

// Err returns a parsing error
func (p *iniParser) Err(text string) error {
	return fmt.Errorf("%s line:%v, col:%v", text, p.lineNo, p.colNo)
}

// parseEmpty expects to parse current token into an EmptyLine object
// 
// An emptyline looks like this:
// <Whitespace (optional)><Comment (optional)>[\n]
//
func (p *iniParser) parseEmptyLine(line *EmptyLine) error {
	switch node := p.currentNode.(type) {
		case *WhitespaceNode:
			switch p.tokenType {
			case Whitespace:
				// Grow the whitespace
				node.content = append(node.content, p.currentByte)
			case Hash:
				fallthrough
			case SemiColon:
				// Save the padding (if any)
				line.Padding = node 
				// Start a new comment
				p.currentNode = &CommentNode{ 
					symbol: p.currentByte,
					content: []byte{},
				}

			case BracketOpen:
				// Switch current line type to section line
				headerLine := &SectionHeaderLine{
					// Preserve padding, if any
					Padding : node,
					Header: &HeaderNode{
						content: []byte{},
					},
				}
				p.currentNode = headerLine.Header
				p.currentLine = headerLine
			default:
				// Check if this byte is a valid KeyByte
				if isKeyByte(p.currentByte) {
					// Switch current line type to KeyValueLine
					keyValueLine := &KeyValueLine{
						// Preserve padding, if any
						Padding: node,
						Key: &KeyNode{
							content : []byte{p.currentByte},
						},
					}
					p.currentLine = keyValueLine
					p.currentNode = keyValueLine.Key
				} else {
					return p.Err("invalid character")
				}
			}
		case *CommentNode:
			// Anything goes in a comment ;)
			node.content = append(node.content, p.currentByte)
	}
	return nil
}

// parseKeyValueLine expects to parse current token into a KeyValueLine object
// 
// A KeyValueLine looks like this:
// <Whitespace (optional)><Key>[=]<SectionHeader><Comment (optional>[\n]
//
// Since the Parser prioritizes an EmptyLine first, Padding is never parsed in
// this function
func (p *iniParser) parseKeyValueLine(line *KeyValueLine) error {
	switch node := p.currentNode.(type) {
				case *KeyNode:
					if p.tokenType == Equals {
						// Transition to value
						p.currentNode = &ValueNode{}
						break
					}
					if isKeyByte(p.currentByte) {
						// Grow key
						node.content = append(node.content, p.currentByte)
					} else {
						return p.Err("invalid character in key") 
					}
				case *ValueNode:
					// If we encounter a comment symbol transfer to comment node
					if (p.tokenType == Hash || p.tokenType == SemiColon) {
						// Check if we are in a quoted string
						if inQuotedString(node.content) {
							// simply append
							node.content = append(node.content, p.currentByte)
						} else{
							// Start a new comment
							p.currentNode = &CommentNode{ 
								symbol: p.currentByte,
								content: []byte{},
							}
						}
						break
					}
					if p.tokenType == Quote {
						// Check if adding an extra quote is legal
						if isExtraQuoteLegal(node.content) {
							node.content = append(node.content, p.currentByte)
						} else {
							return p.Err("illegal quote character in value")
						}
						break
					}
					if p.tokenType == Whitespace {
						node.content = append(node.content, p.currentByte)
						break
					}
					if isValueByte(p.currentByte) {
						if isClosedQuotedString(node.content) {
							return p.Err("illegal character after terminated quoted value")
						}
						node.content = append(node.content, p.currentByte)
						break
					} else {
						return p.Err("illegal character in value")
					}
				case *CommentNode:
					// Anything goes in a comment ;)
					node.content = append(node.content, p.currentByte)
					

			}
			return nil
}

// parseSectionHeaderLine expects to parse current token into a SectionHeaderLine object
// 
// A SectionHeader looks like this:
// <Whitespace (optional)>[[]<HeaderNode>[]]<Whitespace(Optional)><Value>[\n]
//
// Since the Parser prioritizes an EmptyLine first, Padding is never parsed in
// this function
func (p *iniParser) parseSectionHeaderLine(line *SectionHeaderLine) error {
	switch node := p.currentNode.(type) {
		case *CommentNode:
			// Anything goes in a comment ;)
			node.content = append(node.content, p.currentByte)
		case *HeaderNode:
			switch p.tokenType {
			case BracketClose:
				// Transition to PostPad 
				p.currentNode = &WhitespaceNode{}
			case Hash:
				fallthrough
			case SemiColon:
				return p.Err("illegal comment start in bracket")
			default:
				// Grow the header content
				node.content = append(node.content, p.currentByte)
			}
		case *WhitespaceNode:
			// This must be the PostPad
			switch p.tokenType {
			case Hash:
				fallthrough
			case SemiColon:
				// Transition to comment
				p.currentNode = &CommentNode{ 
					symbol: p.currentByte,
					content: []byte{},
				}
			case Whitespace:
				// Grow
				node.content = append(node.content, p.currentByte)
			default:
				return p.Err("illegal non-comment character after closed section header")
			}
		
	}
	return nil
}

func (p *iniParser) advanceLine() {

	// TODO: We should check if nodes on the line are properly terminated 
	if p.previousLine != nil {
		p.previousLine.SetNext(p.currentLine)
		p.currentLine.SetPrev(p.previousLine)
	} else {
		// Link up first line
		p.file.Head = p.currentLine
	}

	// move currentLine cursor
	p.previousLine = p.currentLine

	// Start a new clear line and node
	p.currentLine = &EmptyLine{}
	p.currentNode = &WhitespaceNode{}

	// Track position
	p.lineNo += 1
	p.colNo = 0

}

// parse consumes an io.Reader into a parsed IniFile
func (p *iniParser) parse() (*IniFile, error) {

	p.currentNode = &WhitespaceNode{}
	p.currentLine = &EmptyLine{}

	buf := make([]byte, 1) // slice of length 1
	for {
		_, err := p.input.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read input: %w", err)
		}

		p.currentByte = buf[0]
		p.tokenType = convertToken(p.currentByte)

		// Finish up current line and link up lines
		if p.tokenType == NewLine {
			p.advanceLine()
		}

		switch line := p.currentLine.(type) {
			case *EmptyLine:
				err = p.parseEmptyLine(line)

			case *KeyValueLine:
				err = p.parseKeyValueLine(line)
			
			case *SectionHeaderLine:
				err = p.parseSectionHeaderLine(line)
			default:
				panic("invalid line type")
		}
		if err != nil {
			return nil, err
		}

		// Track position
		p.colNo += 1
	}

	p.file.Tail = p.currentLine

	return p.file, nil
}
