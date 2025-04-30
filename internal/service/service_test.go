package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nyambati/skiff/internal/config"
	"github.com/nyambati/skiff/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestConfig(t *testing.T) string {
	// Create a temporary directory
	tempDir := t.TempDir()
	config.Config = &config.SkiffConfig{
		Path: config.Path{
			Manifests: tempDir,
		},
		Strategy: config.Strategy{
			Template: "{{var.account_id}}/regions/{{var.region}}/{{var.service}}",
		},
	}
	return tempDir
}

func createServiceTypesFile(t *testing.T, tempDir string, content string) {
	serviceTypesPath := filepath.Join(tempDir, "service-types.yaml")
	err := os.WriteFile(serviceTypesPath, []byte(content), 0644)
	require.NoError(t, err)
}

func TestManifestMethods(t *testing.T) {
	t.Run("New Manifest", func(t *testing.T) {
		manifest := New()
		assert.NotNil(t, manifest)
		assert.Equal(t, "v1", manifest.APIVersion)
		assert.NotNil(t, manifest.Types)
		assert.Len(t, manifest.Types, 0)
	})

	t.Run("Add Service Type", func(t *testing.T) {
		manifest := New()
		svcType := &ServiceType{
			Template: "web.tmpl",
			Version:  "1.0.0",
			Group:    "default",
		}

		err := manifest.AddServiceType("web", svcType)
		require.NoError(t, err)

		retrievedType, exists := manifest.GetServiceType("web")
		assert.True(t, exists)
		assert.Equal(t, svcType.Template, retrievedType.Template)
		assert.Equal(t, svcType.Version, retrievedType.Version)
		assert.Equal(t, svcType.Group, retrievedType.Group)
	})

	t.Run("Merge Existing Service Type", func(t *testing.T) {
		manifest := New()
		svcType1 := &ServiceType{
			Template: "web.tmpl",
			Version:  "1.0.0",
			Group:    "default",
		}

		svcType2 := &ServiceType{
			Template: "web-updated.tmpl",
			Version:  "1.1.0",
		}

		err := manifest.AddServiceType("web", svcType1)
		require.NoError(t, err)

		err = manifest.AddServiceType("web", svcType2)
		require.NoError(t, err)

		retrievedType, exists := manifest.GetServiceType("web")
		assert.True(t, exists)
		assert.Equal(t, "web-updated.tmpl", retrievedType.Template)
		assert.Equal(t, "1.1.0", retrievedType.Version)
		assert.Equal(t, "default", retrievedType.Group)
	})

	t.Run("Add Service Type with Invalid Input", func(t *testing.T) {
		manifest := New()
		svcType := &ServiceType{}

		err := manifest.AddServiceType("web", svcType)
		assert.NoError(t, err, "Empty service type should be allowed")

		retrievedType, exists := manifest.GetServiceType("web")
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

		resolvedService, err := service.ResolveType(tempDir)
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

		_, err := service.ResolveType(tempDir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "service type database does not exist")
	})

	t.Run("Resolve Service Type Without Type", func(t *testing.T) {
		tempDir := setupTestConfig(t)
		service := &Service{}

		_, err := service.ResolveType(tempDir)
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

		reconciled := service.Reconcile("123456", metadata)

		assert.Equal(t, "1.0.0", reconciled.Version)
		assert.Equal(t, "123456", reconciled.Inputs["account_id"])
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

		reconciled := service.Reconcile("123456", metadata)

		assert.Equal(t, "engineering", reconciled.Labels["team"])
		assert.Equal(t, "production", reconciled.Labels["environment"])
	})
}

func TestBuildTemplateContext(t *testing.T) {
	t.Run("Build Template Context", func(t *testing.T) {
		service := &Service{
			Type:   "web",
			Region: "us-west-2",
			Scope:  "global",
			ResolvedType: &ServiceType{
				Template: "web.tmpl",
				Group:    "default",
				Version:  "1.0.0",
			},
			Inputs: map[string]any{
				"key1": "value1",
			},
			Dependencies: []Dependency{
				{"service": "dep1"},
				{"service": "dep2"},
			},
		}

		err := service.BuildTemplateContext("web-service", "123456", "TestAccount")
		require.NoError(t, err)

		assert.NotNil(t, service.TemplateContext)
		assert.Equal(t, "123456", service.TemplateContext["account_id"])
		assert.Equal(t, "TestAccount", service.TemplateContext["account_name"])
		assert.Equal(t, "web-service", service.TemplateContext["service"])
		assert.Equal(t, "global", service.TemplateContext["scope"])
		assert.Equal(t, "us-west-2", service.TemplateContext["region"])
		assert.Equal(t, "1.0.0", service.TemplateContext["version"])
		assert.Equal(t, "web.tmpl", service.TemplateContext["template"])
		assert.Equal(t, "default", service.TemplateContext["group"])
		assert.Equal(t, map[string]any{"key1": "value1"}, service.TemplateContext["inputs"])
		assert.Equal(t, []Dependency{
			{"service": "dep1"},
			{"service": "dep2"},
		}, service.TemplateContext["dependencies"])
	})

	t.Run("Build Template Context with Incomplete Resolved Type", func(t *testing.T) {
		service := &Service{
			Type:   "web",
			Region: "us-west-2",
			ResolvedType: &ServiceType{},
		}

		err := service.BuildTemplateContext("web-service", "123456", "TestAccount")
		require.NoError(t, err)

		assert.NotNil(t, service.TemplateContext)
		assert.Equal(t, "123456", service.TemplateContext["account_id"])
		assert.Equal(t, "TestAccount", service.TemplateContext["account_name"])
		assert.Equal(t, "web-service", service.TemplateContext["service"])
		assert.Equal(t, "us-west-2", service.TemplateContext["region"])
		assert.Equal(t, "", service.TemplateContext["template"])
		assert.Equal(t, "", service.TemplateContext["group"])
	})

	t.Run("Build Template Context with Complex Inputs", func(t *testing.T) {
		service := &Service{
			Type:   "web",
			Region: "us-west-2",
			ResolvedType: &ServiceType{
				Template: "web.tmpl",
				Group:    "default",
			},
			Inputs: map[string]any{
				"nested": map[string]any{
					"key": "value",
				},
				"list": []string{"item1", "item2"},
			},
		}

		err := service.BuildTemplateContext("web-service", "123456", "TestAccount")
		require.NoError(t, err)

		assert.NotNil(t, service.TemplateContext)
		assert.Equal(t, map[string]any{
			"nested": map[string]any{"key": "value"},
			"list": []string{"item1", "item2"},
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
			"environment": "production",
		}

		err := service.ResolveTargetPath("web-service", "123456", "TestAccount", metadata)
		require.NoError(t, err)

		assert.Equal(t, "123456/regions/us-west-2/web-service", service.ResolvedTargetPath)
	})
}
