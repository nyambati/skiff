package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name          string
		command       string
		configContent string
		expectError   bool
	}{
		{
			name:    "Init command",
			command: "init",
		},
		{
			name:    "Valid config file",
			command: "test",
			configContent: `
version: v1
verbose: true
strategy:
  description: Default strategy
  template: default.tmpl
path:
  manifests: ./manifests
  templates: ./templates
  terragrunt: ./terragrunt
  strategies: ./strategies
`,
		},
		{
			name:        "Missing config file",
			command:     "test",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.configContent != "" {
				// Create a temporary config file
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, ".skiff")
				err := os.WriteFile(configPath, []byte(tc.configContent), 0644)
				require.NoError(t, err)

				// Change working directory
				oldWd, err := os.Getwd()
				require.NoError(t, err)
				defer os.Chdir(oldWd)

				err = os.Chdir(tempDir)
				require.NoError(t, err)

				// Reset viper
				viper.Reset()
			}

			config, err := New(tc.command)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tc.command != "init" {
					assert.NotNil(t, config)
					assert.Equal(t, "v1", config.Version)
					assert.True(t, config.Verbose)
					assert.Equal(t, "./manifests", config.Path.Manifests)
				}
			}
		})
	}
}

func TestInitProject(t *testing.T) {
	testCases := []struct {
		name         string
		force        bool
		preExisting  bool
		expectedFiles []string
	}{
		{
			name:  "Create new project",
			force: false,
			expectedFiles: []string{
				"manifests/catalog.yaml",
				"templates/terragrunt.default.tmpl",
				".skiff",
			},
		},
		{
			name:         "Force overwrite existing files",
			force:        true,
			preExisting:  true,
			expectedFiles: []string{
				"manifests/catalog.yaml",
				"templates/terragrunt.default.tmpl",
				".skiff",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary directory
			tempDir := t.TempDir()

			// Pre-create files if needed
			if tc.preExisting {
				for _, file := range tc.expectedFiles {
					fullPath := filepath.Join(tempDir, file)
					err := os.MkdirAll(filepath.Dir(fullPath), 0755)
					require.NoError(t, err)
					err = os.WriteFile(fullPath, []byte("existing content"), 0644)
					require.NoError(t, err)
				}
			}

			// Call InitProject
			err := InitProject(tempDir, tc.force)
			require.NoError(t, err)

			// Verify files exist
			for _, file := range tc.expectedFiles {
				fullPath := filepath.Join(tempDir, file)
				assert.FileExists(t, fullPath, "File %s should exist", file)
			}
		})
	}
}

func TestFromContext(t *testing.T) {
	testCases := []struct {
		name        string
		contextFunc func() context.Context
		expectError bool
	}{
		{
			name: "Config in context",
			contextFunc: func() context.Context {
				cfg := &Config{
					Version: "v1",
					Verbose: true,
				}
				return context.WithValue(context.Background(), ContextKey, cfg)
			},
		},
		{
			name: "No config in context",
			contextFunc: func() context.Context {
				return context.Background()
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.contextFunc()

			config, err := FromContext(ctx)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				assert.Equal(t, "v1", config.Version)
				assert.True(t, config.Verbose)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	// Test that critical constants are defined
	assert.NotEmpty(t, ToolName, "ToolName should not be empty")
	assert.NotEmpty(t, CatalogFile, "CatalogFile should not be empty")
	assert.NotEmpty(t, TerragruntTemplateFile, "TerragruntTemplateFile should not be empty")
	assert.NotEmpty(t, SkiffConfigFile, "SkiffConfigFile should not be empty")
}
