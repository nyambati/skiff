/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"strings"

	"github.com/nyambati/skiff/internal/terragrunt"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [plan,apply,destroy]",
	Short: "runs terragrunt command for specified manifest or services",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := terragrunt.Run(cmd.Context(), args[0], flagManifestID, flagLabels, strings.Split(flagArgs, ","), flagDryRun); err != nil {
			cmd.PrintErr(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&flagLabels, "labels", "l", "", "labels to filter terraform configurations to apply to the list of accounts")
	runCmd.Flags().StringVarP(&flagArgs, "args", "a", "", "additional arguments to pass to terragrunt")
	runCmd.Flags().BoolVarP(&flagDryRun, "dry-run", "d", false, "dry run mode")
}
