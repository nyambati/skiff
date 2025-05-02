package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/nyambati/skiff/internal/manifest"
	"github.com/nyambati/skiff/internal/utils"
	"github.com/spf13/cobra"
)

var name string
var metadata string

var addAccountCmd = &cobra.Command{
	Use:   "manifest [flags]",
	Short: "edits manifest files",
	Args:  cobra.MinimumNArgs(0),
	Long: `The manifest command allows you to edit the manifest file.

Examples:
  skiff edit manifest --name my-manifest --metadata env=production,account_id=12345
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		manifest, err := manifest.Read(ctx, name)
		if err != nil {
			return err
		}

		if err := editManifest(ctx, metadata, manifest); err != nil {
			return err
		}

		return manifest.Write(verbose, true)
	},
}

func init() {
	addAccountCmd.Flags().StringVar(&name, "name", "", "manifest identifier ")
	addAccountCmd.Flags().StringVar(&metadata, "metadata", "", "manifestmetadata")
	addAccountCmd.MarkFlagRequired("name")
}

func editManifest(ctx context.Context, metadata string, m *manifest.Manifest) error {

	cfg, err := utils.GetConfigFromContext(ctx)
	if err != nil {
		return err
	}

	if metadata != "" {
		metadata := utils.ParseKeyValueFlag(metadata)

		for k, v := range metadata {
			m.Metadata[strings.ToLower(k)] = v
		}

	}

	content, err := utils.ToYAML(m)
	if err != nil {
		return err
	}

	content, err = utils.EditFile(fmt.Sprintf("%s/%s.yaml", cfg.Manifests, m.Name), content)
	if err != nil {
		return err
	}

	m, err = utils.FromYAML[manifest.Manifest](content)
	if err != nil {
		return err
	}
	return nil
}
