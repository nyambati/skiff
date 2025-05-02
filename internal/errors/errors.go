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

type ConfigNotFoundError struct{}

func (e *ConfigNotFoundError) Error() string {
	return fmt.Sprintf("missing skiff config in context")
}

func NewConfigNotFoundError() *ConfigNotFoundError {
	return &ConfigNotFoundError{}
}
