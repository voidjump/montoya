package montoya

import (
	"fmt"
	"io"
)

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
		file:  &IniFile{},
	}
	return parser.parse()
}

// Err returns a parsing error
func (p *iniParser) Err(text string) error {
	return fmt.Errorf("%s (line:%v, col:%v)", text, p.lineNo, p.colNo)
}

func (p *iniParser) debug(msg string) {
	// fmt.Printf("%s byte:%02x, node: %p, line: %p\n", msg, p.currentByte, p.currentNode, p.currentLine)
}

// parseEmpty expects to parse current token into an EmptyLine object
//
// An emptyline looks like this:
// <Whitespace (optional)><Comment (optional)>[\n]
func (p *iniParser) parseEmptyLine(line *EmptyLine) error {
	switch node := p.currentNode.(type) {
	case *WhitespaceNode:
		switch p.tokenType {
		case Whitespace:
			// Grow the whitespace
			node.content = append(node.content, p.currentByte)
		case CommentStart:
			p.debug("starting comment")
			// Save the padding (if any)
			line.Padding = node
			// Start a new comment
			line.Comment = &CommentNode{
				symbol:  p.currentByte,
				content: []byte{},
			}
			p.currentNode = line.Comment

		case SectionStart:
			// Switch current line type to section line
			headerLine := &SectionHeaderLine{
				// Preserve padding, if any
				Padding: node,
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
						content: []byte{p.currentByte},
					},
				}
				p.currentLine = keyValueLine
				p.currentNode = keyValueLine.Key
			} else {
				return p.Err(fmt.Sprintf("invalid character %02x for empty line", p.currentByte))
			}
		}
	case *CommentNode:
		// Anything goes in a comment ;)

		p.debug("appending")
		fmt.Printf("appending %02x\n", p.currentByte)
		fmt.Printf("node: %p\n", node)
		fmt.Printf("p.currentNode: %p\n", p.currentNode)

		node.content = append(node.content, p.currentByte)
	}
	p.debug("leaving func")
	return nil
}

// parseKeyValueLine expects to parse current token into a KeyValueLine object
//
// A KeyValueLine looks like this:
// <Whitespace (optional)><Key>[=]<Value><Comment (optional>[\n]
//
// Since the Parser prioritizes an EmptyLine first, initial Padding is never parsed here
func (p *iniParser) parseKeyValueLine(line *KeyValueLine) error {
	switch node := p.currentNode.(type) {
	case *KeyNode:
		if p.tokenType == Equals {
			// Transition to value
			line.Value = &ValueNode{}
			p.currentNode = line.Value
			break
		}
		if p.tokenType == Whitespace {
			// Transition to PostKeyPad
			line.PostKeyPad = &WhitespaceNode{content: []byte{p.currentByte}}
			p.currentNode = line.PostKeyPad
			break
		}
		if isKeyByte(p.currentByte) {
			// Grow key
			node.content = append(node.content, p.currentByte)
		} else {
			return p.Err(fmt.Sprintf("invalid character %02x in key", p.currentByte))
		}
	// This must be post-key whitespace
	case *WhitespaceNode:
		if p.tokenType == Equals {
			// Transition to value
			line.Value = &ValueNode{}
			p.currentNode = line.Value
			break
		}
		if p.tokenType == Whitespace {
			// Grow
			node.content = append(node.content, p.currentByte)
			break
		}
		return p.Err(fmt.Sprintf("invalid non-whitespace character %02x in key", p.currentByte))

	case *ValueNode:
		// If we encounter a comment symbol transfer to comment node
		if p.tokenType == CommentStart {
			// Check if we are in a quoted string
			if inQuotedString(node.content) {
				// simply append
				node.content = append(node.content, p.currentByte)
			} else {
				// Start a new comment
				line.Comment = &CommentNode{
					symbol:  p.currentByte,
					content: []byte{},
				}
				p.currentNode = line.Comment
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
		if p.currentByte != B_NULL {
			if isClosedQuotedString(node.content) {
				return p.Err(fmt.Sprintf("illegal character %02x after terminated quoted value", p.currentByte))
			}
			node.content = append(node.content, p.currentByte)
			break
		} else {
			return p.Err(fmt.Sprintf("illegal character %02x in value", p.currentByte))
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
		case SectionEnd:
			// Transition to PostPad
			line.PostPad = &WhitespaceNode{}
			p.currentNode = line.PostPad
		case CommentStart:
			return p.Err("illegal comment start in bracket")
		default:
			// Grow the header content
			node.content = append(node.content, p.currentByte)
		}
	case *WhitespaceNode:
		// This must be the PostPad
		switch p.tokenType {
		case CommentStart:
			// Transition to comment
			line.Comment = &CommentNode{
				symbol:  p.currentByte,
				content: []byte{},
			}
			p.currentNode = line.Comment

		case Whitespace:
			// Grow
			node.content = append(node.content, p.currentByte)
		default:
			return p.Err(fmt.Sprintf("illegal non-comment character %02x after closed section header", p.currentByte))
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
	// TODO a line without a newline but EOF is not currently terminated

	whiteSpace := &WhitespaceNode{} // we need the concrete type
	p.currentNode = whiteSpace
	p.currentLine = &EmptyLine{Padding: whiteSpace}

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
			continue
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
