/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/nyambati/skiff/internal/types"
	"github.com/nyambati/skiff/internal/utils"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [account|service|service-type] [flags]",
	Short: "Add a new account, service, or service type",
	Args:  cobra.MinimumNArgs(0),
}

var addAccountCmd = &cobra.Command{
	Use:   "account [flags]",
	Short: "Add a new account manifest",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		id, _ := cmd.Flags().GetString("id")
		env, _ := cmd.Flags().GetString("environment")
		path, _ := cmd.Flags().GetString("path")
		force, _ := cmd.Flags().GetBool("force")

		manifest := types.Manifest{
			APIVersion: "v1",
			Accounts: types.Account{
				Name: name,
				ID:   id,
			},
			Metadata: map[string]string{"environment": env},
		}

		data, err := manifest.ToYAML()
		if err != nil {
			return err
		}
		utils.CreateDirectory(path, "manifests", true)
		path = filepath.Join(path, "manifests", fmt.Sprintf("%s.yaml", id))
		if err := utils.WriteFile(path, data, false, force); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.AddCommand(addAccountCmd)
	addAccountCmd.Flags().String("name", "", "Account name (required)")
	addAccountCmd.Flags().String("id", "", "Account ID (required)")
	addAccountCmd.Flags().String("environment", "", "Environment tag for metadata (required)")
	addAccountCmd.MarkFlagRequired("name")
	addAccountCmd.MarkFlagRequired("id")
}
