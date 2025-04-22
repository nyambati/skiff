package cmd

import (
	"path/filepath"

	"github.com/nyambati/skiff/internal/account"
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
		return manifest.Write(path, verbose, force)
	},
}

func init() {
	addAccountCmd.Flags().StringVar(&accountName, "name", "", "Account name (required)")
	addAccountCmd.Flags().StringVar(&accountID, "id", "", "Account ID ")
	addAccountCmd.MarkFlagRequired("name")
	addAccountCmd.MarkFlagRequired("id")
}
