/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/nyambati/skiff/internal/catalog"
	"github.com/nyambati/skiff/internal/config"
	"github.com/nyambati/skiff/internal/manifest"
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
		var manifest manifest.Manifest
		var svcCatalog catalog.Catalog

		path := config.Config.Manifests
		if err := manifest.Read(accountID); err != nil {
			utils.PrintErrorAndExit(err)
		}

		if err := svcCatalog.Read(fmt.Sprintf("%s/service-types.yaml", config.Config.Manifests)); err != nil {
			utils.PrintErrorAndExit(err)
		}

		if _, exists := svcCatalog.GetServiceType(serviceType); !exists {
			err := fmt.Errorf("service type %s does not exist, run `skiff add service-type` to add a new service type", serviceType)
			utils.PrintErrorAndExit(err)
		}

		existingContent, err := utils.ToYAML(catalog.Service{
			Type:    serviceType,
			Scope:   catalog.ScopeRegional,
			Version: "edit me",
			Region:  "edit me",
			Inputs: map[string]any{
				"name": serviceName,
			},
			Labels: map[string]any{
				"type":  serviceType,
				"scope": scope,
				"name":  serviceName,
			},
			Dependencies: []catalog.Dependency{},
		})
		if err != nil {
			utils.PrintErrorAndExit(err)
		}

		editContent, err := utils.EditFile(fmt.Sprintf("%s/%s.yaml", path, accountID), existingContent)
		if err != nil {
			utils.PrintErrorAndExit(err)
		}

		svc, err := catalog.ServiceFromYAML(editContent)
		if err != nil {
			utils.PrintErrorAndExit(err)
		}

		manifest.AddService(serviceName, svc)

		if err := manifest.Write(path, verbose, true); err != nil {
			utils.PrintErrorAndExit(err)
		}
		fmt.Printf("✅ Service %s has been added to %s\n", serviceName, filepath.Join(path, accountID))
	},
}

func init() {
	addCmd.AddCommand(addServiceCmd)
	addServiceCmd.Flags().StringVar(&accountID, "account-id", "", "Account manifest name (required)")
	addServiceCmd.Flags().StringVar(&serviceName, "name", "", "Service manifest name (required)")
	addServiceCmd.Flags().StringVar(&serviceType, "type", "", "Service type (required)")
	addServiceCmd.MarkFlagRequired("account-id")
	addServiceCmd.MarkFlagRequired("name")
	addServiceCmd.MarkFlagRequired("type")
}
