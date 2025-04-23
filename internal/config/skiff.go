package config

import (
	"fmt"

	"github.com/spf13/viper"
)

var Config *SkiffConfig

func Load(command string) error {
	viper.AddConfigPath(".")
	viper.SetConfigName(".skiff")
	viper.SetConfigType("yaml")

	if command == "init" {
		Config = new(SkiffConfig)
		return nil
	}

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if err := viper.Unmarshal(&Config); err != nil {
		return fmt.Errorf("Error unmarshalling config: %w", err)
	}
	return nil
}
