/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/nyambati/skiff/internal/template"
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generates terragrunt configurations files from manifests",
	Long: `
Generates terragrunt configurations files from manifests.

Example:
  skiff generate --name my-manifest --labels env=prod,region=us-west-2`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := template.Render(cmd.Context(), flagManifestID, flagLabels, flagDryRun); err != nil {
			cmd.PrintErr(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringVarP(
		&flagManifestID, "manifest", "m", "", "name of the manifest used to generate terraform configurations",
	)
	generateCmd.Flags().StringVarP(
		&flagLabels, "labels", "l", "", "labels to filter terraform configurations to apply to the list of accounts",
	)
	generateCmd.Flags().BoolVarP(&flagDryRun, "dry-run", "d", false, "dry run, generate")
}
