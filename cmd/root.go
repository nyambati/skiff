/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"strings"

	"github.com/nyambati/skiff/internal/config"
	"github.com/spf13/cobra"
)

var verbose bool
var force bool

var rootCmd = &cobra.Command{
	Use:   "skiff",
	Short: "A tool to generate and apply Terragrunt configurations from YAML manifests",
	Long: `Skiff is a CLI tool that helps you define, generate,
and apply infrastructure using a declarative YAML format and Terragrunt.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if err := config.Load(cmd.Name()); err != nil {
			if strings.Contains(err.Error(), "Not Found") {
				cmd.Println("❌ Missing .skiff file. Run `skiff init` to create one ")
				os.Exit(1)
			}
			cmd.PrintErr(err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose mode")
	rootCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "force overwrite")
}
