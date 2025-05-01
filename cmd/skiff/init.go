/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	_ "embed"
	"fmt"
	"path/filepath"

	"github.com/nyambati/skiff/internal/utils"
	"github.com/spf13/cobra"
)

//go:embed templates/terragrunt.default.tmpl
var terragruntDefaultTemplate []byte

//go:embed templates/service-types.yaml
var serviceTypesTemplate []byte

//go:embed templates/skiff.yaml
var skiffConfigTemplate []byte

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [path] [flags]",
	Short: "Initialize a new Skiff project",
	Long: `Creates the required folder structure for a Skiff project, including:
- manifests/ (with an empty service-types.yaml)
- templates/ (with a default terragrunt.default.tmpl)
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return initProject(".", verbose, force)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func initProject(path string, verbose bool, force bool) error {
	config := []struct {
		Folder   string
		File     string
		Template []byte
	}{
		{
			Folder:   "manifests",
			File:     "service-types.yaml",
			Template: serviceTypesTemplate,
		},
		{
			Folder:   "templates",
			File:     "terragrunt.default.tmpl",
			Template: terragruntDefaultTemplate,
		},
		{
			Folder:   ".",
			File:     ".skiff",
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
