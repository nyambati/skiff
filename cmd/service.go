/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/nyambati/skiff/internal/account"
	"github.com/nyambati/skiff/internal/config"
	"github.com/nyambati/skiff/internal/service"
	"github.com/nyambati/skiff/internal/utils"
	"github.com/spf13/cobra"
)

var serviceName string
var serviceType string
var scope string
var region string
var version string
var inputs string

// serviceCmd represents the service command
var addServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Adds service manifests",
	Run: func(cmd *cobra.Command, args []string) {
		var manifest account.Manifest
		var catalog service.Manifest

		path := config.Config.Manifests
		if err := manifest.Read(accountID); err != nil {
			utils.PrintErrorAndExit(err)
		}

		if err := catalog.Read(fmt.Sprintf("%s/service-types.yaml", config.Config.Manifests)); err != nil {
			utils.PrintErrorAndExit(err)
		}

		if _, exists := catalog.GetServiceType(serviceType); !exists {
			err := fmt.Errorf("service type %s does not exist, run `skiff add service-type` to add a new service type", serviceType)
			utils.PrintErrorAndExit(err)
		}

		manifest.AddService(
			serviceName,
			&service.Service{
				Type:    serviceType,
				Scope:   scope,
				Region:  region,
				Version: version,
				Labels:  utils.ParseKeyValueFlag(labels),
				Inputs:  utils.ParseKeyValueFlag(inputs),
			},
		)

		if err := manifest.Write(path, verbose, true); err != nil {
			utils.PrintErrorAndExit(err)
		}
		fmt.Printf("✅ Service %s has been added to %s\n", serviceName, filepath.Join(path, accountID))
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
	addServiceCmd.Flags().StringVar(&labels, "labels", "", "service labels/tags")
	addServiceCmd.Flags().StringVar(&inputs, "inputs", "", "inputs")

	requiredFlags := []string{"account-id", "name", "type"}

	for _, flag := range requiredFlags {
		addServiceCmd.MarkFlagRequired(flag)
	}

}
