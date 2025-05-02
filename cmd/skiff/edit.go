/*
Copyright © 2025 nyambati thomasnyambati@gmail.com
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var editCmd = &cobra.Command{
	Use:   "edit [manifest| catalog] [flags]",
	Short: "edits manifest and catalog files",
	Args:  cobra.MinimumNArgs(0),
}

func init() {
	rootCmd.AddCommand(editCmd)
	editCmd.AddCommand(addAccountCmd)
	editCmd.AddCommand(addCatalogCmd)
}
