package strategy

import (
	"github.com/nyambati/skiff/internal/service"
)

type (
	Config struct {
		Template     string
		Data         TemplateData
		TargetFolder string
	}
	Strategy struct {
		Path string `yaml:"path"`
	}

	StrategyFunc    func(manifests []*service.Manifest, catalog *service.Manifest, labels string) *RenderConfig
	RenderConfig    []Config
	StrategyContext map[string]any
	metadata        map[string]any
	TemplateData    struct {
		service.Service
		Source string
	}
)
