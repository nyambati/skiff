package service

import "github.com/nyambati/skiff/internal/types"

type (
	ServiceType struct {
		Source   string   `yaml:"source,omitempty"`
		Group    string   `yaml:"group,omitempty"`
		Version  string   `yaml:"version,omitempty"`
		Template string   `yaml:"template,omitempty"`
		Outputs  []string `yaml:"outputs"`
	}
	Dependency   map[string]any
	ServiceTypes map[string]ServiceType
	Service      struct {
		Type                 string                `yaml:"type,omitempty"`
		Region               string                `yaml:"region,omitempty"`
		Scope                string                `yaml:"scope,omitempty"`
		Version              string                `yaml:"version,omitempty"`
		Inputs               map[string]any        `yaml:"inputs,omitempty"`
		Labels               map[string]any        `yaml:"labels,omitempty"`
		Dependencies         []Dependency          `yaml:"dependencies,omitempty"`
		ResolvedDependencies []Dependency          `yaml:"resolved_dependencies,omitempty"`
		ResolvedType         *ServiceType          `yaml:"resolvedtype,omitempty"`
		TemplateContext      types.TemplateContext `yaml:"templatecontext,omitempty"`
		ResolvedTargetPath   string                `yaml:"resolvedtargetpath,omitempty"`
	}

	ExportedService struct {
		Type         string         `yaml:"type,omitempty"`
		Region       string         `yaml:"region,omitempty"`
		Scope        string         `yaml:"scope,omitempty"`
		Version      string         `yaml:"version,omitempty"`
		Inputs       map[string]any `yaml:"inputs,omitempty"`
		Labels       map[string]any `yaml:"labels,omitempty"`
		Dependencies []Dependency   `yaml:"dependencies,omitempty"`
	}

	Manifest struct {
		APIVersion string                 `yaml:"apiVersion,omitempty"`
		Types      map[string]ServiceType `yaml:"types"`
	}
)
