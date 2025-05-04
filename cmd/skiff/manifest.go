package cmd

import (
	"github.com/nyambati/skiff/internal/manifest"
	"github.com/spf13/cobra"
)

var addAccountCmd = &cobra.Command{
	Use:   "manifest [flags]",
	Short: "edits manifest files",
	Args:  cobra.MinimumNArgs(0),
	Long: `The manifest command allows you to edit the manifest file.

Examples:
  skiff edit manifest --manifest my-manifest --metadata env=production,account_id=12345
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return manifest.EditManifest(cmd.Context(), flagManifestID, flagMetadata)
	},
}

func init() {
	addAccountCmd.Flags().StringVarP(&flagManifestID, "manifest", "m", "", "manifest identifier ")
	addAccountCmd.Flags().StringVar(&flagMetadata, "metadata", "", "manifestmetadata")
	addAccountCmd.MarkFlagRequired("manifest")
}
