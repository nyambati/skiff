/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [manifest| catalog] [flags]",
	Short: "Add a new account, service, or service type",
	Args:  cobra.MinimumNArgs(0),
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.AddCommand(addAccountCmd)
	addCmd.AddCommand(addCatalogCmd)
}
