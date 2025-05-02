package manifest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nyambati/skiff/internal/catalog"
	"github.com/nyambati/skiff/internal/config"
	"github.com/nyambati/skiff/internal/types"
	"github.com/nyambati/skiff/internal/utils"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func Read(ctx context.Context, manifestName string) (*Manifest, error) {
	cfg, err := utils.GetConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	m := &Manifest{
		APIVersion: "v1",
		Name:       manifestName,
		Metadata:   types.Metadata{config.NameKey: manifestName},
		filepath:   filepath.Join(cfg.Manifests, fmt.Sprintf("%s.yaml", manifestName)),
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
		log.Infof("skipping, manifest %s already exists, use --force to overwrite\n", m.Name)
		return nil
	}

	data = utils.PrependWatermark(string(data), config.ToolName)

	if err := utils.WriteFile(m.filepath, data); err != nil {
		return err
	}

	log.Infof(" ✅ manifest %s has been updated successfully\n", m.Name)
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
	for svcName, svc := range m.Services {
		rSvc, err := svc.ResolveType(ctx)
		if err != nil {
			return err
		}
		// Reconcile service
		rSvc.Reconcile(m.Metadata)

		err = rSvc.ResolveTargetPath(ctx, svcName, m.Metadata)
		if err != nil {
			return err
		}

		rSvc.ResolveDependencies(
			ctx,
			m.Name,
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

	if err := utils.Merge(dest, svc, false); err != nil {
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
