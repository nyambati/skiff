package config

type (
	Path struct {
		Manifests  string `yaml:"manifests"`
		Templates  string `yaml:"templates"`
		Terragrunt string `yaml:"terragrunt"`
	}

	SkiffConfig struct {
		Version  string `yaml:"version"`
		Verbose  bool   `yaml:"verbose"`
		Strategy string `yaml:"strategy"`
		Path     `yaml:"path"`
	}

	InitConfig struct {
		Folder       string
		Template     []byte
		TemplateName string
	}
)
