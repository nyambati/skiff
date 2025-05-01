package skiff

import "fmt"

type ConfigurationError struct {
	Message string
}

func (e *ConfigurationError) Error() string {
	return fmt.Sprintf("Error occured while loading skif configuration: reason=%s", e.Message)
}

func NewConfigurationError(message string) *ConfigurationError {
	return &ConfigurationError{Message: message}
}
