package skiff

import "fmt"

type ConfigurationError struct {
	Message string
}

func (e *ConfigurationError) Error() string {
	return fmt.Sprintf("Error occured while loading skiff configuration: reason=%s", e.Message)
}

func NewConfigurationError(message string) *ConfigurationError {
	return &ConfigurationError{Message: message}
}

type ConfigurationNotFoundError struct{}

func (e *ConfigurationNotFoundError) Error() string {
	return fmt.Sprintf("Missing skiff configuration in the context")
}

func NewConfigurationNotFoundError() *ConfigurationNotFoundError {
	return &ConfigurationNotFoundError{}
}
