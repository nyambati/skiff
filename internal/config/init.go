package config

import (
	_ "embed"
	"fmt"
	"path/filepath"

	"github.com/nyambati/skiff/internal/service"
	"github.com/nyambati/skiff/internal/utils"
)

//go:embed templates/terragrunt.default.tmpl
var terragruntDefaultTemplate []byte

//go:embed templates/service-types.yaml
var serviceTypesTemplate []byte

//go:embed templates/skiff.yaml
var skiffConfigTemplate []byte

func Init(path string, verbose bool, force bool) error {
	manifest := service.New()

	serviceTypesTemplate, err := manifest.ToYAML()
	if err != nil {
		return err
	}

	config := []InitConfig{
		{
			Folder:       "manifests",
			TemplateName: "service-types.yaml",
			Template:     serviceTypesTemplate,
		},
		{
			Folder:       "templates",
			TemplateName: "terragrunt.default.tmpl",
			Template:     terragruntDefaultTemplate,
		},
		{
			Folder:       ".",
			TemplateName: ".skiff",
			Template:     skiffConfigTemplate,
		},
	}

	for _, c := range config {
		if err := utils.CreateDirectory(filepath.Join(path, c.Folder)); err != nil {
			return err
		}

		templatePath := filepath.Join(path, c.Folder, c.TemplateName)
		if utils.FileExists(templatePath) && !force {
			fmt.Printf("⚠️ Skipping, %s already exists, use --force to overwrite\n", templatePath)
			continue
		}
		if err := utils.WriteFile(filepath.Join(path, c.Folder, c.TemplateName), c.Template); err != nil {
			return err
		}
	}
	return nil
}
