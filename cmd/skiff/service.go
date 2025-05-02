/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/nyambati/skiff/internal/catalog"
	"github.com/nyambati/skiff/internal/config"
	skiff "github.com/nyambati/skiff/internal/errors"
	"github.com/nyambati/skiff/internal/manifest"
	"github.com/nyambati/skiff/internal/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var serviceName string
var manifestName string

// serviceCmd represents the service command
var addServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "edits specific service in manifest file",
	Long: `The service command allows you to edit a specific service in the manifest file.

Examples:
  skiff edit service --manifest my-manifest --service my-service
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := addService(cmd.Context(), manifestName); err != nil {
			utils.PrintErrorAndExit(err)
		}
	},
}

func init() {
	editCmd.AddCommand(addServiceCmd)
	addServiceCmd.Flags().StringVar(&manifestName, "manifest", "", "name of the manifest file")
	addServiceCmd.Flags().StringVar(&serviceName, "service", "", "name of the service in the manifest file")
	addServiceCmd.MarkFlagRequired("manifest")
	addServiceCmd.MarkFlagRequired("service")
}

func addService(ctx context.Context, manifestName string) error {
	var svcCatalog catalog.Catalog

	cfg, err := utils.GetConfigFromContext(ctx)
	if err != nil {
		return err
	}

	catalogFilePath := fmt.Sprintf("%s/%s", cfg.Manifests, config.CatalogFile)
	manifestFilePath := fmt.Sprintf("%s/%s.yaml", cfg.Manifests, manifestName)

	manifest, err := manifest.Read(ctx, manifestName)
	if err != nil {
		return err
	}

	if err := svcCatalog.Read(catalogFilePath); err != nil {
		return err
	}

	svc, ok := manifest.GetService(serviceName)
	if !ok {
		svc = catalog.DefaultService(serviceName, "")
	}

	content, err := utils.ToYAML(svc)
	if err != nil {
		return err
	}

	content, err = utils.EditFile(manifestFilePath, content)
	if err != nil {
		return err
	}

	svc, err = utils.FromYAML[catalog.Service](content)
	if err != nil {
		return err
	}

	if _, exists := svcCatalog.GetServiceType(svc.Type); !exists {
		return skiff.NewServiceTypeDoesNotExistError(svc.Type)
	}

	manifest.AddService(serviceName, svc)

	if err := manifest.Write(verbose, true); err != nil {
		return err
	}
	logrus.Infof("✅ Service %s has been added to %s\n", serviceName, manifestFilePath)
	return nil
}
