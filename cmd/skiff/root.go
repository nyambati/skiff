/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"os"
	"strings"

	"github.com/nyambati/skiff/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var verbose bool
var force bool
var logger *logrus.Logger
var rootCmd = &cobra.Command{
	Use:   "skiff",
	Short: "A tool to generate and apply Terragrunt configurations from YAML manifests",
	Long: `Skiff is a CLI tool that helps you define, generate,
and apply infrastructure using a declarative YAML format and Terragrunt.`,
	CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cfg, err := config.New(cmd.Name())
		if err != nil {
			if strings.Contains(err.Error(), "Not Found") {
				cmd.Println("❌ Missing .skiff file. Run `skiff init` to create one ")
				os.Exit(1)
			}
			logrus.Error(err)
			os.Exit(1)
		}

		ctx := context.WithValue(context.Background(), config.ContextKey, cfg)
		cmd.SetContext(ctx)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose mode")
	rootCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "force overwrite")

	logrus.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		PadLevelText:     true,
	})

	logrus.SetLevel(logrus.InfoLevel)

	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}
}
