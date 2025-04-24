package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/nyambati/skiff/internal/config"
	"github.com/nyambati/skiff/internal/service"
	"github.com/spf13/cobra"
)

var serviceTypeName string
var serviceTypeSource string
var serviceTypeGroup string
var serviceTypeVersion string
var serviceTypeTemplate string

var addServiceTypeCmd = &cobra.Command{
	Use:   "service-type [flags]",
	Short: "Add a new service type in service-types.yaml",
	RunE: func(cmd *cobra.Command, args []string) error {
		var manifest service.Manifest
		path := filepath.Join(config.Config.Manifests, "service-types.yaml")

		if err := manifest.Read(path); err != nil {
			return err
		}

		manifest.AddServiceType(
			serviceTypeName,
			&service.ServiceType{
				Source:   serviceTypeSource,
				Group:    serviceTypeGroup,
				Version:  serviceTypeVersion,
				Template: serviceTypeTemplate,
			},
		)

		if err := manifest.Write(path, verbose, force); err != nil {
			return err
		}

		fmt.Printf("✅ Service type '%s' added to %s\n", serviceTypeName, path)
		return nil

	},
}

func init() {
	addServiceTypeCmd.Flags().StringVar(&serviceTypeSource, "source", "", "terraform module source (required)")
	addServiceTypeCmd.Flags().StringVar(&serviceTypeGroup, "group", "", "group to generate terragrunt files")
	addServiceTypeCmd.Flags().StringVar(&serviceTypeVersion, "version", "", "default module version (required)")
	addServiceTypeCmd.Flags().StringVar(&serviceTypeTemplate, "template", "terragrunt.default.tmpl", "terragrunt template")
	addServiceTypeCmd.Flags().StringVar(&serviceTypeName, "name", "", "Service type name (required)")

	addServiceTypeCmd.MarkFlagRequired("source")
	addServiceTypeCmd.MarkFlagRequired("version")
	addServiceTypeCmd.MarkFlagRequired("name")
}
