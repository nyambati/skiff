package catalog

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/nyambati/skiff/internal/types"
	"github.com/nyambati/skiff/internal/utils"
	"gopkg.in/yaml.v2"
)

const (
	ScopeRegional = "regional"
	ScopeGlobal   = "global"
)

func (c *Catalog) ToYAML() ([]byte, error) {
	return yaml.Marshal(c)
}

func (c *Catalog) Read(path string) error {
	buff, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(buff, &c)
}

func (c *Catalog) Write(path string, verbose, force bool) error {
	buff, err := c.ToYAML()
	if err != nil {
		return err
	}

	return utils.WriteFile(path, buff)
}

func NewCatalog() *Catalog {
	return &Catalog{
		APIVersion: "v1",
		Types:      ServiceTypes{},
	}
}

func (c *Catalog) AddServiceType(name string, svcType *ServiceType) error {
	dest, exists := c.GetServiceType(name)
	if !exists {
		c.APIVersion = "v1"
		c.Types[name] = *svcType
		return nil
	}

	if err := utils.Merge(dest, svcType); err != nil {
		return err
	}

	c.Types[name] = *dest
	return nil
}

func (c *Catalog) GetServiceType(name string) (*ServiceType, bool) {
	if len(c.Types) == 0 {
		c.Types = ServiceTypes{}
		return nil, false
	}
	svcType, exists := c.Types[name]
	return &svcType, exists

}

func (s *Service) ResolveType(path string) (*Service, error) {
	if s.Type == "" {
		return nil, fmt.Errorf("service type is required")
	}

	buff, err := os.ReadFile(fmt.Sprintf("%s/service-types.yaml", path))
	if err != nil {
		return nil, err
	}

	var catalog Catalog
	if err := yaml.Unmarshal(buff, &catalog); err != nil {
		return nil, err
	}

	serviceType, exists := catalog.GetServiceType(s.Type)
	if !exists {
		return nil, fmt.Errorf("service type %s does not exist, run `skiff add service-type` to add a new service type", s.Type)
	}
	s.ResolvedType = serviceType
	return s, nil
}

func (s *Service) Reconcile(metadata types.Metadata) *Service {
	if s.Version == "" {
		s.Version = s.ResolvedType.Version
	}

	if len(s.Labels) == 0 {
		s.Labels = map[string]any{}
	}

	for key, value := range metadata {
		s.Labels[key] = value
	}

	if len(s.Inputs) == 0 {
		s.Inputs = map[string]any{}
	}

	s.Inputs["region"] = s.Region
	s.Inputs["tags"] = s.Labels
	return s
}

func (s *Service) buildStrategyContext(svcName string, metadata types.Metadata) types.StrategyContext {
	context := types.StrategyContext{
		"service": svcName,
		"region":  s.Region,
		"type":    s.Type,
		"group":   s.ResolvedType.Group,
	}

	for key, value := range metadata {
		context[key] = value
	}

	for key, value := range s.Labels {
		context[key] = value
	}
	return context
}

func (s *Service) BuildTemplateContext(serviceName string, metadata types.Metadata) error {
	ctx := types.TemplateContext{
		"service": serviceName,
		"scope":   s.Scope,
		"region":  s.Region,
		"version": s.Version,
	}

	data, err := utils.ToMap(s.ResolvedType)
	if err != nil {
		return err
	}

	for k, v := range data {
		ctx[strings.ToLower(k)] = v
	}

	for key, value := range s.Labels {
		ctx[strings.ToLower(key)] = value
	}

	for key, value := range metadata {
		ctx[strings.ToLower(key)] = value
	}

	ctx["inputs"] = s.Inputs
	ctx["dependencies"] = s.Dependencies
	s.TemplateContext = ctx
	return nil
}

func (s *Service) ResolveTargetPath(
	svcName,
	strategyTemplate string,
	metadata types.Metadata,
) error {
	strategyContext := s.buildStrategyContext(svcName, metadata)
	tmpl, err := template.New("target_path").
		Funcs(sprig.FuncMap()).
		Funcs(template.FuncMap{"var": func() types.StrategyContext { return strategyContext }}).
		Parse(strategyTemplate)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return err
	}
	resolvedPath, err := validatePath(&buf)
	s.ResolvedTargetPath = utils.SanitizePath(resolvedPath)
	return err
}

func validatePath(buffer *bytes.Buffer) (string, error) {
	// Check if template markers still exist
	if bytes.Contains(buffer.Bytes(), []byte("{{")) || bytes.Contains(buffer.Bytes(), []byte("}}")) {
		return "", fmt.Errorf("template was not fully rendered: unresolved variables remain in %q", buffer.String())
	}
	return buffer.String(), nil
}

func (s *Service) ResolveDependencies(
	manifestName,
	manifestPath string,
	strategyTemplate string,
	services map[string]Service,
	metadata types.Metadata,
) {
	resolvedDependencies := make([]Dependency, 0, len(s.Dependencies))

	for _, dep := range s.Dependencies {
		depName, ok := dep["service"].(string)
		if !ok {
			continue
		}

		targetSvc, ok := services[depName]
		if !ok {
			continue
		}

		targetSvc.Reconcile(metadata)
		targetSvc.ResolveType(manifestPath)
		targetSvc.ResolveTargetPath(depName, strategyTemplate, metadata)

		relPath, err := filepath.Rel(s.ResolvedTargetPath, targetSvc.ResolvedTargetPath)
		if err != nil {
			continue
		}

		resolvedDep := map[string]any{
			"service":     depName,
			"config_path": fmt.Sprintf("${path_relative_from_include}/%s", relPath),
		}

		for k, v := range dep {
			resolvedDep[k] = v
		}

		fmt.Println(targetSvc.ResolvedType)

		for _, output := range targetSvc.ResolvedType.Outputs {
			s.Inputs[output] = fmt.Sprintf("dependency.%s.%s", depName, output)
		}

		resolvedDependencies = append(resolvedDependencies, resolvedDep)
	}
	s.Dependencies = resolvedDependencies
}

func ServiceFromYAML(data []byte) (*Service, error) {
	var svc Service
	if err := yaml.Unmarshal(data, &svc); err != nil {
		return nil, err
	}
	return &svc, nil
}

func DefaultService(name string) *Service {
	return &Service{
		Type:    "default",
		Scope:   ScopeRegional,
		Version: "1.0",
		Region:  "us-east-1",
		Inputs:  map[string]any{},
		Labels: map[string]any{
			"type":  "default",
			"scope": ScopeRegional,
			"name":  name,
		},
		Dependencies: []Dependency{},
	}
}
