package strategy

import (
	"github.com/nyambati/skiff/internal/account"
	"github.com/nyambati/skiff/internal/service"
)

type (
	Config struct {
		Template     string
		Data         TemplateData
		TargetFolder string
	}
	RenderConfig []Config

	Strategy func(manifests []*account.Manifest, svcTypes *service.Manifest) *RenderConfig

	metadata map[string]any

	TemplateData struct {
		service.Service
		Source string
	}
)
