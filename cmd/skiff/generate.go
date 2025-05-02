/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/nyambati/skiff/internal/template"
	"github.com/spf13/cobra"
)

var labels string
var dryRun bool

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generates terragrunt configurations files from manifests",
	Run: func(cmd *cobra.Command, args []string) {
		if err := template.Render(cmd.Context(), name, labels, dryRun); err != nil {
			cmd.PrintErr(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringVar(&name, "name", "", "name of the manifest used to generate terraform configurations")
	generateCmd.Flags().StringVarP(&labels, "labels", "l", "", "labels to filter terraform configurations to apply to the list of accounts")
	generateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "dry run, generate")
}
