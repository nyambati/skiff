package manifest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nyambati/skiff/internal/catalog"
	"github.com/nyambati/skiff/internal/config"
	skiff "github.com/nyambati/skiff/internal/errors"
	"github.com/nyambati/skiff/internal/types"
	"github.com/nyambati/skiff/internal/utils"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func Read(ctx context.Context, manifestName string) (*Manifest, error) {
	cfg, err := config.FromContext(ctx)
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

func (m *Manifest) Write(force bool) error {
	data, err := m.ToYAML()
	if err != nil {
		return err
	}

	if utils.FileExists(m.filepath) && !force {
		log.Infof("skipping, manifest %s already exists, use --force to overwrite\n", m.Name)
		return nil
	}

	data = utils.PrependWatermark(string(data), config.ToolName)
	fmt.Println("writing to ", m.filepath)
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

func EditManifest(ctx context.Context, name, metadata string) error {

	manifest, err := Read(ctx, name)
	if err != nil {
		return err
	}

	manifestPath := manifest.filepath

	oldManifest, err := utils.ToYAML(manifest)
	if err != nil {
		return err
	}

	if metadata != "" {
		metadata := utils.ParseKeyValueFlag(metadata)

		for k, v := range metadata {
			manifest.Metadata[strings.ToLower(k)] = v
		}

	}

	content, err := utils.ToYAML(manifest)
	if err != nil {
		return err
	}

	content, err = utils.EditFile(manifestPath, content)
	if err != nil {
		return err
	}

	manifest, err = utils.FromYAML[Manifest](content)
	if err != nil {
		return err
	}

	manifest.filepath = manifestPath

	if !utils.ShouldWrite(oldManifest, content) {
		return nil
	}

	return manifest.Write(true)
}

func AddService(ctx context.Context, manifestName, serviceName string) error {
	var svcCatalog catalog.Catalog

	cfg, err := config.FromContext(ctx)
	if err != nil {
		return err
	}

	catalogFilePath := fmt.Sprintf("%s/%s", cfg.Manifests, config.CatalogFile)
	manifestFilePath := fmt.Sprintf("%s/%s.yaml", cfg.Manifests, manifestName)

	manifest, err := Read(ctx, manifestName)
	if err != nil {
		return err
	}

	if err := svcCatalog.Read(catalogFilePath); err != nil {
		return err
	}

	svc, ok := manifest.GetService(serviceName)
	if !ok {
		svc = catalog.DefaultService(serviceName, "")
	}

	oldContent, err := utils.ToYAML(svc)
	if err != nil {
		return err
	}

	newContent, err := utils.EditFile(manifestFilePath, oldContent)
	if err != nil {
		return err
	}

	svc, err = utils.FromYAML[catalog.Service](newContent)
	if err != nil {
		return err
	}

	if _, exists := svcCatalog.GetServiceType(svc.Type); !exists {
		return skiff.NewServiceTypeDoesNotExistError(svc.Type)
	}

	if !utils.ShouldWrite(oldContent, newContent) {
		return nil
	}

	manifest.AddService(serviceName, svc)

	if err := manifest.Write(true); err != nil {
		return err
	}
	logrus.Infof("✅ Service %s has been added to %s\n", serviceName, manifestFilePath)
	return nil
}
