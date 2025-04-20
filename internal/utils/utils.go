package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func CreateDirectory(base string, path string, quiet bool) error {
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

func WriteFile(path string, content []byte, quiet, force bool) error {
	if _, err := os.Stat(path); err == nil && !force {
		if !quiet {
			fmt.Printf("⚠️  File exists, skipping: %s\n", path)
		}
		return nil
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to write file %s: %v\n", path, err)
		return err
	}
	if !quiet {
		fmt.Printf("✅ Created file: %s\n", path)
	}
	return nil
}
