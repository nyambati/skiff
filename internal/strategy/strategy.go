package strategy

import (
	"context"
	"path/filepath"

	"github.com/nyambati/skiff/internal/catalog"
	"github.com/nyambati/skiff/internal/config"
	"github.com/nyambati/skiff/internal/manifest"
	"github.com/nyambati/skiff/internal/utils"
)

var defaultTemplate = "terragrunt.default.tmpl"

// Execute applies the rendering strategy to the provided manifests and service catalog,
// and returns a pointer to a RenderConfig slice. The function takes the following
// arguments:
//
//   - manifests: a slice of pointers to Manifest, representing the account manifests
//     to be processed
//   - catalog: a pointer to Manifest, representing the service catalog
//   - labels: a comma-separated string of key-value pairs, used to filter the services
//     to be processed
//
// The function iterates over the provided manifests and their services, and applies
// the following steps:
//
//   - For each service, it checks if the service has the specified labels. If not,
//     it skips the service
//   - For each service, it retrieves the service type definition from the service catalog
//     and builds a strategy context
//   - For each service, it evaluates the target path using the strategy context and
//     the path template specified in the strategy configuration
//   - For each service, it sets the target folder to the evaluated target path, and
//     sets the template path to the default template if the service type definition
//     does not specify a template
//   - For each service, it appends a new Config to the renderConfigs slice, with the
//     template path, target folder, and service data
//
// The function returns a pointer to the renderConfigs slice.
func Execute(ctx context.Context, manifests []*manifest.Manifest, catalog *catalog.Catalog, labels string) *RenderConfig {
	cfg, ok := ctx.Value("config").(*config.Config)
	if !ok {
		return nil
	}
	renderConfigs := make(RenderConfig, 0, len(manifests))
	for _, m := range manifests {
		for _, svc := range m.Services {
			if labels != "" && !utils.HasLabels(svc.Labels, utils.ParseKeyValueFlag(labels)) {
				continue
			}

			templatePath := svc.ResolvedType.Template
			if templatePath == "" {
				templatePath = defaultTemplate
			}

			templatePath = filepath.Join(cfg.Templates, templatePath)

			renderConfigs = append(renderConfigs, Config{
				Template:     templatePath,
				TargetFolder: utils.SanitizePath(filepath.Join(cfg.Terragrunt, svc.ResolvedTargetPath)),
				Context:      &svc.TemplateContext,
			})
		}
	}
	return &renderConfigs
}
