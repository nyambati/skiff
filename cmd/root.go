/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "skiff",
	Short: "A tool to generate and apply Terragrunt configurations from YAML manifests",
	Long: `Skiff is a CLI tool that helps you define, generate,
and apply infrastructure using a declarative YAML format and Terragrunt.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
