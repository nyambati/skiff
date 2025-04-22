/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	_ "embed"
	"fmt"
	"path/filepath"

	"github.com/nyambati/skiff/internal/service"
	"github.com/nyambati/skiff/internal/utils"
	"github.com/spf13/cobra"
)

//go:embed templates/terragrunt.default.tmpl
var terragruntDefaultTemplate []byte

//go:embed templates/service-types.yaml
var serviceTypesTemplate []byte

type Config struct {
	Folder       string
	Template     []byte
	TemplateName string
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [path] [flags]",
	Short: "Initialize a new Skiff project",
	Long: `Creates the required folder structure for a Skiff project, including:
- manifests/ (with an empty service-types.yaml)
- templates/ (with a default terragrunt.default.tmpl)
`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return initSkiff(basePath, verbose, force)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&basePath, "path", "p", ".", "path to base output directory,default to ./")
}

func initSkiff(path string, verbose bool, force bool) error {
	manifest := service.New()

	serviceTypesTemplate, err := manifest.ToYAML()
	if err != nil {
		return err
	}

	config := []Config{
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
	}

	for _, c := range config {
		if err := utils.CreateDirectory(filepath.Join(path, c.Folder)); err != nil {
			return err
		}

		templatePath := filepath.Join(basePath, c.Folder, c.TemplateName)
		if utils.FileExists(templatePath) && !force {
			fmt.Printf("⚠️ Skipping, %s already exists, use --force to overwrite\n", templatePath)
			continue
		}
		if err := utils.WriteFile(filepath.Join(basePath, c.Folder, c.TemplateName), c.Template); err != nil {
			return err
		}
	}
	return nil
}
