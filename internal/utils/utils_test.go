package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteFile(t *testing.T) {
	t.Run("Write File Successfully", func(t *testing.T) {
		// Create a temporary directory
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "test.txt")

		// Write content to the file
		content := []byte("Hello, World!")
		err := WriteFile(filePath, content)
		require.NoError(t, err)

		// Verify file contents
		readContent, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, content, readContent)
	})

	t.Run("Write File to Non-Existent Directory", func(t *testing.T) {
		// Create a temporary directory
		tempDir := t.TempDir()
		nonExistentDir := filepath.Join(tempDir, "non-existent", "dir")
		filePath := filepath.Join(nonExistentDir, "test.txt")

		// Ensure directory is not created
		_, err := os.Stat(nonExistentDir)
		require.True(t, os.IsNotExist(err))

		// Write content to the file
		content := []byte("Hello, World!")
		err = os.MkdirAll(filepath.Dir(filePath), 0755)
		require.NoError(t, err)
		err = WriteFile(filePath, content)
		require.NoError(t, err)

		// Verify file contents
		readContent, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, content, readContent)
	})
}

func TestParseKeyValueFlag(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected map[string]any
	}{
		{
			name:  "Simple Key-Value Pairs",
			input: "key1=value1,key2=value2",
			expected: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name:  "Pairs with Whitespace",
			input: " key1 = value1 , key2 = value2 ",
			expected: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name:     "Empty Input",
			input:    "",
			expected: map[string]any{},
		},
		{
			name:     "Invalid Pairs",
			input:    "key1,=value2,key3=",
			expected: map[string]any{},
		},
		{
			name:  "Pairs with Multiple Equal Signs",
			input: "key1=value1=extra,key2=value2",
			expected: map[string]any{
				"key1": "value1=extra",
				"key2": "value2",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseKeyValueFlag(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestToMap(t *testing.T) {
	t.Run("Convert Struct to Map", func(t *testing.T) {
		type TestStruct struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}

		input := TestStruct{
			Name:  "test",
			Value: 42,
		}

		result, err := ToMap[TestStruct](input)
		require.NoError(t, err)

		assert.Equal(t, "test", result["name"])
		assert.Equal(t, float64(42), result["value"])
	})

	t.Run("Convert Primitive Struct", func(t *testing.T) {
		type Primitive struct {
			Value string `json:""`
		}

		input := Primitive{Value: "test"}
		result, err := ToMap[Primitive](input)
		require.NoError(t, err)

		assert.Equal(t, map[string]any{"Value": "test"}, result)
	})
}
