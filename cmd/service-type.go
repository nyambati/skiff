package cmd

import (
	"os"
	"path/filepath"

	"dario.cat/mergo"
	"github.com/nyambati/skiff/internal/types"
	"github.com/nyambati/skiff/internal/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var serviceTypeName string
var serviceTypeSource string
var serviceTypeFolder string
var serviceTypeVersion string
var serviceTypeTemplate string
var kind string = "ServiceTypeDefinition"

var addServiceTypeCmd = &cobra.Command{
	Use:   "service-type [flags]",
	Short: "Add a new service type in service-types.yaml",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := filepath.Join(basePath, "manifests", "service-types.yaml")
		manifest, err := readServiceTypeManifest(path)
		if err != nil {
			return err
		}

		overrideManifest := types.ServiceTypeManifest{
			APIVersion: "v1",
			Kind:       kind,
			Types: map[string]types.ServiceType{
				serviceTypeName: {
					Source:   serviceTypeSource,
					Folder:   serviceTypeFolder,
					Version:  serviceTypeVersion,
					Template: serviceTypeTemplate,
				},
			},
		}

		if err := mergo.Merge(&manifest, &overrideManifest, mergo.WithOverride); err != nil {
			return err
		}

		data, err := manifest.ToYAML()
		if err != nil {
			return err
		}
		return utils.WriteFile(path, data, quiet, true)
	},
}

func readServiceTypeManifest(path string) (types.ServiceTypeManifest, error) {
	var serviceTypeManifest types.ServiceTypeManifest
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return serviceTypeManifest, err
	}
	if err := yaml.Unmarshal(data, &serviceTypeManifest); err != nil {
		return serviceTypeManifest, err
	}
	return serviceTypeManifest, nil
}

func init() {
	addServiceTypeCmd.Flags().StringVar(&serviceTypeSource, "source", "", "terraform module source (required)")
	addServiceTypeCmd.Flags().StringVar(&serviceTypeFolder, "folder", "", "folder to generate terragrunt files")
	addServiceTypeCmd.Flags().StringVar(&serviceTypeVersion, "version", "", "default module version (required)")
	addServiceTypeCmd.Flags().StringVar(&serviceTypeTemplate, "template", "terragrunt.default.tmpl", "terragrunt template")
	addServiceTypeCmd.Flags().StringVar(&serviceTypeName, "name", "", "Service type name (required)")

	addServiceTypeCmd.MarkFlagRequired("source")
	addServiceTypeCmd.MarkFlagRequired("version")
	addServiceTypeCmd.MarkFlagRequired("name")
}
