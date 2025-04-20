/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	_ "embed"
	"path/filepath"

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
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		quiet, _ := cmd.Flags().GetBool("quiet")
		force, _ := cmd.Flags().GetBool("force")

		basePath := "."
		if len(args) >= 1 {
			basePath = args[0]
		}

		cfg := []Config{
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

		for _, c := range cfg {
			if err := utils.CreateDirectory(basePath, c.Folder, quiet); err != nil {
				return err
			}

			if err := utils.WriteFile(filepath.Join(basePath, c.Folder, c.TemplateName), c.Template, quiet, force); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
