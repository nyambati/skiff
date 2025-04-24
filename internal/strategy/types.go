package strategy

import (
	"github.com/nyambati/skiff/internal/service"
)

type (
	Config struct {
		Template     string
		Context      *TemplateContext
		TargetFolder string
	}
	Strategy struct {
		Description string `yaml:"description"`
		Path        string `yaml:"path"`
	}

	StrategyFunc    func(manifests []*service.Manifest, catalog *service.Manifest, labels string) *RenderConfig
	RenderConfig    []Config
	StrategyContext map[string]any
	metadata        map[string]any
	TemplateContext map[string]any
)
