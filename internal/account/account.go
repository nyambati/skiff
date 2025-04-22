package account

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nyambati/skiff/internal/service"
	"github.com/nyambati/skiff/internal/utils"
	"gopkg.in/yaml.v2"
)

func New(version, name, id string) *Manifest {
	return &Manifest{
		APIVersion: "v1",
		Account: Account{
			Name: name,
			ID:   id,
		},
	}
}

func (m *Manifest) Write(path string, verbose, force bool) error {
	data, err := m.ToYAML()
	if err != nil {
		return err
	}

	path = filepath.Join(path, fmt.Sprintf("%s.yaml", m.Account.ID))

	if err := utils.WriteFile(path, data); err != nil {
		return err
	}
	return nil

}

func (m *Manifest) ToYAML() ([]byte, error) {
	return yaml.Marshal(m)
}

func (m *Manifest) Read(path, accountID string) error {
	buff, err := os.ReadFile(fmt.Sprintf("%s/%s.yaml", path, accountID))
	if err != nil {
		return err
	}
	return yaml.Unmarshal(buff, &m)
}

func (m *Manifest) AddService(name string, svc *service.Service) error {
	dest, exists := m.GetService(name)
	if !exists {
		m.Services[name] = *svc
	}

	if err := utils.Merge(dest, svc); err != nil {
		return err
	}

	m.Services[name] = *dest
	return nil
}

func (m *Manifest) GetService(name string) (*service.Service, bool) {
	if len(m.Services) == 0 {
		m.Services = map[string]service.Service{}
		return nil, false
	}
	svc, exists := m.Services[name]
	return &svc, exists
}
