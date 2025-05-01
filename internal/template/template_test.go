package template

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nyambati/skiff/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var skiffConfig *config.Config

func setupTestManifestsDir(t *testing.T) string {
	// Create a temporary directory for test manifests
	tempDir, err := os.MkdirTemp("", "skiff")
	require.NoError(t, err)

	// Ensure config is initialized
	if skiffConfig == nil {
		skiffConfig = &config.Config{}
	}

	// Backup and modify the config
	originalManifestsDir := skiffConfig.Manifests
	t.Cleanup(func() {
		skiffConfig.Manifests = originalManifestsDir
		os.RemoveAll(tempDir)
	})
	skiffConfig.Manifests = tempDir

	return tempDir
}

func TestGetManifestIDs(t *testing.T) {
	tempDir := setupTestManifestsDir(t)

	// Create some test manifest files
	testFiles := []string{
		"account1.yaml",
		"account2.yaml",
		"catalog.yaml", // should be ignored
	}

	for _, filename := range testFiles {
		fullPath := filepath.Join(tempDir, filename)
		err := os.WriteFile(fullPath, []byte("test content"), 0644)
		require.NoError(t, err)
	}

	// Test with empty accountID (should return all non-service-types files)
	accountIDs, err := getManifestIdetifiers("", tempDir)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"account1.yaml", "account2.yaml"}, accountIDs)

	// Test with specific accountID
	accountIDs, err = getManifestIdetifiers("account1.yaml", tempDir)
	require.NoError(t, err)
	assert.Equal(t, []string{"account1.yaml"}, accountIDs)
}

func TestToInputs(t *testing.T) {
	testCases := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "Simple map",
			input:    map[string]string{"key": "value"},
			expected: "inputs = \"map[key:value]\"\n",
		},
		{
			name:     "Nested map",
			input:    map[string]interface{}{"nested": map[string]string{"key": "value"}},
			expected: "inputs = {\n  nested = \"map[key:value]\"\n}\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := toInputs(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestToProp(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "Simple map",
			input:    map[string]string{"key": "value"},
			expected: "",
		},
		{
			name:     "Multiple keys sorted",
			input:    map[string]string{"b": "2", "a": "1"},
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := toProp(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestToObject(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "Simple string",
			input:    "test",
			expected: "\"test\"",
		},
		{
			name:     "Simple map",
			input:    map[string]string{"key": "value"},
			expected: "\"map[key:value]\"",
		},
		{
			name:     "List of strings",
			input:    []string{"a", "b", "c"},
			expected: "\"[a b c]\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := toObject(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRenderWithIndent(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		level    int
		expected string
	}{
		{
			name:     "Simple string",
			input:    "test",
			level:    1,
			expected: "\"test\"",
		},
		{
			name:     "Nested map",
			input:    map[string]string{"key": "value"},
			level:    1,
			expected: "\"map[key:value]\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := renderWithIndent(tc.input, tc.level)
			assert.Equal(t, tc.expected, result)
		})
	}
}
