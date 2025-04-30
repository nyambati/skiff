package manifest

import "github.com/nyambati/skiff/internal/catalog"

type (
	Account struct {
		Name string `yaml:"name"`
		ID   string `yaml:"id"`
	}

	Manifest struct {
		APIVersion string                     `yaml:"apiVersion,omitempty"`
		Account    Account                    `yaml:"account,omitempty"`
		Metadata   map[string]any             `yaml:"metadata,omitempty"`
		Services   map[string]catalog.Service `yaml:"services"`
	}
)
