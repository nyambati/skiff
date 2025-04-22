/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/nyambati/skiff/internal/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var basePath string
var verbose bool
var force bool
var skiffConfig types.SkiffConfig

var rootCmd = &cobra.Command{
	Use:   "skiff",
	Short: "A tool to generate and apply Terragrunt configurations from YAML manifests",
	Long: `Skiff is a CLI tool that helps you define, generate,
and apply infrastructure using a declarative YAML format and Terragrunt.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", true, "quiet mode")
	rootCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "force overwrite")

	viper.AddConfigPath(".")
	viper.SetConfigName(".skiff")
	viper.SetConfigType("yaml")

	// bind viper to cobra
	viper.BindPFlag("path", rootCmd.PersistentFlags().Lookup("path"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("force", rootCmd.PersistentFlags().Lookup("force"))

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	if err := viper.Unmarshal(&skiffConfig); err != nil {
		fmt.Println("Error unmarshalling config:", err)
	}
}
