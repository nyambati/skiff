package catalog

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/nyambati/skiff/internal/config"
	"github.com/nyambati/skiff/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var skiffConfig *config.Config

func setupTestConfig(t *testing.T) string {
	// Create a temporary directory
	tempDir := t.TempDir()
	skiffConfig = &config.Config{
		Path: config.Path{
			Manifests: tempDir,
		},
		Strategy: config.Strategy{
			Template: "{{var.id}}/regions/{{var.region}}/{{var.service}}",
		},
	}
	return tempDir
}

func createServiceTypesFile(t *testing.T, tempDir string, content string) {
	serviceTypesPath := filepath.Join(tempDir, config.CatalogFile)
	err := os.WriteFile(serviceTypesPath, []byte(content), 0644)
	require.NoError(t, err)
}

func TestManifestMethods(t *testing.T) {
	t.Run("New Manifest", func(t *testing.T) {
		catalog := NewCatalog()
		assert.NotNil(t, catalog)
		assert.Equal(t, "v1", catalog.APIVersion)
		assert.NotNil(t, catalog.Types)
		assert.Len(t, catalog.Types, 0)
	})

	t.Run("Add Service Type", func(t *testing.T) {
		catalog := NewCatalog()
		svcType := &ServiceType{
			Template: "web.tmpl",
			Version:  "1.0.0",
			Group:    "default",
		}

		err := catalog.AddServiceType("web", svcType, true)
		require.NoError(t, err)

		retrievedType, exists := catalog.GetServiceType("web")
		assert.True(t, exists)
		assert.Equal(t, svcType.Template, retrievedType.Template)
		assert.Equal(t, svcType.Version, retrievedType.Version)
		assert.Equal(t, svcType.Group, retrievedType.Group)
	})

	t.Run("Merge Existing Service Type", func(t *testing.T) {
		catalog := NewCatalog()
		svcType1 := &ServiceType{
			Template: "web.tmpl",
			Version:  "1.0.0",
			Group:    "default",
		}

		svcType2 := &ServiceType{
			Template: "web-updated.tmpl",
			Version:  "1.1.0",
		}

		err := catalog.AddServiceType("web", svcType1, true)
		require.NoError(t, err)

		err = catalog.AddServiceType("web", svcType2, true)
		require.NoError(t, err)

		retrievedType, exists := catalog.GetServiceType("web")
		assert.True(t, exists)
		assert.Equal(t, "web-updated.tmpl", retrievedType.Template)
		assert.Equal(t, "1.1.0", retrievedType.Version)
		assert.Equal(t, "default", retrievedType.Group)
	})

	t.Run("Add Service Type with Invalid Input", func(t *testing.T) {
		catalog := NewCatalog()
		svcType := &ServiceType{}

		err := catalog.AddServiceType("web", svcType, true)
		assert.NoError(t, err, "Empty service type should be allowed")

		retrievedType, exists := catalog.GetServiceType("web")
		assert.True(t, exists)
		assert.Equal(t, "", retrievedType.Template)
		assert.Equal(t, "", retrievedType.Version)
		assert.Equal(t, "", retrievedType.Group)
	})
}

func TestServiceResolveType(t *testing.T) {
	t.Run("Resolve Existing Service Type", func(t *testing.T) {
		tempDir := setupTestConfig(t)
		createServiceTypesFile(t, tempDir, `
apiVersion: v1
types:
  web:
    template: web.tmpl
    version: 1.0.0
    group: default
`)

		service := &Service{
			Type: "web",
		}

		ctx := context.WithValue(context.Background(), "config", skiffConfig)

		resolvedService, err := service.ResolveType(ctx)
		require.NoError(t, err)
		assert.NotNil(t, resolvedService.ResolvedType)
		assert.Equal(t, "web.tmpl", resolvedService.ResolvedType.Template)
		assert.Equal(t, "1.0.0", resolvedService.ResolvedType.Version)
	})

	t.Run("Resolve Non-Existing Service Type", func(t *testing.T) {
		tempDir := setupTestConfig(t)
		createServiceTypesFile(t, tempDir, `
apiVersion: v1
types:
  web:
    template: web.tmpl
`)

		service := &Service{
			Type: "database",
		}
		ctx := context.WithValue(context.Background(), "config", skiffConfig)
		_, err := service.ResolveType(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "service type database does not exist")
	})

	t.Run("Resolve Service Type Without Type", func(t *testing.T) {
		// tempDir := setupTestConfig(t)
		service := &Service{}
		ctx := context.WithValue(context.Background(), "config", skiffConfig)
		_, err := service.ResolveType(ctx)
		require.Error(t, err)
		assert.Equal(t, "service type is required", err.Error())
	})
}

func TestServiceReconcile(t *testing.T) {
	t.Run("Reconcile Service", func(t *testing.T) {
		service := &Service{
			ResolvedType: &ServiceType{
				Version: "1.0.0",
			},
			Region: "us-west-2",
		}

		metadata := types.Metadata{
			"environment": "production",
		}

		reconciled := service.Reconcile(metadata)

		assert.Equal(t, "1.0.0", reconciled.Version)
		assert.Equal(t, "us-west-2", reconciled.Inputs["region"])
		assert.Equal(t, "production", reconciled.Labels["environment"])
	})

	t.Run("Reconcile Service with Existing Labels", func(t *testing.T) {
		service := &Service{
			ResolvedType: &ServiceType{
				Version: "1.0.0",
			},
			Region: "us-west-2",
			Labels: map[string]any{
				"team": "engineering",
			},
		}

		metadata := types.Metadata{
			"environment": "production",
		}

		reconciled := service.Reconcile(metadata)

		assert.Equal(t, "engineering", reconciled.Labels["team"])
		assert.Equal(t, "production", reconciled.Labels["environment"])
	})
}

func TestBuildTemplateContext(t *testing.T) {
	metadata := types.Metadata{
		"id":   "123456",
		"name": "TestAccount",
	}

	t.Run("Build Template Context", func(t *testing.T) {
		service := &Service{
			Type:   "web",
			Region: "us-west-2",
			Scope:  "global",
			ResolvedType: &ServiceType{
				Group:   "default",
				Version: "1.0.0",
			},
			Inputs: map[string]any{
				"key1": "value1",
			},
			Dependencies: []Dependency{
				{"service": "dep1"},
				{"service": "dep2"},
			},
		}

		err := service.BuildTemplateContext("web-service", metadata)
		require.NoError(t, err)

		assert.NotNil(t, service.TemplateContext)
		assert.Equal(t, "123456", service.TemplateContext["id"])
		assert.Equal(t, "TestAccount", service.TemplateContext["name"])
		assert.Equal(t, "web-service", service.TemplateContext["service"])
		assert.Equal(t, "web", service.TemplateContext["type"])
		assert.Equal(t, "global", service.TemplateContext["scope"])
		assert.Equal(t, "us-west-2", service.TemplateContext["region"])
		assert.Equal(t, "1.0.0", service.TemplateContext["version"])
		assert.Equal(t, "default", service.TemplateContext["group"])
		assert.Equal(t, map[string]any{"key1": "value1"}, service.TemplateContext["inputs"])
		assert.Equal(t, []Dependency{
			{"service": "dep1"},
			{"service": "dep2"},
		}, service.TemplateContext["dependencies"])
		assert.NotNil(t, service.TemplateContext["terraform"])
		assert.NotNil(t, service.TemplateContext["body"])
	})

	t.Run("Build Template Context with Incomplete Resolved Type", func(t *testing.T) {
		service := &Service{
			Type:         "web",
			Region:       "us-west-2",
			ResolvedType: &ServiceType{},
		}

		err := service.BuildTemplateContext("web-service", metadata)
		require.NoError(t, err)

		assert.NotNil(t, service.TemplateContext)
		assert.Equal(t, "web-service", service.TemplateContext["service"])
		assert.Equal(t, "web", service.TemplateContext["type"])
		assert.Equal(t, "us-west-2", service.TemplateContext["region"])
		assert.Equal(t, "", service.TemplateContext["group"])
		assert.Equal(t, "", service.TemplateContext["version"])
		assert.Nil(t, service.TemplateContext["inputs"])
		assert.Nil(t, service.TemplateContext["dependencies"])
	})

	t.Run("Build Template Context with Complex Inputs", func(t *testing.T) {
		service := &Service{
			Type:   "web",
			Region: "us-west-2",
			ResolvedType: &ServiceType{
				Group: "default",
			},
			Inputs: map[string]any{
				"nested": map[string]any{
					"key": "value",
				},
				"list": []string{"item1", "item2"},
			},
		}

		err := service.BuildTemplateContext("web-service", metadata)
		require.NoError(t, err)

		assert.NotNil(t, service.TemplateContext)
		assert.Equal(t, map[string]any{
			"nested": map[string]any{"key": "value"},
			"list":   []string{"item1", "item2"},
		}, service.TemplateContext["inputs"])
	})
}

func TestResolveTargetPath(t *testing.T) {
	t.Run("Resolve Target Path", func(t *testing.T) {
		setupTestConfig(t)

		service := &Service{
			Type:   "web",
			Region: "us-west-2",
			ResolvedType: &ServiceType{
				Group: "default",
			},
		}

		metadata := types.Metadata{
			"id":          "123456",
			"environment": "production",
		}

		ctx := context.WithValue(context.Background(), "config", skiffConfig)

		err := service.ResolveTargetPath(ctx, "web-service", metadata)
		require.NoError(t, err)

		assert.Equal(t, "123456/regions/us-west-2/web-service", service.ResolvedTargetPath)
	})
}
