package service

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"

	"github.com/nyambati/skiff/internal/config"
	"github.com/nyambati/skiff/internal/types"
	"github.com/nyambati/skiff/internal/utils"
	"gopkg.in/yaml.v2"
)

func (m *Manifest) ToYAML() ([]byte, error) {
	return yaml.Marshal(m)
}

func (m *Manifest) Read(path string) error {
	buff, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(buff, &m)
}

func (m *Manifest) Write(path string, verbose, force bool) error {
	buff, err := m.ToYAML()
	if err != nil {
		return err
	}

	return utils.WriteFile(path, buff)
}

func New() *Manifest {
	return &Manifest{
		APIVersion: "v1",
	}
}

func (m *Manifest) AddServiceType(name string, svcType *ServiceType) error {
	dest, exists := m.GetServiceType(name)
	if !exists {
		m.APIVersion = "v1"
		m.Types[name] = *svcType
		return nil
	}

	if err := utils.Merge(dest, svcType); err != nil {
		return err
	}

	m.Types[name] = *dest
	return nil
}

func (m *Manifest) GetServiceType(name string) (*ServiceType, bool) {
	if len(m.Types) == 0 {
		m.Types = ServiceTypes{}
		return nil, false
	}
	svcType, exists := m.Types[name]
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

	var manifest Manifest
	if err := yaml.Unmarshal(buff, &manifest); err != nil {
		return nil, err
	}

	serviceType, exists := manifest.GetServiceType(s.Type)
	if !exists {
		return nil, fmt.Errorf("service type %s does not exist, run `skiff add service-type` to add a new service type", s.Type)
	}
	s.ResolvedType = serviceType
	return s, nil
}

func (s *Service) Reconcile(accountId string, metadata types.Metadata) *Service {
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

	s.Inputs["account_id"] = accountId
	s.Inputs["region"] = s.Region
	s.Inputs["tags"] = s.Labels
	return s
}

func (s *Service) buildStrategyContext(
	svcName,
	accountID,
	accountName string,
	metadata types.Metadata,
) types.StrategyContext {
	context := types.StrategyContext{
		"service":      svcName,
		"region":       s.Region,
		"type":         s.Type,
		"account_id":   accountID,
		"account_name": accountName,
		"group":        s.ResolvedType.Group,
	}

	for key, value := range metadata {
		context[key] = value
	}

	for key, value := range s.Labels {
		context[key] = value
	}
	return context
}

func (s *Service) BuildTemplateContext(serviceName, accountID, accountName string) error {
	ctx := types.TemplateContext{
		"account_id":   accountID,
		"account_name": accountName,
		"service":      serviceName,
		"scope":        s.Scope,
		"region":       s.Region,
		"version":      s.Version,
	}
	data, err := utils.ToMap(s.ResolvedType)
	if err != nil {
		return err
	}
	for k, v := range data {
		ctx[strings.ToLower(k)] = v
	}

	ctx["inputs"] = s.Inputs
	ctx["dependencies"] = s.Dependencies
	ctx["resolved_dependencies"] = s.ResolvedDependencies
	s.TemplateContext = ctx
	return nil
}

func (s *Service) ResolveTargetPath(
	svcName,
	accountID,
	accountName string,
	metadata types.Metadata,
) error {
	strategyContext := s.buildStrategyContext(svcName, accountID, accountName, metadata)
	tmpl, err := template.New("target_path").
		Funcs(sprig.FuncMap()).
		Funcs(template.FuncMap{"var": func() types.StrategyContext { return strategyContext }}).
		Parse(config.Config.Strategy.Template)
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
	accountID,
	accountName,
	strategyTemplate string,
	services map[string]Service,
	metadata types.Metadata,
) {
	resolvedDependencies := make([]Dependency, 0, len(s.Dependencies))

	for _, dep := range s.Dependencies {
		depName := dep["service"].(string)
		targetSvc, ok := services[depName]
		if !ok {
			continue
		}

		targetSvc.Reconcile(accountID, metadata)
		targetSvc.ResolveType(config.Config.Manifests)
		targetSvc.ResolveTargetPath(depName, accountID, accountName, metadata)

		relPath, err := filepath.Rel(s.ResolvedTargetPath, targetSvc.ResolvedTargetPath)
		if err != nil {
			continue
		}

		resolvedDep := map[string]interface{}{
			"service":     depName,
			"config_path": fmt.Sprintf("${path_relative_from_include}/%s", relPath),
		}
		for k, v := range dep {
			resolvedDep[k] = v
		}
		resolvedDependencies = append(resolvedDependencies, resolvedDep)
	}
	s.ResolvedDependencies = resolvedDependencies
}
