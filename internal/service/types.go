package service

import "github.com/nyambati/skiff/internal/types"

type (
	ServiceType struct {
		Source   string `yaml:"source,omitempty"`
		Group    string `yaml:"group,omitempty"`
		Version  string `yaml:"version,omitempty"`
		Template string `yaml:"template,omitempty"`
	}

	ServiceTypes map[string]ServiceType

	Service struct {
		Type                string           `yaml:"type,omitempty"`
		ResolvedType        *ServiceType     `yaml:"resolved_type,omitempty"`
		Region              string           `yaml:"region,omitempty"`
		Scope               string           `yaml:"scope,omitempty"`
		Version             string           `yaml:"version,omitempty"`
		Inputs              map[string]any   `yaml:"inputs,omitempty"`
		Labels              map[string]any   `yaml:"labels,omitempty"`
		Dependencies        []map[string]any `yaml:"dependencies,omitempty"`
		ResolveDependencies []map[string]any `yaml:"resolve_dependencies,omitempty"`
		StrategyContext     types.StrategyContext
		TemplateContext     types.TemplateContext
		ResolvedTargetPath  string
	}

	Manifest struct {
		APIVersion string                 `yaml:"apiVersion,omitempty"`
		Types      map[string]ServiceType `yaml:"types"`
	}
)
