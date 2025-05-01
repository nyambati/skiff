package manifest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nyambati/skiff/internal/catalog"
	"github.com/nyambati/skiff/internal/config"
	"github.com/nyambati/skiff/internal/types"
	"github.com/nyambati/skiff/internal/utils"
	"gopkg.in/yaml.v2"
)

func Read(ctx context.Context, manifestName string) (*Manifest, error) {
	config, ok := ctx.Value("config").(*config.Config)
	if !ok {
		return nil, fmt.Errorf("config not found")
	}

	m := &Manifest{
		APIVersion: "v1",
		Name:       manifestName,
		Metadata:   types.Metadata{"name": manifestName},
		filepath:   filepath.Join(config.Manifests, fmt.Sprintf("%s.yaml", manifestName)),
	}

	if err := m.read(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Manifest) Write(verbose, force bool) error {
	data, err := m.ToYAML()
	if err != nil {
		return err
	}

	if utils.FileExists(m.filepath) && !force {
		fmt.Printf("skipping, manifest %s already exists, use --force to overwrite\n", m.Name)
		return nil
	}

	if err := utils.WriteFile(m.filepath, data); err != nil {
		return err
	}

	return nil
}

func (m *Manifest) ToYAML() ([]byte, error) {
	return yaml.Marshal(m)
}

func (m *Manifest) read() error {
	if !utils.FileExists(m.filepath) {
		return nil
	}

	buff, err := os.ReadFile(m.filepath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(buff, &m); err != nil {
		return err
	}
	return nil
}

func (m *Manifest) Resolve(ctx context.Context) error {
	config, ok := ctx.Value("config").(*config.Config)
	if !ok {
		return fmt.Errorf("config not found")
	}

	for svcName, svc := range m.Services {
		rSvc, err := svc.ResolveType(strings.TrimSuffix(m.filepath, m.Name))
		if err != nil {
			return err
		}
		// Reconcile service
		rSvc.Reconcile(m.Metadata)

		err = rSvc.ResolveTargetPath(svcName, config.Strategy.Template, m.Metadata)
		if err != nil {
			return err
		}

		rSvc.ResolveDependencies(
			m.Name,
			config.Manifests,
			config.Strategy.Template,
			m.Services,
			m.Metadata,
		)

		if err := rSvc.BuildTemplateContext(svcName, m.Metadata); err != nil {
			return err
		}

		m.Services[svcName] = *rSvc
	}
	return nil
}

func (m *Manifest) AddService(name string, svc *catalog.Service) error {
	dest, exists := m.GetService(name)
	if !exists {
		m.Services[name] = *svc
		return nil
	}

	cleaned := CleanService(dest)

	if err := utils.Merge(&cleaned, svc); err != nil {
		return err
	}
	m.Services[name] = *dest
	return nil
}

func (m *Manifest) GetService(name string) (*catalog.Service, bool) {
	if len(m.Services) == 0 {
		m.Services = map[string]catalog.Service{}
		return nil, false
	}
	svc, exists := m.Services[name]
	return &svc, exists
}

func CleanService(in *catalog.Service) catalog.Service {
	return catalog.Service{
		Type:         in.Type,
		Region:       in.Region,
		Scope:        in.Scope,
		Version:      in.Version,
		Inputs:       in.Inputs,
		Labels:       in.Labels,
		Dependencies: in.Dependencies,
	}
}
