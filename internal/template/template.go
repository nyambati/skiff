package template

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/nyambati/skiff/internal/catalog"
	"github.com/nyambati/skiff/internal/config"
	"github.com/nyambati/skiff/internal/manifest"
	"github.com/nyambati/skiff/internal/strategy"
	"github.com/sirupsen/logrus"
)

// getRenderConfig retrieves the render configuration based on the provided strategy name,
// account ID, and labels. It reads the service catalog from the catalog file,
// loads the account manifests, and applies the selected strategy to generate the render
// configuration. Returns a pointer to the RenderConfig and an error if any issues occur
// during processing.

func GetRenderConfig(ctx context.Context, manifestID, labels string) (*strategy.RenderConfig, error) {
	var catalog catalog.Catalog
	cfg, err := config.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	serviceTypesPath := fmt.Sprintf("%s/%s", cfg.Manifests, config.CatalogFile)
	if err := catalog.Read(serviceTypesPath); err != nil {
		return nil, err
	}
	manifests, err := loadManifests(ctx, manifestID)
	if err != nil {
		return nil, err
	}

	return strategy.Execute(ctx, manifests, &catalog, labels), nil
}

// loadManifests reads the account manifests from the manifests folder based on the provided
// account ID or IDs. If an empty string is provided, it reads all account manifests in the
// folder. It returns a slice of pointers to Manifest and an error if any issues occur during
// processing.
func loadManifests(ctx context.Context, manifestID string) ([]*manifest.Manifest, error) {
	var manifests []*manifest.Manifest

	cfg, err := config.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	accounts, err := getManifestIdetifiers(manifestID, cfg.Manifests)
	if err != nil {
		return nil, err
	}

	for _, accountID := range accounts {
		accountID = strings.TrimSuffix(accountID, filepath.Ext(accountID))

		m, err := manifest.Read(ctx, accountID)
		if err != nil {
			return nil, err
		}
		if err := m.Resolve(ctx); err != nil {
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
func getManifestIdetifiers(manifestName, manifestPath string) ([]string, error) {
	if manifestName != "" {
		return []string{manifestName}, nil
	}

	manifestDir, err := os.ReadDir(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifests directory: %w", err)
	}

	var manifestIDs []string
	for _, entry := range manifestDir {
		if isValidIdentifier(entry) {
			manifestIDs = append(manifestIDs, entry.Name())
		}
	}

	if len(manifestIDs) == 0 {
		return nil, fmt.Errorf("no account manifests found in %s", manifestPath)
	}

	return manifestIDs, nil
}

func isValidIdentifier(entry os.DirEntry) bool {
	return !entry.IsDir() &&
		entry.Name() != config.CatalogFile &&
		strings.HasSuffix(entry.Name(), ".yaml")
}

// Render generates Terragrunt configuration files based on the provided strategy,
// account ID, and labels. It retrieves the rendering configuration and parses the
// specified templates. If dryRun is true, it only prints the rendered output without
// writing to files. Otherwise, it creates the necessary directories and writes the
// rendered files to the specified target folders. Returns an error if any issues occur
// during the rendering process.

func Render(ctx context.Context, manifestID, labels string, dryRun bool) error {
	configs, err := GetRenderConfig(ctx, manifestID, labels)
	if err != nil {
		return err
	}

	for _, cfg := range *configs {
		funcMaps := sprig.TxtFuncMap()
		funcMaps["terraform_atrributes"] = func() string {
			terraform, ok := (*cfg.Context)["terraform"].(map[string]interface{})
			if !ok {
				panic("terraform is not a map[string]interface{}")
			}
			return RenderTerraformAttrs(terraform)

		}

		funcMaps["service_config"] = func() string {
			body, ok := (*cfg.Context)["body"].(map[string]interface{})
			if !ok {
				panic("terraform is not a map[string]interface{}")
			}
			return RenderToHCL(body)

		}
		// funcMaps["var"] = func() types.TemplateContext { return *cfg.Context }
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
			logrus.
				Infof("🧪 [Dry Run] Would render: %s\n", outputPath)
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
