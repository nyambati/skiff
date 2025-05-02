package catalog

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/nyambati/skiff/internal/config"
	"github.com/nyambati/skiff/internal/types"
	"github.com/nyambati/skiff/internal/utils"
	"gopkg.in/yaml.v3"
)

func (c *Catalog) ToYAML() ([]byte, error) {
	var buff bytes.Buffer
	encoder := yaml.NewEncoder(&buff)
	encoder.SetIndent(2)
	defer encoder.Close()

	if err := encoder.Encode(c); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
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

func (c *Catalog) AddServiceType(name string, svcType *ServiceType, append bool) error {
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

func (s *Service) ResolveType(ctx context.Context) (*Service, error) {
	cfg, ok := ctx.Value(config.ContextKey).(*config.Config)
	if !ok {
		return nil, fmt.Errorf("config not found")
	}
	if s.Type == "" {
		return nil, fmt.Errorf("service type is required")
	}

	buff, err := os.ReadFile(fmt.Sprintf("%s/%s", cfg.Manifests, config.CatalogFile))
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

	for key, value := range s.Labels {
		metadata[key] = value
	}

	if len(s.Inputs) == 0 {
		s.Inputs = map[string]any{}
	}

	s.Inputs[config.RegionKey] = s.Region
	s.Inputs[config.TagsKey] = metadata
	s.Labels = metadata
	return s
}

func (s *Service) buildStrategyContext(svcName string, metadata types.Metadata) types.StrategyContext {
	context := types.StrategyContext{
		config.ServiceKey: svcName,
		config.RegionKey:  s.Region,
		config.TypeKey:    s.Type,
		config.GroupKey:   s.ResolvedType.Group,
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
		config.ServiceKey: serviceName,
		config.ScopeKey:   s.Scope,
		config.RegionKey:  s.Region,
		config.VersionKey: s.Version,
	}

	data, err := utils.ToMap(s.ResolvedType)
	if err != nil {
		return err
	}

	for k, v := range data {
		ctx[strings.ToLower(k)] = v
	}

	for key, value := range metadata {
		ctx[strings.ToLower(key)] = value
	}

	for key, value := range s.Labels {
		ctx[strings.ToLower(key)] = value
	}

	ctx[config.InputsKey] = s.Inputs
	ctx[config.DependencyKey] = s.Dependencies
	s.TemplateContext = ctx
	return nil
}

func (s *Service) ResolveTargetPath(
	ctx context.Context,
	svcName string,
	metadata types.Metadata,
) error {
	cfg, ok := ctx.Value(config.ContextKey).(*config.Config)
	if !ok {
		return fmt.Errorf("config not found")
	}
	strategyContext := s.buildStrategyContext(svcName, metadata)
	tmpl, err := template.New("target_path").
		Funcs(sprig.FuncMap()).
		Funcs(template.FuncMap{"var": func() types.StrategyContext { return strategyContext }}).
		Parse(cfg.Strategy.Template)
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
	ctx context.Context,
	manifestName string,
	services map[string]Service,
	metadata types.Metadata,
) {
	resolvedDependencies := make([]Dependency, 0, len(s.Dependencies))

	for _, dep := range s.Dependencies {
		depName, ok := dep[config.ServiceKey].(string)
		if !ok {
			continue
		}

		targetSvc, ok := services[depName]
		if !ok {
			continue
		}

		targetSvc.Reconcile(metadata)
		targetSvc.ResolveType(ctx)
		targetSvc.ResolveTargetPath(ctx, depName, metadata)

		relPath, err := filepath.Rel(s.ResolvedTargetPath, targetSvc.ResolvedTargetPath)
		if err != nil {
			continue
		}

		resolvedDep := map[string]any{
			config.ServiceKey: depName,
			"config_path":     fmt.Sprintf("${path_relative_from_include}/%s", relPath),
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

func FromYAML[T any](data []byte) (*T, error) {
	var inter T
	if err := yaml.Unmarshal(data, &inter); err != nil {
		return nil, err
	}
	return &inter, nil
}

func DefaultService(name, serviceType string) *Service {
	return &Service{
		Type:    serviceType,
		Scope:   config.ScopeRegional,
		Version: "1.0",
		Region:  "us-east-1",
		Inputs:  map[string]any{},
		Labels: map[string]any{
			config.TypeKey:  serviceType,
			config.ScopeKey: config.ScopeRegional,
			config.NameKey:  name,
		},
		Dependencies: []Dependency{},
	}
}
