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

type ServiceTypeDoesNotExistError struct {
	Type string
}

func (e *ServiceTypeDoesNotExistError) Error() string {
	return fmt.Sprintf("service type %s does not exist in the catalog, please check your catalog configuration", e.Type)
}

func NewServiceTypeDoesNotExistError(serviceType string) *ServiceTypeDoesNotExistError {
	return &ServiceTypeDoesNotExistError{Type: serviceType}
}
