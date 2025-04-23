/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/nyambati/skiff/internal/template"
	"github.com/spf13/cobra"
)

var labels string
var dryRun bool

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates terragrunt configurations files from manifests",
	Run: func(cmd *cobra.Command, args []string) {
		if err := template.Render("default", accountID, labels, dryRun); err != nil {
			cmd.PrintErr(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringVar(&accountID, "account-id", "", "account id to generate terraform configurations for")
	generateCmd.Flags().StringVarP(&labels, "labels", "l", "", "labels to filter terraform configurations to apply to the list of accounts")
	generateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "dry run, generate")
}
