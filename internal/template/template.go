package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/nyambati/skiff/internal/account"
	"github.com/nyambati/skiff/internal/config"
	"github.com/nyambati/skiff/internal/service"
	"github.com/nyambati/skiff/internal/strategy"
)

func getRenderConfig(strtgy, accountID, filters string) (*strategy.RenderConfig, error) {
	var catalog service.Manifest

	if err := catalog.Read(fmt.Sprintf("%s/service-types.yaml", config.Config.Manifests)); err != nil {
		return nil, err
	}

	manifests, err := loadManifests(accountID, filters)
	if err != nil {
		return nil, err
	}

	s := strategy.GetStrategy(strtgy)

	return s(manifests, &catalog), nil

}

func loadManifests(accountID, filters string) ([]*account.Manifest, error) {
	var accounts []string
	var manifests []*account.Manifest

	if accountID != "" {
		accounts = []string{accountID}
	} else {
		dir, err := os.ReadDir(config.Config.Manifests)
		if err != nil {
			return nil, err
		}
		for _, f := range dir {
			if !f.IsDir() && !strings.Contains(f.Name(), "service-types") {
				accounts = append(accounts, f.Name())
			}
		}
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

func Render(strategy, accountID, labels string, dryRun bool) error {
	configs, err := getRenderConfig(strategy, accountID, labels)
	if err != nil {
		return err
	}

	for _, cfg := range *configs {
		// Ensure target folder exists
		if err := os.MkdirAll(cfg.TargetFolder, 0755); err != nil {
			return fmt.Errorf("failed to create folder %s: %w", cfg.TargetFolder, err)
		}
		outputPath := filepath.Join(cfg.TargetFolder, "terragrunt.hcl")
		tmpl, err := template.ParseFiles(cfg.Template)
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}

		if dryRun {
			fmt.Printf("🧪 [Dry Run] Would render: %s\n", outputPath)
			continue
		}

		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", outputPath, err)
		}
		defer file.Close()

		if err := tmpl.Execute(file, cfg.Data); err != nil {
			return fmt.Errorf("failed to render template to %s: %w", outputPath, err)
		}

		fmt.Printf("✅ Rendered: %s\n", outputPath)
	}
	return nil
}
