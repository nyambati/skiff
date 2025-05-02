/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/nyambati/skiff/internal/terragrunt"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
func terragruntCommands(commands []string) []*cobra.Command {
	cobraCommands := make([]*cobra.Command, 0, len(commands))
	for _, command := range commands {
		cobraCommands = append(cobraCommands, &cobra.Command{
			Use:   command,
			Short: "runs terragrunt " + command + " command for specified manifest or services",
			Run: func(cmd *cobra.Command, args []string) {
				if err := terragrunt.Run(cmd.Context(), command, name, labels, dryRun); err != nil {
					cmd.PrintErr(err)
					os.Exit(1)
				}
			},
		})
	}
	return cobraCommands
}

func init() {
	commands := terragruntCommands([]string{"plan", "apply", "destroy"})
	for _, cmd := range commands {
		rootCmd.AddCommand(cmd)
		cmd.Flags().StringVarP(&labels, "labels", "l", "", "labels to filter terraform configurations to apply to the list of accounts")
		cmd.Flags().BoolVar(&dryRun, "dry-run", false, "dry run mode")
	}
}
