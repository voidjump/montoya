package montoya

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testParse(input string) (*IniFile, error) {
	reader := strings.NewReader(input)
	return Parse(reader)
}

// Test an empty file yields no lines
func TestParseEmptyFile(t *testing.T) {
		file, err := testParse("")

		assert.NoError(t, err)
		assert.NotNil(t, file)
		
		line := file.Head

		assert.Nil(t, line)
}

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
		file, err := testParse(string([]byte{
			0x20, // space
			0x09, // tab
			0x0D, // carriage return
			0x0A, // line feed (newline)
		}))

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
		assert.Equal(t, []byte{0x20, 0x09, 0x0D}, whiteSpaceResult.content)
}

// Test only comment and a newline yields a comment node and empty whitespace node
func TestParseEmptyLineComment(t *testing.T) {
		file, err := testParse(string("# This is a	comment!\n"))

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

		assert.Equal(t, []byte(" This is a	comment!"), commentResult.content)
}
