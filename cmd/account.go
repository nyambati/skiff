package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/nyambati/skiff/internal/account"
	"github.com/nyambati/skiff/internal/utils"
	"github.com/spf13/cobra"
)

var accountName string
var accountID string

var addAccountCmd = &cobra.Command{
	Use:   "account [flags]",
	Short: "Add a new account manifest",
	RunE: func(cmd *cobra.Command, args []string) error {
		manifest := account.New("v1", accountName, accountID)
		path := filepath.Join(basePath, "manifests")

		if utils.FileExists(filepath.Join(path, fmt.Sprintf("%s.yaml", accountID))) && !force {
			fmt.Printf("⚠️ Skipping, %s already exists, use --force to overwrite\n", path)
			return nil
		}

		if err := manifest.Write(path, verbose, force); err != nil {
			return err
		}
		fmt.Printf("✅ Account %s has been added to %s\n", accountName, path)
		return nil
	},
}

func init() {
	addAccountCmd.Flags().StringVar(&accountName, "name", "", "Account name (required)")
	addAccountCmd.Flags().StringVar(&accountID, "id", "", "Account ID ")
	addAccountCmd.MarkFlagRequired("name")
	addAccountCmd.MarkFlagRequired("id")
}
