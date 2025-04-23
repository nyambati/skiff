package strategy

import (
	"path/filepath"

	"github.com/nyambati/skiff/internal/account"
	"github.com/nyambati/skiff/internal/config"
	"github.com/nyambati/skiff/internal/service"
)

var defaultTemplate = "terragrunt.default.tmpl"

func Variation4(manifests []*account.Manifest, catalog *service.Manifest) *RenderConfig {
	renderConfigs := make(RenderConfig, 0, len(manifests))
	for _, manifest := range manifests {
		accountID := manifest.Account.ID
		rootFolder := filepath.Join(config.Config.Path.Terragrunt, accountID)
		for name, svc := range manifest.Services {
			typeDef, ok := catalog.Types[svc.Type]
			if !ok {
				continue
			}
			svc.Version = ResolveVersion(svc, typeDef)
			mergeInputs(&svc, manifest)
			var folder string
			if svc.Scope == "global" {
				folder = filepath.Join(rootFolder, "global", name)
			} else {
				folder = filepath.Join(rootFolder, "regions", svc.Region, typeDef.Folder, name)
			}
			templatePath := typeDef.Template
			if templatePath == "" {
				templatePath = defaultTemplate
			}
			templatePath = filepath.Join(config.Config.Templates, templatePath)
			renderConfigs = append(renderConfigs, Config{
				Template:     templatePath,
				TargetFolder: folder,
				Data:         TemplateData{Service: svc, Source: typeDef.Source},
			})
		}
	}
	return &renderConfigs
}

func ResolveVersion(svc service.Service, st service.ServiceType) string {
	if svc.Version != "" {
		return svc.Version
	}
	return st.Version
}

func mergeInputs(service *service.Service, manifest *account.Manifest) {
	mergedTags := make(map[string]any)
	for key, value := range manifest.Metadata {
		mergedTags[key] = value
	}
	for key, value := range service.Metadata {
		mergedTags[key] = value
	}
	service.Inputs["account_id"] = manifest.Account.ID
	service.Inputs["region"] = service.Region
	service.Inputs["tags"] = mergedTags
}
