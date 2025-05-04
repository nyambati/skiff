package cmd

import (
	"github.com/nyambati/skiff/internal/catalog"
	"github.com/spf13/cobra"
)

var addCatalogCmd = &cobra.Command{
	Use:   "catalog [flags]",
	Short: "edits the catalog file",
	Long: `The catalog command allows you to edit the catalog file.

Examples:
  skiff edit catalog --type service --values source=github.com/my-org/my-repo
`,
	Args: cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return catalog.AddServiceType(cmd.Context(), flagServiceTypeName, flagValues)
	},
}

func init() {
	addCatalogCmd.Flags().StringVar(&flagServiceTypeName, "type", "t", "service type name (required)")
	addCatalogCmd.Flags().StringVar(&flagValues, "values", "v", "service type values in key=value pairs (optional)")
	addCatalogCmd.MarkFlagRequired("type")
}
