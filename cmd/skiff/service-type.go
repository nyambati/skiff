package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/nyambati/skiff/internal/catalog"
	"github.com/nyambati/skiff/internal/config"
	"github.com/spf13/cobra"
)

var serviceTypeName string
var serviceTypeSource string
var serviceTypeGroup string
var serviceTypeVersion string
var serviceTypeTemplate string

var addCatalogCmd = &cobra.Command{
	Use:   "catalog [flags]",
	Short: "Add a new service type to catalog",
	RunE: func(cmd *cobra.Command, args []string) error {
		var svcCatalog catalog.Catalog
		cfg, ok := cmd.Context().Value("config").(*config.Config)
		if !ok {
			return fmt.Errorf("config not found")
		}
		path := filepath.Join(cfg.Manifests, "service-types.yaml")

		if err := svcCatalog.Read(path); err != nil {
			return err
		}

		svcCatalog.AddServiceType(
			serviceTypeName,
			&catalog.ServiceType{
				Source:   serviceTypeSource,
				Group:    serviceTypeGroup,
				Version:  serviceTypeVersion,
				Template: serviceTypeTemplate,
			},
		)

		if err := svcCatalog.Write(path, verbose, force); err != nil {
			return err
		}

		fmt.Printf("✅ Service type '%s' added to %s\n", serviceTypeName, path)
		return nil

	},
}

func init() {
	addCatalogCmd.Flags().StringVar(&serviceTypeSource, "source", "", "terraform module source (required)")
	addCatalogCmd.Flags().StringVar(&serviceTypeGroup, "group", "", "group to generate terragrunt files")
	addCatalogCmd.Flags().StringVar(&serviceTypeVersion, "version", "", "default module version (required)")
	addCatalogCmd.Flags().StringVar(&serviceTypeTemplate, "template", "terragrunt.default.tmpl", "terragrunt template")
	addCatalogCmd.Flags().StringVar(&serviceTypeName, "name", "", "Service type name (required)")

	addCatalogCmd.MarkFlagRequired("source")
	addCatalogCmd.MarkFlagRequired("version")
	addCatalogCmd.MarkFlagRequired("name")
}
