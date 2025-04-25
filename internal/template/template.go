package template

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/nyambati/skiff/internal/account"
	"github.com/nyambati/skiff/internal/config"
	"github.com/nyambati/skiff/internal/service"
	"github.com/nyambati/skiff/internal/strategy"
	"github.com/nyambati/skiff/internal/types"
)

// getRenderConfig retrieves the render configuration based on the provided strategy name,
// account ID, and labels. It reads the service catalog from the service-types.yaml file,
// loads the account manifests, and applies the selected strategy to generate the render
// configuration. Returns a pointer to the RenderConfig and an error if any issues occur
// during processing.

func GetRenderConfig(accountID, labels string) (*strategy.RenderConfig, error) {
	var serviceCatalog service.Manifest
	serviceTypesPath := fmt.Sprintf("%s/service-types.yaml", config.Config.Manifests)
	if err := serviceCatalog.Read(serviceTypesPath); err != nil {
		return nil, err
	}
	manifests, err := loadManifests(accountID)
	if err != nil {
		return nil, err
	}

	return strategy.Execute(manifests, &serviceCatalog, labels), nil
}

// loadManifests reads the account manifests from the manifests folder based on the provided
// account ID or IDs. If an empty string is provided, it reads all account manifests in the
// folder. It returns a slice of pointers to Manifest and an error if any issues occur during
// processing.
func loadManifests(accountID string) ([]*account.Manifest, error) {
	var manifests []*account.Manifest

	accounts, err := getAccountIDs(accountID)
	if err != nil {
		return nil, err
	}

	for _, accountID := range accounts {
		accountID = strings.TrimSuffix(accountID, filepath.Ext(accountID))
		m := new(account.Manifest)
		err := m.Read(config.Config.Manifests, accountID)
		if err != nil {
			return nil, err
		}
		manifests = append(manifests, m)
	}
	return manifests, nil
}

// getAccountIDs reads the account manifest IDs from the manifests folder based on the
// provided account ID. If an empty string is provided, it reads all account manifest IDs
// in the folder. It returns a slice of strings containing the account IDs and an error
// if any issues occur during processing.
func getAccountIDs(accountID string) ([]string, error) {
	if accountID != "" {
		return []string{accountID}, nil
	}

	manifestDir, err := os.ReadDir(config.Config.Manifests)
	if err != nil {
		return nil, err
	}

	var accountIDs []string
	for _, entry := range manifestDir {
		if !entry.IsDir() && !strings.Contains(entry.Name(), "service-types") {
			accountIDs = append(accountIDs, entry.Name())
		}
	}
	return accountIDs, nil
}

// Render generates Terragrunt configuration files based on the provided strategy,
// account ID, and labels. It retrieves the rendering configuration and parses the
// specified templates. If dryRun is true, it only prints the rendered output without
// writing to files. Otherwise, it creates the necessary directories and writes the
// rendered files to the specified target folders. Returns an error if any issues occur
// during the rendering process.

func Render(accountID, labels string, dryRun bool) error {
	configs, err := GetRenderConfig(accountID, labels)
	if err != nil {
		return err
	}

	for _, cfg := range *configs {
		funcMaps := sprig.TxtFuncMap()
		funcMaps["toObject"] = toObject
		funcMaps["toProp"] = toProp
		funcMaps["var"] = func() types.TemplateContext { return *cfg.Context }
		tmpl, err := template.New("").Funcs(funcMaps).ParseFiles(cfg.Template)
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}

		outputPath := filepath.Join(cfg.TargetFolder, "terragrunt.hcl")

		if dryRun {
			var buff bytes.Buffer
			if err := tmpl.ExecuteTemplate(&buff, filepath.Base(cfg.Template), nil); err != nil {
				return fmt.Errorf("failed to render template to %s: %w", outputPath, err)
			}
			fmt.Printf("🧪 [Dry Run] Would render: %s\n", outputPath)
			fmt.Println(buff.String())
			continue
		}

		// Ensure target folder exists
		if err := os.MkdirAll(cfg.TargetFolder, 0755); err != nil {
			return fmt.Errorf("failed to create folder %s: %w", cfg.TargetFolder, err)
		}

		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", outputPath, err)
		}
		defer file.Close()
		if err := tmpl.ExecuteTemplate(file, filepath.Base(cfg.Template), nil); err != nil {
			return fmt.Errorf("failed to render template to %s: %w", outputPath, err)
		}
		fmt.Printf("✅ Rendered: %s\n", outputPath)
	}
	return nil
}

func toProp(v interface{}) string {
	normalized := normalizeYAMLTypes(v)

	val, ok := normalized.(map[string]interface{})
	if !ok {
		return ""
	}

	indent := func(l int) string {
		return strings.Repeat("  ", l)
	}

	var lines []string
	keys := make([]string, 0, len(val))
	for k := range val {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		line := fmt.Sprintf("%s%s = %s", indent(1), k, renderWithIndent(val[k], 2))
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// toObject takes an arbitrary Go value and renders it as an HCL string,
// indenting the output with two spaces. It's intended to be used as a
// template function to inject arbitrary data into HCL templates.
func toObject(v interface{}) string {
	return renderWithIndent(normalizeYAMLTypes(v), 1)
}

// renderWithIndent takes an arbitrary Go value and renders it as an HCL string,
// indenting the output with the given number of spaces. It's intended to be used
// as a template function to inject arbitrary data into HCL templates.
//
// The function handles the following types as follows:
//
// - map[string]interface{}: renders as a nested object with sorted keys
// - []interface{}: renders as a list of elements
// - string: renders as a quoted string
// - bool, int, float64: renders as the raw value
// - all other types: renders as a quoted string
func renderWithIndent(v interface{}, level int) string {
	indent := func(l int) string {
		return strings.Repeat("  ", l)
	}

	switch val := v.(type) {
	case map[string]interface{}:
		var out strings.Builder
		out.WriteString("{\n")
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys) // Consistent order
		for _, k := range keys {
			out.WriteString(fmt.Sprintf("%s%s = %s\n", indent(level), k, renderWithIndent(val[k], level+1)))
		}
		out.WriteString(indent(level-1) + "}")
		return out.String()

	case []interface{}:
		var out strings.Builder
		out.WriteString("[\n")
		for _, item := range val {
			out.WriteString(fmt.Sprintf("%s%s,\n", indent(level), renderWithIndent(item, level+1)))
		}
		out.WriteString(indent(level-1) + "]")
		return out.String()

	case string:
		return fmt.Sprintf("\"%s\"", val)
	case bool, int, float64:
		return fmt.Sprintf("%v", val)
	default:
		return fmt.Sprintf("\"%v\"", val)
	}
}

func normalizeYAMLTypes(input interface{}) interface{} {
	switch v := input.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{})
		for key, value := range v {
			m[fmt.Sprintf("%v", key)] = normalizeYAMLTypes(value)
		}
		return m
	case []interface{}:
		for i, val := range v {
			v[i] = normalizeYAMLTypes(val)
		}
		return v
	default:
		return v
	}
}
