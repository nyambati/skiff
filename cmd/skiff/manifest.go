package cmd

import (
	"strings"

	"github.com/nyambati/skiff/internal/manifest"
	"github.com/nyambati/skiff/internal/utils"
	"github.com/spf13/cobra"
)

var name string
var metadata string

var addAccountCmd = &cobra.Command{
	Use:   "manifest [flags]",
	Short: "Add a new account manifest",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		manifest, err := manifest.Read(ctx, name)
		if err != nil {
			return err
		}

		metadata := utils.ParseKeyValueFlag(metadata)

		for k, v := range metadata {
			manifest.Metadata[strings.ToLower(k)] = v
		}

		return manifest.Write(verbose, force)
	},
}

func init() {
	addAccountCmd.Flags().StringVar(&name, "name", "", "manifest identifier ")
	addAccountCmd.Flags().StringVar(&metadata, "metadata", "", "manifestmetadata")
	addAccountCmd.MarkFlagRequired("name")
}
