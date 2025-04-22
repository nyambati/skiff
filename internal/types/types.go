package types

type (
	SkiffConfig struct {
		Path struct {
			Base       string `yaml:"base"`
			Manifests  string `yaml:"manifests"`
			Templates  string `yaml:"templates"`
			Terragrunt string `yaml:"terragrunt"`
		} `yaml:"path"`
	}
)
