package manifest

import (
	"github.com/nyambati/skiff/internal/catalog"
	"github.com/nyambati/skiff/internal/types"
)

type (
	Manifest struct {
		Name       string                     `yaml:"-"`
		APIVersion string                     `yaml:"apiVersion,omitempty"`
		Metadata   types.Metadata             `yaml:"metadata,omitempty"`
		Services   map[string]catalog.Service `yaml:"services"`
		filepath   string                     `yaml:"-"`
	}
)
