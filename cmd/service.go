/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"path/filepath"

	"github.com/nyambati/skiff/internal/account"
	"github.com/nyambati/skiff/internal/service"
	"github.com/nyambati/skiff/internal/utils"
	"github.com/spf13/cobra"
)

var serviceName string
var serviceType string
var scope string
var region string
var version string
var metadata string
var inputs string

// serviceCmd represents the service command
var addServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Adds service manifests",
	RunE: func(cmd *cobra.Command, args []string) error {
		var manifest account.Manifest
		path := filepath.Join(basePath, "manifests")
		if err := manifest.Read(path, accountID); err != nil {
			return err
		}

		manifest.AddService(
			serviceName,
			&service.Service{
				Type:     serviceType,
				Scope:    scope,
				Region:   region,
				Version:  version,
				Metadata: utils.ParseKeyValueFlag(metadata),
				Inputs:   utils.ParseKeyValueFlag(inputs),
			},
		)
		return manifest.Write(path, verbose, force)
	},
}

func init() {
	addCmd.AddCommand(addServiceCmd)
	addServiceCmd.Flags().StringVar(&accountID, "account-id", "", "Account manifest name (required)")
	addServiceCmd.Flags().StringVar(&serviceName, "name", "", "Name of the service (required)")
	addServiceCmd.Flags().StringVar(&serviceType, "type", "", "Service type (required)")
	addServiceCmd.Flags().StringVar(&scope, "scope", "", "Scope: global or regional (required)")
	addServiceCmd.Flags().StringVar(&region, "region", "", "AWS region (required)")
	addServiceCmd.Flags().StringVar(&version, "version", "", "Optional module version override")
	addServiceCmd.Flags().StringVar(&metadata, "metadata", "", "metadata")
	addServiceCmd.Flags().StringVar(&inputs, "inputs", "", "inputs")

	requiredFlags := []string{"account-id", "name"}

	for _, flag := range requiredFlags {
		addServiceCmd.MarkFlagRequired(flag)
	}

}
