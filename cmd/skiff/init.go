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
	Short: "initializes a new skiff project",
	Long: `
Generates skiff folder structrue with default manifests.

Creates:
	manifest > stores manifest and catalog files
	templates > stores default and self defined service templates
	terragrunt > the output folder for generated hcl files
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return config.InitProject(".", force)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
