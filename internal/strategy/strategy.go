package strategy

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/nyambati/skiff/internal/account"
	"github.com/nyambati/skiff/internal/config"
	"github.com/nyambati/skiff/internal/service"
	"github.com/nyambati/skiff/internal/utils"
)

var defaultTemplate = "terragrunt.default.tmpl"

func evaluateTargetPath(path string, ctx StrategyContext) (string, error) {
	tmpl, err := template.New("target_path").
		Funcs(sprig.FuncMap()).
		Funcs(template.FuncMap{"var": func() StrategyContext { return ctx }}).
		Parse(path)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return "", err
	}
	return validatePath(&buf)
}

func validatePath(buffer *bytes.Buffer) (string, error) {
	// Check if template markers still exist
	if bytes.Contains(buffer.Bytes(), []byte("{{")) || bytes.Contains(buffer.Bytes(), []byte("}}")) {
		return "", fmt.Errorf("template was not fully rendered: unresolved variables remain in %q", buffer.String())
	}
	return buffer.String(), nil
}

func buildTemplateContext(
	serviceName string,
	serviceType *service.ServiceType,
	svc *service.Service,
	manifest *account.Manifest,
) (*TemplateContext, error) {
	ctx := TemplateContext{
		"account_id":   manifest.Account.ID,
		"account_name": manifest.Account.Name,
		"service":      serviceName,
		"scope":        svc.Scope,
		"region":       svc.Region,
		"version":      svc.Version,
	}

	data, err := utils.ToMap(serviceType)
	if err != nil {
		return nil, err
	}

	for k, v := range data {
		ctx[strings.ToLower(k)] = v
	}

	data, err = utils.ToMap(svc)
	if err != nil {
		return nil, err
	}

	for k, v := range data {
		ctx[strings.ToLower(k)] = v
	}

	return &ctx, nil
}

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
func Execute(manifests []*account.Manifest, catalog *service.Manifest, labels string) *RenderConfig {
	renderConfigs := make(RenderConfig, 0, len(manifests))
	for _, manifest := range manifests {
		for svcName, svc := range manifest.Services {
			if labels != "" && !utils.HasLabels(svc.Labels, utils.ParseKeyValueFlag(labels)) {
				continue
			}

			typeDef, ok := catalog.Types[svc.Type]
			if !ok {
				continue
			}

			context := buildStrategyContext(svcName, &svc, &typeDef, manifest)

			path, err := evaluateTargetPath(config.Config.Strategy.Template, context)
			if err != nil {
				continue
			}

			ctx, err := buildTemplateContext(svcName, &typeDef, &svc, manifest)
			if err != nil {
				continue
			}

			templatePath := typeDef.Template
			if templatePath == "" {
				templatePath = defaultTemplate
			}
			templatePath = filepath.Join(config.Config.Templates, templatePath)

			renderConfigs = append(renderConfigs, Config{
				Template:     templatePath,
				TargetFolder: SanitizePath(filepath.Join(config.Config.Terragrunt, path)),
				Context:      ctx,
			})
		}
	}
	return &renderConfigs
}

// Reconcile the service with the service type and manifest.
//
// It sets the service version to the service type version if not specified.
// It initializes the service labels map if not present.
// It copies the manifest metadata to the service labels.
// It sets three special inputs: "account_id", "region", and "tags" (the service labels).
func reconcileService(
	service *service.Service,
	serviceType *service.ServiceType,
	manifest *account.Manifest,
) {
	if service.Version == "" {
		service.Version = serviceType.Version
	}

	if len(service.Labels) == 0 {
		service.Labels = map[string]any{}
	}

	for key, value := range manifest.Metadata {
		service.Labels[key] = value
	}

	service.Inputs["account_id"] = manifest.Account.ID
	service.Inputs["region"] = service.Region
	service.Inputs["tags"] = service.Labels
}

func buildStrategyContext(
	svcName string,
	service *service.Service,
	serviceType *service.ServiceType,
	manifest *account.Manifest,
) StrategyContext {
	// Reconcile service
	reconcileService(service, serviceType, manifest)

	context := StrategyContext{
		"service":      svcName,
		"region":       service.Region,
		"type":         service.Type,
		"account_id":   manifest.Account.ID,
		"account_name": manifest.Account.Name,
		"group":        serviceType.Group,
	}

	for key, value := range manifest.Metadata {
		context[key] = value
	}

	for key, value := range service.Labels {
		context[key] = value
	}

	return context
}

// SanitizePath normalizes and cleans a given path string by removing any
// extraneous whitespace, line endings, and empty fragments. It first
// normalizes line endings and splits the input into path parts, then
// trims whitespace and removes empty segments. The resulting clean
// path is returned as a slash-delimited string.

func SanitizePath(input string) string {
	// Normalize line endings and split into path parts
	lines := strings.FieldsFunc(input, func(r rune) bool {
		return r == '\n' || r == '\r' || r == '\t'
	})
	// Split again by "/" and trim all fragments
	var cleanParts []string
	for _, line := range lines {
		parts := strings.Split(line, "/")
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				cleanParts = append(cleanParts, trimmed)
			}
		}
	}
	// Rejoin into clean slash-delimited path
	return strings.Join(cleanParts, "/")
}
