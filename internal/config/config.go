package config

import (
	"context"
	_ "embed"
	"path/filepath"

	skiff "github.com/nyambati/skiff/internal/errors"
	"github.com/nyambati/skiff/internal/utils"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func New(command string) (*Config, error) {
	config := new(Config)

	if command == "init" {
		return config, nil
	}

	viper.AddConfigPath(".")
	viper.SetConfigName(".skiff")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, skiff.NewConfigurationError("failed to unmarshall config")
	}

	return config, nil
}

//go:embed templates/terragrunt.default.tmpl
var terragruntDefaultTemplate []byte

//go:embed templates/catalog.yaml
var serviceTypesTemplate []byte

//go:embed templates/skiff.yaml
var skiffConfigTemplate []byte

func InitProject(path string, force bool) error {
	config := []struct {
		Folder   string
		File     string
		Template []byte
	}{
		{
			Folder:   "manifests",
			File:     CatalogFile,
			Template: serviceTypesTemplate,
		},
		{
			Folder:   "templates",
			File:     TerragruntTemplateFile,
			Template: terragruntDefaultTemplate,
		},
		{
			Folder:   ".",
			File:     SkiffConfigFile,
			Template: skiffConfigTemplate,
		},
	}

	for _, c := range config {

		if err := utils.CreateDirectory(filepath.Join(path, c.Folder)); err != nil {
			return err
		}
		templatePath := filepath.Join(path, c.Folder, c.File)
		if c.Folder == "." {
			templatePath = filepath.Join(c.Folder, c.File)
		}

		if utils.FileExists(templatePath) && !force {
			logrus.Printf("skipping, %s already exists, use --force to overwrite\n", templatePath)
			continue
		}

		if err := utils.WriteFile(filepath.Join(path, c.Folder, c.File), c.Template); err != nil {
			return err
		}
	}

	return nil
}

func FromContext(ctx context.Context) (*Config, error) {
	config, ok := ctx.Value(ContextKey).(*Config)
	if !ok {
		return nil, skiff.NewConfigNotFoundError()
	}
	return config, nil
}
