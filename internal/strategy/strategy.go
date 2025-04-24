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

			svc.Version = ResolveVersion(svc, typeDef)
			mergeInputs(&svc, manifest)

			context := buildStrategyContext(svcName, typeDef.Folder, svc, manifest)
			evaluatedPath, err := evaluateTargetPath(config.Config.Strategy.Template, context)
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
				TargetFolder: SanitizePath(filepath.Join(config.Config.Terragrunt, evaluatedPath)),
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

func buildStrategyContext(
	svcName string,
	svcTypeFolder string,
	svc service.Service,
	manifest *account.Manifest,
) StrategyContext {
	context := StrategyContext{
		"service":      svcName,
		"region":       svc.Region,
		"type":         svc.Type,
		"account_id":   manifest.Account.ID,
		"account_name": manifest.Account.Name,
		"folder":       svcTypeFolder,
	}

	for key, value := range manifest.Metadata {
		context[key] = value
	}

	for key, value := range svc.Metadata {
		context[key] = value
	}

	return context
}

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
