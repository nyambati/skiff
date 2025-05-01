package config

type (
	Path struct {
		Manifests  string `yaml:"manifests"`
		Templates  string `yaml:"templates"`
		Terragrunt string `yaml:"terragrunt"`
		Strategies string `yaml:"strategies"`
	}

	Strategy struct {
		Description string `yaml:"description"`
		Template    string `yaml:"template"`
	}

	Config struct {
		Version  string   `yaml:"version"`
		Verbose  bool     `yaml:"verbose"`
		Strategy Strategy `yaml:"strategy"`
		Path     `yaml:"path"`
	}

	InitConfig struct {
		Folder       string
		Template     []byte
		TemplateName string
	}
)
