/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	_ "embed"

	"github.com/nyambati/skiff/internal/config"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [path] [flags]",
	Short: "Initialize a new Skiff project",
	Long: `Creates the required folder structure for a Skiff project, including:
- manifests/ (with an empty service-types.yaml)
- templates/ (with a default terragrunt.default.tmpl)
`,

	RunE: func(cmd *cobra.Command, args []string) error {
		basePath := "."
		return config.Init(basePath, verbose, force)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
