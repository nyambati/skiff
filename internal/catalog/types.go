package catalog

import "github.com/nyambati/skiff/internal/types"

type (
	ServiceType struct {
		Source   string   `yaml:"source,omitempty"`
		Group    string   `yaml:"group,omitempty"`
		Version  string   `yaml:"version,omitempty"`
		Template string   `yaml:"template,omitempty"`
		Outputs  []string `yaml:"outputs"`
	}

	Dependency map[string]any

	ServiceTypes map[string]ServiceType

	Service struct {
		Type                 string                `yaml:"type,omitempty"`
		Region               string                `yaml:"region,omitempty"`
		Scope                string                `yaml:"scope,omitempty"`
		Version              string                `yaml:"version,omitempty"`
		Inputs               map[string]any        `yaml:"inputs"`
		Labels               map[string]any        `yaml:"labels"`
		Dependencies         []Dependency          `yaml:"dependencies"`
		ResolvedDependencies []Dependency          `yaml:"-"`
		ResolvedType         *ServiceType          `yaml:"-"`
		TemplateContext      types.TemplateContext `yaml:"-"`
		ResolvedTargetPath   string                `yaml:"-"`
	}

	Catalog struct {
		APIVersion string                 `yaml:"apiVersion,omitempty"`
		Types      map[string]ServiceType `yaml:"types"`
	}
)
