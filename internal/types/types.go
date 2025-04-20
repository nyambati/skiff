package types

import (
	"gopkg.in/yaml.v2"
)

type (
	Account struct {
		Name string `yaml:"name"`
		ID   string `yaml:"id"`
	}

	ServiceType struct {
		Source   string `yaml:"source,omitempty"`
		Folder   string `yaml:"folder,omitempty"`
		Version  string `yaml:"version,omitempty"`
		Template string `yaml:"template,omitempty"`
	}

	Service struct{}

	AccountManifest struct {
		APIVersion string            `yaml:"apiVersion,omitempty"`
		Kind       string            `yaml:"kind,omitempty"`
		Accounts   Account           `yaml:"account,omitempty"`
		Metadata   map[string]string `yaml:"metadata,omitempty"`
		Services   []Service         `yaml:"services"`
	}

	ServiceTypeManifest struct {
		APIVersion string                 `yaml:"apiVersion,omitempty"`
		Kind       string                 `yaml:"kind,omitempty"`
		Types      map[string]ServiceType `yaml:"types"`
	}
)

func (m *AccountManifest) ToYAML() ([]byte, error) {
	return yaml.Marshal(m)
}

func (svcType *ServiceTypeManifest) ToYAML() ([]byte, error) {
	return yaml.Marshal(svcType)
}
