package montoya

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: Add unit tests for Terminate cases
// TODO: Fuzz test, tests sometimes failing

// testParse runs Parse with the input byte string, constructing a reader from it
func testParse(input ...any) (*IniFile, error) {
	var buf bytes.Buffer
	for _, source := range input {
		switch castSource := source.(type) {
		case byte:
			buf.WriteByte(castSource)
		case []byte:
			buf.Write(castSource)
		case string:
			buf.WriteString(castSource)
		default:
			panic("Invalid type passed to testParse")
		}
	}
	reader := bytes.NewReader(buf.Bytes())
	return Parse(reader)
}

// Test an empty file yields no lines
func TestParseEmptyFile(t *testing.T) {
	file, err := testParse([]byte{})

	assert.NoError(t, err)
	assert.NotNil(t, file)

	line := file.Head

	assert.Nil(t, line)
}

// Test the parser errors are formatted correctly
func TestParseErrFormat(t *testing.T) {
	p := iniParser{
		lineNo: 42,
		colNo:  80,
	}
	err := p.Err("test error")
	assert.Error(t, err)
	assert.ErrorContains(t, err, "test error (line:42, col:80)")
}

////////////////////////////////////////////////////////////////////////////////
// EMPTY LINE CASES ////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// Test only a newline yields an empty line with an empty whitespace node
func TestParseEmptyLine(t *testing.T) {
	file, err := testParse(B_NEWLINE)

	assert.NoError(t, err)
	assert.NotNil(t, file)

	line := file.Head

	assert.NotNil(t, line)

	concrete, ok := line.(*EmptyLine)
	require.True(t, ok)

	assert.NotNil(t, concrete)

	assert.NotNil(t, concrete.Padding)
	assert.Nil(t, concrete.Comment)
	assert.Len(t, concrete.Padding.content, 0)
}

// Test only whitespace and a newline yields a whitespace node
func TestParseEmptyLineWhitespace(t *testing.T) {
	padding := fuzzWhiteSpace(10)
	file, err := testParse(padding, B_NEWLINE)

	assert.NoError(t, err)
	assert.NotNil(t, file)

	line := file.Head

	assert.NotNil(t, line)

	concrete, ok := line.(*EmptyLine)
	require.True(t, ok)

	assert.NotNil(t, concrete)

	assert.NotNil(t, concrete.Padding)
	assert.Nil(t, concrete.Comment)

	whiteSpaceResult := concrete.Padding
	assert.Equal(t, padding, whiteSpaceResult.content)
}

// Test only comment and a newline yields a comment node and empty whitespace node
func TestParseEmptyLineComment(t *testing.T) {
	comment := fuzzComment()
	file, err := testParse(comment, B_NEWLINE)

	assert.NoError(t, err)
	assert.NotNil(t, file)

	line := file.Head

	assert.NotNil(t, line)

	concrete, ok := line.(*EmptyLine)
	require.True(t, ok)

	assert.NotNil(t, concrete)

	assert.NotNil(t, concrete.Padding)
	assert.NotNil(t, concrete.Comment)

	assert.Equal(t, comment[0], concrete.Comment.symbol)
	assert.Equal(t, comment[1:], concrete.Comment.content)
}

// Test a comment with some whitespace yields the whitespace and the comment node
func TestParseEmptyLineCommentAndPadding(t *testing.T) {
	padding := fuzzWhiteSpace(20)
	comment := fuzzComment()
	file, err := testParse(padding, comment, B_NEWLINE)

	assert.NoError(t, err)
	assert.NotNil(t, file)

	line := file.Head

	assert.NotNil(t, line)

	concrete, ok := line.(*EmptyLine)
	require.True(t, ok)

	assert.NotNil(t, concrete)

	assert.NotNil(t, concrete.Padding)
	assert.NotNil(t, concrete.Comment)

	assert.Equal(t, padding, concrete.Padding.content)

	assert.Equal(t, comment[0], concrete.Comment.symbol)
	assert.Equal(t, comment[1:], concrete.Comment.content)
}

// Test a double comment still only yields one comment
func TestParseEmptyLineOnlyASingleComment(t *testing.T) {
	comment := string(fuzzComment()) + string(fuzzComment())
	file, err := testParse(comment, B_NEWLINE)

	assert.NoError(t, err)
	assert.NotNil(t, file)

	line := file.Head

	assert.NotNil(t, line)

	concrete, ok := line.(*EmptyLine)
	require.True(t, ok)

	assert.NotNil(t, concrete)

	assert.NotNil(t, concrete.Comment)
	assert.NotNil(t, concrete.Padding)

	assert.Equal(t, []byte(comment)[0], concrete.Comment.symbol)
	assert.Equal(t, []byte(comment)[1:], concrete.Comment.content)
}

// Test illegal key byte on an empty line creates an error
func TestIllegalKeyByteOnEmptyLineIsError(t *testing.T) {
	lineError := fuzzFromSet(invalidKeyByteSet)
	file, err := testParse(lineError, B_NEWLINE)

	assert.Error(t, err)
	assert.Nil(t, file)

	assert.ErrorContains(t, err, fmt.Sprintf("invalid character %02x for empty line", lineError[0]))
}

////////////////////////////////////////////////////////////////////////////////
// SECTION LINE CASES //////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// Test a section header with a newline yields a section node and an empty padding node
func TestParseSimpleSectionNoPadding(t *testing.T) {
	section := fuzzSection(false)
	file, err := testParse(B_BRACKET, section, B_BRACKETCLOSE, B_NEWLINE)

	assert.NoError(t, err)
	assert.NotNil(t, file)

	line := file.Head

	assert.NotNil(t, line)

	concrete, ok := line.(*SectionHeaderLine)
	require.True(t, ok)

	assert.NotNil(t, concrete)

	assert.NotNil(t, concrete.Padding)
	assert.NotNil(t, concrete.Header)
	assert.NotNil(t, concrete.PostPad)
	assert.Nil(t, concrete.Comment)

	assert.Len(t, concrete.Padding.content, 0)
	assert.Len(t, concrete.PostPad.content, 0)
	assert.Equal(t, section, concrete.Header.content)
}

// Test a section header and pre-header padding yields the section and the whitespace node, and empty Postpad
func TestParseSimpleSectionWithPadding(t *testing.T) {
	padding := fuzzWhiteSpace(20)
	section := fuzzSection(false)

	file, err := testParse(padding, B_BRACKET, section, B_BRACKETCLOSE, B_NEWLINE)

	assert.NoError(t, err)
	assert.NotNil(t, file)

	line := file.Head

	assert.NotNil(t, line)

	concrete, ok := line.(*SectionHeaderLine)
	require.True(t, ok)

	assert.NotNil(t, concrete)

	assert.NotNil(t, concrete.Padding)
	assert.NotNil(t, concrete.Header)
	assert.NotNil(t, concrete.PostPad)
	assert.Nil(t, concrete.Comment)

	assert.Len(t, concrete.PostPad.content, 0)
	assert.Equal(t, padding, concrete.Padding.content)
	assert.Equal(t, section, concrete.Header.content)
}

// Test a section header with whitespace before and after the section yields all 3 nodes
func TestParseSimpleSectionWithPaddingAndPostPad(t *testing.T) {
	padding := fuzzWhiteSpace(10)
	padding2 := fuzzWhiteSpace(10)
	section := fuzzSection(false)

	file, err := testParse(padding, B_BRACKET, section, B_BRACKETCLOSE, padding2, B_NEWLINE)

	assert.NoError(t, err)
	assert.NotNil(t, file)

	line := file.Head

	assert.NotNil(t, line)

	concrete, ok := line.(*SectionHeaderLine)
	require.True(t, ok)

	assert.NotNil(t, concrete)

	assert.NotNil(t, concrete.Padding)
	assert.NotNil(t, concrete.Header)
	assert.NotNil(t, concrete.PostPad)
	assert.Nil(t, concrete.Comment)

	assert.Equal(t, padding, concrete.Padding.content)
	assert.Equal(t, padding2, concrete.PostPad.content)
	assert.Equal(t, section, concrete.Header.content)
}

// Test a comment after a section is populated
func TestParseSimpleSectionWithWhitespaceAndComment(t *testing.T) {
	padding := fuzzWhiteSpace(10)
	padding2 := fuzzWhiteSpace(10)
	comment := fuzzComment()
	section := fuzzSection(false)
	file, err := testParse(padding, B_BRACKET, section, B_BRACKETCLOSE, padding2, comment, B_NEWLINE)

	assert.NoError(t, err)
	assert.NotNil(t, file)

	line := file.Head

	assert.NotNil(t, line)

	concrete, ok := line.(*SectionHeaderLine)
	require.True(t, ok)

	assert.NotNil(t, concrete)

	assert.NotNil(t, concrete.Padding)
	assert.NotNil(t, concrete.Header)
	assert.NotNil(t, concrete.PostPad)
	assert.NotNil(t, concrete.Comment)

	assert.Equal(t, padding, concrete.Padding.content)
	assert.Equal(t, padding2, concrete.PostPad.content)
	assert.Equal(t, section, concrete.Header.content)
	assert.Equal(t, comment[1:], concrete.Comment.content)
	assert.Equal(t, comment[0], concrete.Comment.symbol)
}

// Test a comment-ed out section does not yield a SectionHeaderLine
func TestCommentedSectionYieldsEmptyLine(t *testing.T) {
	commentByte := fuzzyChoice(commentStartBytes)
	section := fuzzSection(false)
	file, err := testParse(commentByte, B_BRACKET, section, B_BRACKETCLOSE, B_NEWLINE)

	assert.NoError(t, err)
	assert.NotNil(t, file)

	line := file.Head

	assert.NotNil(t, line)

	assert.IsType(t, &EmptyLine{}, line)
}

// Test a comment in a section is not allowed
func TestCommentInHeaderNotAllowed(t *testing.T) {
	section := fuzzSection(false)
	comment := fuzzComment()
	file, err := testParse(B_BRACKET, section, comment, B_BRACKETCLOSE, B_NEWLINE)

	assert.Error(t, err)
	assert.Nil(t, file)
	assert.ErrorContains(t, err, "illegal comment start in bracket")
}

// Test non-whitespace after a section and before comment is not allowed
func TestNonWhitespaceAfterHeaderNotAllowed(t *testing.T) {
	section := fuzzSection(false)
	comment := fuzzComment()
	illegalByte := fuzzyChoice(invalidWhitespaceByteset)
	file, err := testParse(B_BRACKET, section, B_BRACKETCLOSE, illegalByte, comment, B_NEWLINE)

	assert.Error(t, err)
	assert.Nil(t, file)
	assert.ErrorContains(t, err, fmt.Sprintf("illegal non-comment character %02x after closed section header", illegalByte))
}

////////////////////////////////////////////////////////////////////////////////
// KEY VALUE LINE CASES ////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// Test that a simple key value line is parsed correctly
func TestParseKeyValueLine(t *testing.T) {
	key := fuzzKey()
	value := fuzzValue(false)
	file, err := testParse(key, B_EQUALS, value, B_NEWLINE)

	assert.NoError(t, err)
	assert.NotNil(t, file)

	line := file.Head

	assert.NotNil(t, line)

	concrete, ok := line.(*KeyValueLine)
	require.True(t, ok)

	assert.NotNil(t, concrete)

	assert.NotNil(t, concrete.Padding)
	assert.NotNil(t, concrete.Key)
	assert.NotNil(t, concrete.Value)
	assert.Nil(t, concrete.PostKeyPad)
	assert.Nil(t, concrete.Comment)

	assert.Len(t, concrete.Padding.content, 0)
	assert.Equal(t, key, concrete.Key.content)
	assert.Equal(t, value, concrete.Value.content)
}

// Test that a simple key value line with pre-key whitespace is parsed correctly
func TestParseKeyValueLinePreWhitespace(t *testing.T) {
	whitespace := fuzzWhiteSpace(10)
	key := fuzzKey()
	value := fuzzValue(false)
	file, err := testParse(whitespace, key, B_EQUALS, value, B_NEWLINE)

	assert.NoError(t, err)
	assert.NotNil(t, file)

	line := file.Head

	assert.NotNil(t, line)

	concrete, ok := line.(*KeyValueLine)
	require.True(t, ok)

	assert.NotNil(t, concrete)

	assert.NotNil(t, concrete.Padding)
	assert.NotNil(t, concrete.Key)
	assert.NotNil(t, concrete.Value)
	assert.Nil(t, concrete.Comment)

	assert.Equal(t, whitespace, concrete.Padding.content)
	assert.Equal(t, key, concrete.Key.content)
	assert.Equal(t, value, concrete.Value.content)
}

// Test that post-Key whitespace is parsed correctly
func TestParseKeyValueLinePostWhitespace(t *testing.T) {
	whitespace := fuzzWhiteSpace(10)
	key := fuzzKey()
	value := fuzzValue(false)
	file, err := testParse(key, whitespace, B_EQUALS, value, B_NEWLINE)

	assert.NoError(t, err)
	assert.NotNil(t, file)

	line := file.Head

	assert.NotNil(t, line)

	concrete, ok := line.(*KeyValueLine)
	require.True(t, ok)

	assert.NotNil(t, concrete)

	assert.NotNil(t, concrete.Padding)
	assert.NotNil(t, concrete.Key)
	assert.NotNil(t, concrete.PostKeyPad)
	assert.NotNil(t, concrete.Value)
	assert.Nil(t, concrete.Comment)

	assert.Len(t, concrete.Padding.content, 0)
	assert.Equal(t, key, concrete.Key.content)
	assert.Len(t, whitespace, 10)
	assert.Equal(t, whitespace, concrete.PostKeyPad.content)
	assert.Equal(t, value, concrete.Value.content)
}

// Test a key value line with a comment is parsed correctly
func TestParseKeyValueComment(t *testing.T) {
	whitespace := fuzzWhiteSpace(10)
	key := fuzzKey()
	value := fuzzValue(false)
	comment := fuzzComment()
	file, err := testParse(key, whitespace, B_EQUALS, value, comment, B_NEWLINE)

	assert.NoError(t, err)
	assert.NotNil(t, file)

	line := file.Head

	assert.NotNil(t, line)

	concrete, ok := line.(*KeyValueLine)
	require.True(t, ok)

	assert.NotNil(t, concrete)

	assert.NotNil(t, concrete.Padding)
	assert.NotNil(t, concrete.Key)
	assert.NotNil(t, concrete.PostKeyPad)
	assert.NotNil(t, concrete.Value)
	assert.NotNil(t, concrete.Comment)

	assert.Len(t, concrete.Padding.content, 0)
	assert.Equal(t, key, concrete.Key.content)
	assert.Equal(t, whitespace, concrete.PostKeyPad.content)
	assert.Equal(t, value, concrete.Value.content)
	assert.Equal(t, comment[1:], concrete.Comment.content)
	assert.Equal(t, comment[0], concrete.Comment.symbol)
}

// Test a value can be inside a unquoted string
func TestUnQuotedValue(t *testing.T) {
	whitespace := fuzzWhiteSpace(10)
	key := fuzzKey()
	value := fuzzValue(false)
	comment := fuzzComment()
	file, err := testParse(key, whitespace, B_EQUALS, value, comment, B_NEWLINE)

	assert.NoError(t, err)
	assert.NotNil(t, file)

	line := file.Head

	assert.NotNil(t, line)

	concrete, ok := line.(*KeyValueLine)
	require.True(t, ok)

	assert.NotNil(t, concrete)

	assert.NotNil(t, concrete.Padding)
	assert.NotNil(t, concrete.Key)
	assert.NotNil(t, concrete.PostKeyPad)
	assert.NotNil(t, concrete.Value)
	assert.NotNil(t, concrete.Comment)

	assert.Len(t, concrete.Padding.content, 0)
	assert.Equal(t, key, concrete.Key.content)
	assert.Equal(t, whitespace, concrete.PostKeyPad.content)
	assert.Equal(t, value, concrete.Value.content)
	assert.Equal(t, comment[1:], concrete.Comment.content)
	assert.Equal(t, comment[0], concrete.Comment.symbol)
}

// Test a value can be inside a quoted string
func TestQuotedValue(t *testing.T) {
	whitespace := fuzzWhiteSpace(10)
	key := fuzzKey()
	value := fuzzValue(true)
	comment := fuzzComment()
	file, err := testParse(key, whitespace, B_EQUALS, value, comment, B_NEWLINE)

	assert.NoError(t, err)
	assert.NotNil(t, file)

	line := file.Head

	assert.NotNil(t, line)

	concrete, ok := line.(*KeyValueLine)
	require.True(t, ok)

	assert.NotNil(t, concrete)

	assert.NotNil(t, concrete.Padding)
	assert.NotNil(t, concrete.Key)
	assert.NotNil(t, concrete.PostKeyPad)
	assert.NotNil(t, concrete.Value)
	assert.NotNil(t, concrete.Comment)

	assert.Len(t, concrete.Padding.content, 0)
	assert.Equal(t, key, concrete.Key.content)
	assert.Equal(t, whitespace, concrete.PostKeyPad.content)
	assert.Equal(t, value, concrete.Value.content)
	assert.Equal(t, comment[1:], concrete.Comment.content)
	assert.Equal(t, comment[0], concrete.Comment.symbol)
}

// Test an invalid key byte in a key returns an error
func TestInvalidKeyByteInKeyIsError(t *testing.T) {
	key := fuzzKey()
	invalidByte := fuzzyChoice(invalidKeyByteSet)
	value := fuzzValue(false)
	file, err := testParse(key, invalidByte, B_EQUALS, value, B_NEWLINE)

	assert.Error(t, err)
	assert.Nil(t, file)
	assert.ErrorContains(t, err, fmt.Sprintf("invalid character %02x in key", invalidByte))
}

// Test a non-whitespace byte after a key returns an error
func TestNonWhiteSpaceAfterKeyRaisesError(t *testing.T) {
	key := fuzzKey()
	whiteSpace := fuzzWhiteSpace(10)
	invalidByte := fuzzyChoice(invalidWhitespaceByteset)
	value := fuzzValue(false)
	file, err := testParse(key, whiteSpace, invalidByte, B_EQUALS, value, B_NEWLINE)

	assert.Error(t, err)
	assert.Nil(t, file)
	assert.ErrorContains(t, err, fmt.Sprintf("invalid non-whitespace character %02x in key", invalidByte))
}

// Test a quote inside an unquoted string is illegal
func TestExtraValueQuoteErrorInUnquotedString(t *testing.T) {
	key := fuzzKey()
	value := fuzzValue(false) // value without quotes
	file, err := testParse(key, B_EQUALS, value, B_QUOTE, B_NEWLINE)

	assert.Error(t, err)
	assert.Nil(t, file)
	assert.ErrorContains(t, err, "illegal quote character in value")
}

// Test a quote after a terminated quoted string is illegal
func TestExtraValueQuoteErrorAfterTerminatedQuoted(t *testing.T) {
	key := fuzzKey()
	value := fuzzValue(true) // value with quotes
	file, err := testParse(key, B_EQUALS, value, B_QUOTE, B_NEWLINE)

	assert.Error(t, err)
	assert.Nil(t, file)
	assert.ErrorContains(t, err, "illegal quote character in value")
}

// Test a non-whitespace character after a terminated quoted string is illegal
func TestNonWhitespaceAfterTerminatedValueIllegal(t *testing.T) {
	key := fuzzKey()
	value := fuzzValue(true) // value with quotes
	illegalByte := fuzzyChoice(invalidWhitespaceByteset)
	file, err := testParse(key, B_EQUALS, value, illegalByte, B_NEWLINE)

	assert.Error(t, err)
	assert.Nil(t, file)
	assert.ErrorContains(t, err, fmt.Sprintf("illegal character %02x after terminated quoted value", illegalByte))
}

// Test a Null Byte inside an unquoted value is illegal
func TestNullByteInUnquotedValueIllegal(t *testing.T) {
	key := fuzzKey()
	value := fuzzValue(false) // value without quotes
	illegalByte := B_NULL
	file, err := testParse(key, B_EQUALS, value, illegalByte, B_NEWLINE)

	assert.Error(t, err)
	assert.Nil(t, file)
	assert.ErrorContains(t, err, fmt.Sprintf("illegal character %02x in value", illegalByte))
}

// Test a Null Byte inside a quoted value is illegal
func TestNullByteInRuotedValueIllegal(t *testing.T) {
	key := fuzzKey()
	value := fuzzValue(false) // value without quotes, we add them manually
	illegalByte := B_NULL
	file, err := testParse(key, B_EQUALS, B_QUOTE, value, illegalByte, B_QUOTE, B_NEWLINE)

	assert.Error(t, err)
	assert.Nil(t, file)
	assert.ErrorContains(t, err, fmt.Sprintf("illegal character %02x in value", illegalByte))
}
