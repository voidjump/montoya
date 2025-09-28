package montoya

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: Test errors

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
