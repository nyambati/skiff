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
