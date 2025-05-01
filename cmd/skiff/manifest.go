package cmd

import (
	"fmt"

	"github.com/nyambati/skiff/internal/config"
	"github.com/nyambati/skiff/internal/manifest"
	"github.com/spf13/cobra"
)

var name string

var addAccountCmd = &cobra.Command{
	Use:   "manifest [flags]",
	Short: "Add a new account manifest",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		config, ok := ctx.Value("config").(*config.Config)
		if !ok {
			return fmt.Errorf("config not found")
		}

		manifest, err := manifest.Read(ctx, name)
		if err != nil {
			return err
		}

		if err := manifest.Write(verbose, force); err != nil {
			return err
		}
		fmt.Printf("✅ Account %s has been added to %s\n", name, config.Manifests)
		return nil
	},
}

func init() {
	addAccountCmd.Flags().StringVar(&name, "name", "", "manifest identifier ")
	addAccountCmd.MarkFlagRequired("name")
}
