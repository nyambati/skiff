/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nyambati/skiff/internal/catalog"
	"github.com/nyambati/skiff/internal/config"
	"github.com/nyambati/skiff/internal/manifest"
	"github.com/nyambati/skiff/internal/utils"
	"github.com/spf13/cobra"
)

var serviceName string
var serviceType string
var manifestName string

// serviceCmd represents the service command
var addServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Adds service manifests",
	Run: func(cmd *cobra.Command, args []string) {

		cfg, ok := cmd.Context().Value("config").(*config.Config)
		if !ok {
			cmd.PrintErr("config not founc")
			os.Exit(1)
		}

		if err := addService(cmd.Context(), manifestName, cfg.Manifests); err != nil {
			utils.PrintErrorAndExit(err)
		}
	},
}

func init() {
	addCmd.AddCommand(addServiceCmd)
	addServiceCmd.Flags().StringVar(&manifestName, "manifest", "", "Account manifest name (required)")
	addServiceCmd.Flags().StringVar(&serviceName, "name", "", "Service manifest name (required)")
	addServiceCmd.Flags().StringVar(&serviceType, "type", "", "Service type (required)")
	addServiceCmd.MarkFlagRequired("id")
	addServiceCmd.MarkFlagRequired("name")
	addServiceCmd.MarkFlagRequired("type")
}

func addService(ctx context.Context, manifestName string, path string) error {
	var svcCatalog catalog.Catalog

	manifest, err := manifest.Read(ctx, manifestName)
	if err != nil {
		return err
	}

	if err := svcCatalog.Read(fmt.Sprintf("%s/%s", path, config.CatalogFile)); err != nil {
		return err
	}

	if _, exists := svcCatalog.GetServiceType(serviceType); !exists {
		err := fmt.Errorf("service type %s does not exist, run `skiff add service-type` to add a new service type", serviceType)
		return err
	}

	existingContent, err := utils.ToYAML(catalog.DefaultService(serviceName, serviceType))
	if err != nil {
		return err
	}

	editContent, err := utils.EditFile(fmt.Sprintf("%s/%s.yaml", path, manifestName), existingContent)
	if err != nil {
		return err
	}

	svc, err := catalog.ServiceFromYAML(editContent)
	if err != nil {
		return err
	}

	manifest.AddService(serviceName, svc)

	if err := manifest.Write(verbose, true); err != nil {
		return err
	}
	fmt.Printf("✅ Service %s has been added to %s\n", serviceName, filepath.Join(path, manifestName))
	return nil
}
