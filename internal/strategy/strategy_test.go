package strategy

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/nyambati/skiff/internal/catalog"
	"github.com/nyambati/skiff/internal/config"
	"github.com/nyambati/skiff/internal/manifest"
	"github.com/nyambati/skiff/internal/types"
	"github.com/stretchr/testify/assert"
)

var skiffConfig *config.Config

func TestExecute(t *testing.T) {
	// Setup mock config
	skiffConfig = &config.Config{
		Path: config.Path{
			Templates:  "mock/templates",
			Terragrunt: "mock/terragrunt",
		},
	}

	testCases := []struct {
		name           string
		manifests      []*manifest.Manifest
		catalog        *catalog.Catalog
		labels         string
		expectedConfig *RenderConfig
	}{
		{
			name: "Single service with no labels",
			manifests: []*manifest.Manifest{{
				Services: map[string]catalog.Service{
					"test-service": {
						ResolvedType: &catalog.ServiceType{
							Template: "",
						},
						ResolvedTargetPath: "test/path",
						TemplateContext: types.TemplateContext{
							"name": "test-service",
						},
					},
				},
			}},
			catalog: &catalog.Catalog{},
			labels:  "",
			expectedConfig: &RenderConfig{{
				Template:     filepath.Join(skiffConfig.Path.Templates, defaultTemplate),
				TargetFolder: filepath.Join(skiffConfig.Path.Terragrunt, "test/path"),
				Context: &types.TemplateContext{
					"name": "test-service",
				},
			}},
		},
		{
			name: "Service with custom template and matching labels",
			manifests: []*manifest.Manifest{{
				Services: map[string]catalog.Service{
					"custom-service": {
						Labels: map[string]any{
							"env": "dev",
						},
						ResolvedType: &catalog.ServiceType{
							Template: "custom.tmpl",
						},
						ResolvedTargetPath: "custom/path",
						TemplateContext: types.TemplateContext{
							"name": "custom-service",
						},
					},
				},
			}},
			catalog: &catalog.Catalog{},
			labels:  "env=dev",
			expectedConfig: &RenderConfig{{
				Template:     filepath.Join(skiffConfig.Path.Templates, "custom.tmpl"),
				TargetFolder: filepath.Join(skiffConfig.Path.Terragrunt, "custom/path"),
				Context: &types.TemplateContext{
					"name": "custom-service",
				},
			}},
		},
		{
			name: "Service with non-matching labels",
			manifests: []*manifest.Manifest{{
				Services: map[string]catalog.Service{
					"prod-service": {
						Labels: map[string]any{
							"env": "prod",
						},
						ResolvedType: &catalog.ServiceType{
							Template: "prod.tmpl",
						},
						ResolvedTargetPath: "prod/path",
						TemplateContext: types.TemplateContext{
							"name": "prod-service",
						},
					},
				},
			}},
			catalog:        &catalog.Catalog{},
			labels:         "env=dev",
			expectedConfig: &RenderConfig{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "config", skiffConfig)
			result := Execute(ctx, tc.manifests, tc.catalog, tc.labels)
			assert.Equal(t, tc.expectedConfig, result)
		})
	}
}
