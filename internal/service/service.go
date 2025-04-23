package service

import (
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
