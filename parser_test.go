package montoya

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: Fuzz non-whitespace strings (comments, section names, etc.)
// TODO: Test errors

// testParse runs Parse with the input string, constructing a reader from it
func testParse(input string) (*IniFile, error) {
	reader := strings.NewReader(input)
	return Parse(reader)
}

// fuzzWhiteSpace creates a Fuzzy bytestring of whitespace for testing
func fuzzWhiteSpace(numChars int) []byte {
	var whitespace []byte
	choices := []byte{
		0x20, // space
		0x09, // tab
		0x0D, // carriage return
	}
	for range numChars {
		choice := choices[rand.Intn(3)] 
		whitespace = append(whitespace, choice)
	}
	return whitespace
}

// Test an empty file yields no lines
func TestParseEmptyFile(t *testing.T) {
		file, err := testParse("")

		assert.NoError(t, err)
		assert.NotNil(t, file)
		
		line := file.Head

		assert.Nil(t, line)
}

////////////////////////////////////////////////////////////////////////////////
// EMPTY LINE CASES ////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// Test only a newline yields an empty line without nodes
func TestParseEmptyLine(t *testing.T) {
		file, err := testParse("\n")

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
		file, err := testParse(string(padding) + "\n")

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
	for _, symbol := range []byte{'#',';'} {
		t.Run("symbol +" + string(symbol), func(t *testing.T) {
			file, err := testParse(string(symbol) + " This is a	comment!\n")

			assert.NoError(t, err)
			assert.NotNil(t, file)
			
			line := file.Head
			fmt.Printf("test line %p\n", line)

			assert.NotNil(t, line)
			
			concrete, ok := line.(*EmptyLine)
			require.True(t, ok)

			assert.NotNil(t, concrete)
			
			whiteSpaceResult := concrete.Padding
			commentResult := concrete.Comment

			assert.NotNil(t, whiteSpaceResult)
			assert.NotNil(t, commentResult)

			assert.Equal(t, symbol, commentResult.symbol)
			assert.Equal(t, []byte(" This is a	comment!"), commentResult.content)
		})
	}
}

// Test a comment with some whitespace yields the whitespace and the comment node 
func TestParseEmptyLineCommentAndPadding(t *testing.T) {
	for _, symbol := range []byte{'#',';'} {
		t.Run("symbol +" + string(symbol), func(t *testing.T) {
			padding := fuzzWhiteSpace(20)
			file, err := testParse(string(padding) + string(symbol) + " This is a	comment!\n")

			assert.NoError(t, err)
			assert.NotNil(t, file)
			
			line := file.Head
			fmt.Printf("test line %p\n", line)

			assert.NotNil(t, line)
			
			concrete, ok := line.(*EmptyLine)
			require.True(t, ok)

			assert.NotNil(t, concrete)
			
			whiteSpaceResult := concrete.Padding
			commentResult := concrete.Comment

			assert.NotNil(t, whiteSpaceResult)
			assert.NotNil(t, commentResult)

			assert.Equal(t, padding, whiteSpaceResult.content)

			assert.Equal(t, symbol, commentResult.symbol)
			assert.Equal(t, []byte(" This is a	comment!"), commentResult.content)
		})
	}
}

// Test a double comment still only yields one comment 
func TestParseEmptyLineOnlyASingleComment(t *testing.T) {
	for _, symbol := range []byte{'#',';'} {
		strSymbol := string(symbol)
		t.Run("symbol +" + strSymbol, func(t *testing.T) {
			expected :=" This is a	" + strSymbol + "comment!" 
			file, err := testParse(strSymbol + expected + "\n")

			assert.NoError(t, err)
			assert.NotNil(t, file)
			
			line := file.Head
			fmt.Printf("test line %p\n", line)

			assert.NotNil(t, line)
			
			concrete, ok := line.(*EmptyLine)
			require.True(t, ok)

			assert.NotNil(t, concrete)
			
			whiteSpaceResult := concrete.Padding
			commentResult := concrete.Comment

			assert.NotNil(t, whiteSpaceResult)
			assert.NotNil(t, commentResult)

			assert.Equal(t, symbol, commentResult.symbol)
			assert.Equal(t, []byte(expected), commentResult.content)
		})
	}
}

////////////////////////////////////////////////////////////////////////////////
// SECTION LINE CASES //////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func TestParseSimpleSectionNoPadding(t *testing.T) {
	file, err := testParse("[foo]\n")

	assert.NoError(t, err)
	assert.NotNil(t, file)
	
	line := file.Head

	assert.NotNil(t, line)
	
	concrete, ok := line.(*SectionHeaderLine)
	require.True(t, ok)

	assert.NotNil(t, concrete)
	
	assert.NotNil(t, concrete.Padding)
	assert.NotNil(t, concrete.Header)
	assert.Nil(t, concrete.PostPad)
	assert.Nil(t, concrete.Comment)

	paddingResult := concrete.Padding
	sectionResult := concrete.Header

	assert.Len(t, paddingResult.content, 0)
	assert.Equal(t, []byte("foo"), sectionResult.content)
}


func TestParseSimpleSectionWithPadding(t *testing.T) {
	padding := fuzzWhiteSpace(20) 

	file, err := testParse(string(padding) + "[foo]\n")

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

	paddingResult := concrete.Padding
	sectionResult := concrete.Header
	postPadResult := concrete.PostPad

	assert.Len(t, postPadResult.content, 0)
	assert.Equal(t, padding, paddingResult.content)
	assert.Equal(t, []byte("foo"), sectionResult.content)
}

func TestParseSimpleSectionWithPaddingAndPostPad(t *testing.T) {
	padding := fuzzWhiteSpace(10)
	padding2 := fuzzWhiteSpace(10)

	file, err := testParse(string(padding) + "[foo]" + string(padding2) + "\n")

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

	paddingResult := concrete.Padding
	sectionResult := concrete.Header
	paddingResult2 := concrete.PostPad

	assert.Equal(t, padding, paddingResult.content)
	assert.Equal(t, padding2, paddingResult2.content)
	assert.Equal(t, []byte("foo"), sectionResult.content)
}

func TestParseSimpleSectionWithWhitespaceAndComment(t *testing.T) {
	for _, symbol := range []byte{'#',';'} {
		strSymbol := string(symbol)
		t.Run("symbol +" + strSymbol, func(t *testing.T) {
			padding := fuzzWhiteSpace(10)
			padding2 := fuzzWhiteSpace(10)

			file, err := testParse(string(padding) + "[foo]" + string(padding2) + strSymbol + "test Comment;# !\n")

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

			paddingResult := concrete.Padding
			sectionResult := concrete.Header
			paddingResult2 := concrete.PostPad
			commentResult:= concrete.Comment

			assert.Equal(t, padding, paddingResult.content)
			assert.Equal(t, padding2, paddingResult2.content)
			assert.Equal(t, []byte("foo"), sectionResult.content)
			assert.Equal(t, []byte("test Comment;# !"), commentResult.content)
		})
	}
}