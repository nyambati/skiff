package account

import (
	"github.com/nyambati/skiff/internal/service"
)

type (
	Account struct {
		Name string `yaml:"name"`
		ID   string `yaml:"id"`
	}

	Manifest struct {
		APIVersion string                     `yaml:"apiVersion,omitempty"`
		Account    Account                    `yaml:"account,omitempty"`
		Metadata   map[string]any             `yaml:"metadata,omitempty"`
		Services   map[string]service.Service `yaml:"services"`
	}
)
