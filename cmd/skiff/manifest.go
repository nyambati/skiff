package cmd

import (
	"github.com/nyambati/skiff/internal/manifest"
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
		return manifest.EditManifest(cmd.Context(), name, metadata)
	},
}

func init() {
	addAccountCmd.Flags().StringVar(&name, "name", "", "manifest identifier ")
	addAccountCmd.Flags().StringVar(&metadata, "metadata", "", "manifestmetadata")
	addAccountCmd.MarkFlagRequired("name")
}
