package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/nyambati/skiff/internal/catalog"
	"github.com/nyambati/skiff/internal/config"
	"github.com/nyambati/skiff/internal/utils"
	"github.com/spf13/cobra"
)

var serviceTypeName string
var values string

var addCatalogCmd = &cobra.Command{
	Use:   "catalog [flags]",
	Short: "edits the catalog file",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := utils.GetConfigFromContext(cmd.Context())
		if err != nil {
			return err
		}

		svcCatalog := catalog.NewCatalog()

		path := filepath.Join(cfg.Manifests, config.CatalogFile)

		if err := svcCatalog.Read(path); err != nil {
			return err
		}

		if err := addServiceType(cmd.Context(), serviceTypeName, values, svcCatalog); err != nil {
			return err
		}

		if err := svcCatalog.Write(path, verbose, true); err != nil {
			return err
		}

		fmt.Printf("✅ Service %s has been added to %s\n", serviceName, filepath.Join(path, manifestName))
		return nil

	},
}

func init() {
	addCatalogCmd.Flags().StringVar(&serviceTypeName, "name", "", "service type name (required)")
	addCatalogCmd.Flags().StringVar(&values, "values", "", "service type values in key=value pairs (optional)")
	addCatalogCmd.MarkFlagRequired("name")
}

func buildCatalogFromValues(values string) (*catalog.ServiceType, error) {
	var outputs []string

	valuesMap := utils.ParseKeyValueFlag(values)

	outputValues, ok := valuesMap["outputs"].(string)
	if ok && outputValues != "" {
		outputs = strings.Split(outputValues, ":")
	}

	valuesMap["outputs"] = outputs

	service := &catalog.ServiceType{}

	if err := utils.StructFromMap(valuesMap, service); err != nil {
		return nil, err
	}

	if service.Template == "" {
		service.Template = config.TerragruntTemplateFile
	}

	return service, nil
}

func addServiceType(ctx context.Context, serviceTypeName string, values string, svcCatalog *catalog.Catalog) error {

	cfg, err := utils.GetConfigFromContext(ctx)
	if err != nil {
		return err
	}

	serviceType, exists := svcCatalog.GetServiceType(serviceTypeName)
	if !exists {
		serviceType = &catalog.ServiceType{
			Template: config.TerragruntTemplateFile,
		}
	}

	if values != "" {
		serviceType, err := buildCatalogFromValues(values)
		if err != nil {
			return err
		}

		return svcCatalog.AddServiceType(serviceTypeName, serviceType, true)

	}

	existingContent, err := utils.ToYAML(serviceType)
	if err != nil {
		return err
	}

	editContent, err := utils.EditFile(fmt.Sprintf("%s/%s.yaml", cfg.Manifests, config.CatalogFile), existingContent)
	if err != nil {
		return err
	}

	svc, err := catalog.FromYAML[catalog.ServiceType](editContent)
	if err != nil {
		return err
	}

	svcCatalog.AddServiceType(serviceTypeName, svc, false)
	return nil
}
