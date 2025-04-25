package strategy

import (
	"github.com/nyambati/skiff/internal/account"
	"github.com/nyambati/skiff/internal/service"
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

	StrategyFunc func(manifests []*service.Manifest, catalog *service.Manifest, labels string) *RenderConfig
	RenderConfig []Config
	metadata     map[string]any
	LayoutFunc   func(
		serviceName string,
		svc *service.Service,
		typeDef *service.ServiceType,
		manifest *account.Manifest,
		service service.Service,
	) (string, error)
)
