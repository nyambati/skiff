package strategy

import (
	"github.com/nyambati/skiff/internal/types"
)

type (
	Config struct {
		Template     string
		Context      *types.TemplateContext
		TargetFolder string
	}

	Strategy struct {
		Description string `yaml:"description"`
		Path        string `yaml:"path"`
	}

	RenderConfig []Config
)
