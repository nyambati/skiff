/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/nyambati/skiff/internal/manifest"
	"github.com/nyambati/skiff/internal/utils"
	"github.com/spf13/cobra"
)

// serviceCmd represents the service command
var addServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "edits specific service in manifest file",
	Long: `The service command allows you to edit a specific service in the manifest file.

Examples:
  skiff edit service --manifest my-manifest --service my-service
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := manifest.AddService(cmd.Context(), flagManifestID, flagServiceName); err != nil {
			utils.PrintErrorAndExit(err)
		}
	},
}

func init() {
	editCmd.AddCommand(addServiceCmd)
	addServiceCmd.Flags().StringVarP(&flagManifestID, "manifest", "m", "", "name of the manifest file")
	addServiceCmd.Flags().StringVarP(&flagServiceName, "service", "s", "", "name of the service in the manifest file")
	addServiceCmd.MarkFlagRequired("manifest")
	addServiceCmd.MarkFlagRequired("service")
}
