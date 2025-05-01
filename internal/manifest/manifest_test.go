package manifest

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/nyambati/skiff/internal/catalog"
	"github.com/nyambati/skiff/internal/config"
	"github.com/nyambati/skiff/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var skiffConfig *config.Config

func createTempManifestFile(t *testing.T, content string) string {
	// Create a temporary directory
	tempDir := t.TempDir()
	// Create a temporary manifest file
	manifestPath := filepath.Join(tempDir, "1234567890.yaml")
	err := os.WriteFile(manifestPath, []byte(content), 0644)
	require.NoError(t, err)

	skiffConfig = &config.Config{
		Path: config.Path{
			Manifests: tempDir,
		},
	}

	return "1234567890"
}

func TestManifestRead(t *testing.T) {
	testCases := []struct {
		name             string
		manifestContent  string
		expectedMetadata types.Metadata
		expectedServices map[string]catalog.Service
	}{
		{
			name: "Valid manifest with single service",
			manifestContent: `
apiVersion: v1
account:

metadata:
  name: Test Account
  id: "1234567890"
services:
  web-service:
    type: web
    region: us-east-1
    labels:
      env: production
`,
			expectedMetadata: types.Metadata{
				"name": "Test Account",
				"id":   "1234567890",
			},
			expectedServices: map[string]catalog.Service{
				"web-service": {
					Type:   "web",
					Region: "us-east-1",
					Labels: map[string]any{
						"env": "production",
					},
				},
			},
		},
		{
			name: "Manifest with multiple services",
			manifestContent: `
apiVersion: v1
metadata:
  name: Multi Service Account
  id: "0987654321"
services:
  db-service:
    type: database
    region: us-west-2
  api-service:
    type: api
    region: us-east-1
`,
			expectedMetadata: types.Metadata{
				"name": "Multi Service Account",
				"id":   "0987654321",
			},
			expectedServices: map[string]catalog.Service{
				"db-service": {
					Type:   "database",
					Region: "us-west-2",
				},
				"api-service": {
					Type:   "api",
					Region: "us-east-1",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary manifest file
			manifestName := createTempManifestFile(t, tc.manifestContent)

			ctx := context.WithValue(context.Background(), "config", skiffConfig)

			// Create a new Manifest and read the file

			manifest, err := Read(ctx, manifestName)
			require.NoError(t, err)

			// Verify metadata details
			assert.Equal(t, tc.expectedMetadata, manifest.Metadata)

			// Verify services
			require.Len(t, manifest.Services, len(tc.expectedServices))
			for serviceName, expectedService := range tc.expectedServices {
				service, exists := manifest.Services[serviceName]
				require.True(t, exists, "Service %s should exist", serviceName)

				assert.Equal(t, expectedService.Type, service.Type)
				assert.Equal(t, expectedService.Region, service.Region)
				assert.Equal(t, expectedService.Labels, service.Labels)
			}
		})
	}
}

func TestManifestResolve(t *testing.T) {
	t.Run("Resolve method", func(t *testing.T) {
		// Create a temporary directory
		tempDir := t.TempDir()
		skiffConfig = &config.Config{
			Path: config.Path{
				Manifests: tempDir,
			},
			Strategy: config.Strategy{
				Template: "default.tmpl",
			},
		}

		// Create service-types.yaml
		serviceTypesContent := `
apiVersion: v1
types:
  web:
    template: web.tmpl
    type: web
    version: latest
    group: default
`
		serviceTypesPath := filepath.Join(tempDir, "service-types.yaml")
		err := os.WriteFile(serviceTypesPath, []byte(serviceTypesContent), 0644)
		require.NoError(t, err)

		// Create a test manifest with services
		ctx := context.WithValue(context.Background(), "config", skiffConfig)

		m, err := Read(ctx, "1234567890")

		require.NoError(t, err)
		// Call Resolve method
		err = m.Resolve(ctx)

		// Verify no error occurs
		assert.NoError(t, err)
	})
}
