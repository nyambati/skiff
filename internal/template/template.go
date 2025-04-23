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
)

func getRenderConfig(strategyName, accountID, labels string) (*strategy.RenderConfig, error) {
	var serviceCatalog service.Manifest
	serviceTypesPath := fmt.Sprintf("%s/service-types.yaml", config.Config.Manifests)
	if err := serviceCatalog.Read(serviceTypesPath); err != nil {
		return nil, err
	}
	manifests, err := loadManifests(accountID)
	if err != nil {
		return nil, err
	}
	selectedStrategy := strategy.GetStrategy(strategyName)
	return selectedStrategy(manifests, &serviceCatalog, labels), nil
}

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

func Render(strategy, accountID, labels string, dryRun bool) error {
	configs, err := getRenderConfig(strategy, accountID, labels)
	if err != nil {
		return err
	}

	for _, cfg := range *configs {

		tmpl, err := template.New("").Funcs(template.FuncMap{"toHCL": toHCL}).Funcs(sprig.TxtFuncMap()).ParseFiles(cfg.Template)
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}

		outputPath := filepath.Join(cfg.TargetFolder, "terragrunt.hcl")

		if dryRun {
			var buff bytes.Buffer
			if err := tmpl.ExecuteTemplate(&buff, filepath.Base(cfg.Template), cfg.Data); err != nil {
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
		if err := tmpl.ExecuteTemplate(file, filepath.Base(cfg.Template), cfg.Data); err != nil {
			return fmt.Errorf("failed to render template to %s: %w", outputPath, err)
		}
		fmt.Printf("✅ Rendered: %s\n", outputPath)
	}
	return nil
}

func toHCL(v interface{}) string {
	return renderWithIndent(v, 1)
}

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
