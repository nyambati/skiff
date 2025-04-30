package strategy

import (
	"path/filepath"
	"testing"

	"github.com/nyambati/skiff/internal/account"
	"github.com/nyambati/skiff/internal/config"
	"github.com/nyambati/skiff/internal/service"
	"github.com/nyambati/skiff/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	// Setup mock config
	config.Config = &config.SkiffConfig{
		Path: config.Path{
			Templates:  "mock/templates",
			Terragrunt: "mock/terragrunt",
		},
	}

	testCases := []struct {
		name           string
		manifests      []*account.Manifest
		catalog        *service.Manifest
		labels         string
		expectedConfig *RenderConfig
	}{
		{
			name: "Single service with no labels",
			manifests: []*account.Manifest{{
				Services: map[string]service.Service{
					"test-service": {
						ResolvedType: &service.ServiceType{
							Template: "",
						},
						ResolvedTargetPath: "test/path",
						TemplateContext: types.TemplateContext{
							"name": "test-service",
						},
					},
				},
			}},
			catalog: &service.Manifest{},
			labels:  "",
			expectedConfig: &RenderConfig{{
				Template:     filepath.Join(config.Config.Path.Templates, defaultTemplate),
				TargetFolder: filepath.Join(config.Config.Path.Terragrunt, "test/path"),
				Context: &types.TemplateContext{
					"name": "test-service",
				},
			}},
		},
		{
			name: "Service with custom template and matching labels",
			manifests: []*account.Manifest{{
				Services: map[string]service.Service{
					"custom-service": {
						Labels: map[string]any{
							"env": "dev",
						},
						ResolvedType: &service.ServiceType{
							Template: "custom.tmpl",
						},
						ResolvedTargetPath: "custom/path",
						TemplateContext: types.TemplateContext{
							"name": "custom-service",
						},
					},
				},
			}},
			catalog: &service.Manifest{},
			labels:  "env=dev",
			expectedConfig: &RenderConfig{{
				Template:     filepath.Join(config.Config.Path.Templates, "custom.tmpl"),
				TargetFolder: filepath.Join(config.Config.Path.Terragrunt, "custom/path"),
				Context: &types.TemplateContext{
					"name": "custom-service",
				},
			}},
		},
		{
			name: "Service with non-matching labels",
			manifests: []*account.Manifest{{
				Services: map[string]service.Service{
					"prod-service": {
						Labels: map[string]any{
							"env": "prod",
						},
						ResolvedType: &service.ServiceType{
							Template: "prod.tmpl",
						},
						ResolvedTargetPath: "prod/path",
						TemplateContext: types.TemplateContext{
							"name": "prod-service",
						},
					},
				},
			}},
			catalog:        &service.Manifest{},
			labels:         "env=dev",
			expectedConfig: &RenderConfig{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Execute(tc.manifests, tc.catalog, tc.labels)
			assert.Equal(t, tc.expectedConfig, result)
		})
	}
}
