/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

//go:embed templates/terragrunt.default.tmpl
var terragruntDefaultTemplate []byte

//go:embed templates/service-types.yaml
var serviceTypesTemplate []byte

type Config struct {
	Folder       string
	Template     []byte
	TemplateName string
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [path] [flags]",
	Short: "Initialize a new Skiff project",
	Long: `Creates the required folder structure for a Skiff project, including:
- manifests/ (with an empty service-types.yaml)
- templates/ (with a default terragrunt.default.tmpl)
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		quiet, _ := cmd.Flags().GetBool("quiet")
		force, _ := cmd.Flags().GetBool("force")

		basePath := "."
		if len(args) >= 1 {
			basePath = args[0]
		}

		cfg := []Config{
			{
				Folder:       "manifests",
				TemplateName: "service-types.yaml",
				Template:     serviceTypesTemplate,
			},
			{
				Folder:       "templates",
				TemplateName: "terragrunt.default.tmpl",
				Template:     terragruntDefaultTemplate,
			},
		}

		for _, c := range cfg {
			if err := createDirectory(basePath, c.Folder, quiet); err != nil {
				return err
			}
			createFile(filepath.Join(basePath, c.Folder, c.TemplateName), c.Template, quiet, force)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolP("quiet", "q", false, "Quiet mode")
	initCmd.Flags().BoolP("force", "f", false, "Force overwrite if the base directory already exists")
}

// createDirectories creates a directory at the given path, relative to the
// base directory. If the directory already exists, it will be overwritten
// if the force flag is set.
func createDirectory(base string, path string, quiet bool) error {
	path = filepath.Join(base, path)
	if err := os.MkdirAll(path, 0755); err != nil && !os.IsExist(err) {
		fmt.Fprintf(os.Stderr, "Error creating directory %s: %s\n", path, err)
		return err
	}
	if !quiet {
		fmt.Printf("Created directory %s\n", path)
	}
	return nil
}

func createFile(path string, content []byte, quiet, force bool) {
	if _, err := os.Stat(path); err == nil && !force {
		if !quiet {
			fmt.Printf("⚠️  File exists, skipping: %s\n", path)
		}
		return
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to write file %s: %v\n", path, err)
		os.Exit(1)
	}
	if !quiet {
		fmt.Printf("✅ Created file: %s\n", path)
	}
}
