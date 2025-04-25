package service

import (
	"fmt"
	"os"

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
