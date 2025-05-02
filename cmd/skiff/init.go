/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	_ "embed"
	"fmt"
	"path/filepath"

	"github.com/nyambati/skiff/internal/config"
	"github.com/nyambati/skiff/internal/utils"
	"github.com/spf13/cobra"
)

//go:embed templates/terragrunt.default.tmpl
var terragruntDefaultTemplate []byte

//go:embed templates/catalog.yaml
var serviceTypesTemplate []byte

//go:embed templates/skiff.yaml
var skiffConfigTemplate []byte

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [path] [flags]",
	Short: "initializes a new skiff project",
	Long:  "creates the required folder structure for a skiff project.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return initProject(".", force)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func initProject(path string, force bool) error {
	config := []struct {
		Folder   string
		File     string
		Template []byte
	}{
		{
			Folder:   "manifests",
			File:     config.CatalogFile,
			Template: serviceTypesTemplate,
		},
		{
			Folder:   "templates",
			File:     config.TerragruntTemplateFile,
			Template: terragruntDefaultTemplate,
		},
		{
			Folder:   ".",
			File:     config.SkiffConfigFile,
			Template: skiffConfigTemplate,
		},
	}

	for _, c := range config {

		if err := utils.CreateDirectory(filepath.Join(path, c.Folder)); err != nil {
			return err
		}

		templatePath := filepath.Join(path, c.Folder, c.File)
		if utils.FileExists(templatePath) && !force {
			fmt.Printf("⚠️ Skipping, %s already exists, use --force to overwrite\n", templatePath)
			continue
		}

		if err := utils.WriteFile(filepath.Join(path, c.Folder, c.File), c.Template); err != nil {
			return err
		}
	}

	return nil
}
