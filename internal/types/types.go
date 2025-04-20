package types

import (
	"gopkg.in/yaml.v2"
)

type (
	Account struct {
		Name string `yaml:"name"`
		ID   string `yaml:"id"`
	}

	Service struct {
		Name    string `yaml:"name"`
		Type    string `yaml:"type"`
		Source  string `yaml:"source"`
		Folder  string `yaml:"folder"`
		Version string `yaml:"version"`
	}

	Manifest struct {
		APIVersion string            `yaml:"apiVersion"`
		Accounts   Account           `yaml:"account"`
		Metadata   map[string]string `yaml:"metadata"`
		Services   []Service         `yaml:"services"`
	}
)

func (m *Manifest) ToYAML() ([]byte, error) {
	return yaml.Marshal(m)
}
