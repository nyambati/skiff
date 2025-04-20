package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/nyambati/skiff/internal/types"
	"github.com/nyambati/skiff/internal/utils"
	"github.com/spf13/cobra"
)

var accountName string
var accountID string

var addAccountCmd = &cobra.Command{
	Use:   "account [flags]",
	Short: "Add a new account manifest",
	RunE: func(cmd *cobra.Command, args []string) error {
		manifest := types.AccountManifest{
			APIVersion: "v1",
			Kind:       "AccountDefinition",
			Accounts: types.Account{
				Name: accountName,
				ID:   accountID,
			},
		}

		data, err := manifest.ToYAML()
		if err != nil {
			return err
		}
		utils.CreateDirectory(basePath, "manifests", true)
		path := filepath.Join(basePath, "manifests", fmt.Sprintf("%s.yaml", accountID))
		if err := utils.WriteFile(path, data, false, force); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	addAccountCmd.Flags().StringVar(&accountName, "name", "", "Account name (required)")
	addAccountCmd.Flags().StringVar(&accountID, "id", "", "Account ID ")
	addAccountCmd.MarkFlagRequired("name")
	addAccountCmd.MarkFlagRequired("id")
}
