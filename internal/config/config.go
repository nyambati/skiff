package config

import (
	skiff "github.com/nyambati/skiff/internal/errors"
	"github.com/spf13/viper"
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
