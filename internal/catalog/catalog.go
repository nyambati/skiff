package catalog

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/nyambati/skiff/internal/config"
	"github.com/nyambati/skiff/internal/types"
	"github.com/nyambati/skiff/internal/utils"
	"github.com/sirupsen/logrus"
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

func (c *Catalog) Write(path string, force bool) error {
	buff, err := c.ToYAML()
	if err != nil {
		return err
	}

	return utils.WriteFile(path, utils.PrependWatermark(string(buff), config.ToolName))
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

	if err := utils.Merge(dest, svcType, append); err != nil {
		return err
	}

	// find better way to merge
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

	maps.Copy(metadata, s.Labels)

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

	maps.Copy(context, metadata)
	maps.Copy(context, s.Labels)

	return context
}

// BuildTemplateContext creates a TemplateContext for the service based on the provided
// service name, metadata, and resolved service type. The TemplateContext is used to
// render the service's Terragrunt configuration file.
//
// The generated TemplateContext will contain the following keys:
//
//   - config.ServiceKey: the name of the service
//   - config.RegionKey: the region of the service
//   - config.TypeKey: the type of the service
//   - config.GroupKey: the group of the service
//   - config.InputsKey: the inputs of the service
//   - config.DependencyKey: the dependencies of the service
//   - config.VersionKey: the version of the service
//   - config.ScopeKey: the scope of the service (regional or global)
//   - config.TerraformKey: a map containing the source of the Terraform module
//     as a key-value pair (config.SourceKey)
//   - config.BodyKey: a map containing the dependencies and inputs of the service
//     as key-value pairs (config.DependencyKey and config.InputsKey)
//
// The generated TemplateContext will also contain any additional metadata and
// labels that are provided.
func (s *Service) BuildTemplateContext(serviceName string, metadata types.Metadata) error {
	ctx := types.TemplateContext{
		config.ServiceKey:    serviceName,
		config.RegionKey:     s.Region,
		config.TypeKey:       s.Type,
		config.GroupKey:      s.ResolvedType.Group,
		config.InputsKey:     s.Inputs,
		config.DependencyKey: s.Dependencies,
		config.VersionKey:    s.ResolvedType.Version,
		config.ScopeKey:      s.Scope,
		config.TerraformKey: map[string]interface{}{
			config.SourceKey: fmt.Sprintf("%s?ref=%s", s.ResolvedType.Source, s.ResolvedType.Version),
		},
		config.BodyKey: map[string]interface{}{
			config.DependencyKey: s.Dependencies,
			config.InputsKey:     s.Inputs,
		},
	}

	maps.Copy(ctx, metadata)
	maps.Copy(ctx, s.Labels)

	s.TemplateContext = ctx
	return nil
}

func (s *Service) ResolveTargetPath(
	ctx context.Context,
	svcName string,
	metadata types.Metadata,
) error {
	cfg, err := config.FromContext(ctx)
	if err != nil {
		return err
	}

	strategyContext := s.buildStrategyContext(svcName, metadata)

	tmpl, err := template.New("").
		Option("missingkey=error").
		Funcs(sprig.FuncMap()).
		Funcs(template.FuncMap{config.VarKey: func() types.StrategyContext { return strategyContext }}).
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
			config.ServiceKey:    depName,
			config.ConfigPathKey: relPath,
		}

		for k, v := range dep {
			resolvedDep[k] = v
		}

		fmt.Println(targetSvc.ResolvedType)

		for _, output := range targetSvc.ResolvedType.Outputs {
			s.Inputs[output] = fmt.Sprintf("__dependency.%s.%s", depName, output)
		}

		resolvedDependencies = append(resolvedDependencies, resolvedDep)
	}
	s.Dependencies = resolvedDependencies
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

func buildCatalogFromValues(values string) (*ServiceType, error) {
	var outputs []string

	valuesMap := utils.ParseKeyValueFlag(values)

	outputValues, ok := valuesMap[config.OutputsKey].(string)
	if ok && outputValues != "" {
		outputs = strings.Split(outputValues, ":")
	}

	valuesMap[config.OutputsKey] = outputs

	service := &ServiceType{}

	if err := utils.StructFromMap(valuesMap, service); err != nil {
		return nil, err
	}

	if service.Template == "" {
		service.Template = config.TerragruntTemplateFile
	}

	return service, nil
}

func AddServiceType(ctx context.Context, serviceTypeName string, values string) error {
	var oldCatalog []byte
	var newCatalog []byte

	cfg, err := config.FromContext(ctx)
	if err != nil {
		return err
	}

	svcCatalog := NewCatalog()

	path := filepath.Join(cfg.Manifests, config.CatalogFile)

	if err := svcCatalog.Read(path); err != nil {
		return err
	}

	oldCatalog, err = utils.ToYAML(svcCatalog)
	if err != nil {
		return err
	}

	serviceType, exists := svcCatalog.GetServiceType(serviceTypeName)
	if !exists {
		serviceType = &ServiceType{
			Template: config.TerragruntTemplateFile,
		}
	}

	if values != "" {
		serviceType, err := buildCatalogFromValues(values)
		if err != nil {
			return err
		}

		return svcCatalog.AddServiceType(serviceTypeName, serviceType, true)

	}

	existingContent, err := utils.ToYAML(serviceType)
	if err != nil {
		return err
	}

	editContent, err := utils.EditFile(path, existingContent)
	if err != nil {
		return err
	}

	svc, err := utils.FromYAML[ServiceType](editContent)
	if err != nil {
		return err
	}

	svcCatalog.AddServiceType(serviceTypeName, svc, false)

	newCatalog, err = utils.ToYAML(svcCatalog)
	if err != nil {
		return err
	}

	if !utils.ShouldWrite(oldCatalog, newCatalog) {
		return nil
	}

	if err := svcCatalog.Write(path, true); err != nil {
		return err
	}

	logrus.Printf("✅ service type %s has been added successfuly\n", serviceTypeName)

	return nil
}
