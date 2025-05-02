package config

import (
	skiff "github.com/nyambati/skiff/internal/errors"
	"github.com/spf13/viper"
)

const (
	ToolName               = "skiff"
	CatalogFile            = "catalog.yaml"
	TerragruntTemplateFile = "terragrunt.default.tmpl"
	SkiffConfigFile        = ".skiff"
	ScopeRegional          = "regional"
	ScopeGlobal            = "global"
	ServiceKey             = "service"
	GroupKey               = "group"
	ContextKey             = "config"
	RegionKey              = "region"
	TypeKey                = "type"
	TagsKey                = "tags"
	ScopeKey               = "scope"
	VersionKey             = "version"
	InputsKey              = "inputs"
	DependencyKey          = "dependencies"
	NameKey                = "name"
)

func New(command string) (*Config, error) {
	config := new(Config)

	if command == "init" {
		return config, nil
	}

	viper.AddConfigPath(".")
	viper.SetConfigName(".skiff")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, skiff.NewConfigurationError("failed to unmarshall config")
	}

	return config, nil
}
