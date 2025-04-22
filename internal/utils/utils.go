package utils

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

func CreateDirectory(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error creating directory %s: %s\n", path, err)
		return err
	}
	return nil
}

func WriteFile(path string, content []byte) error {
	if err := os.WriteFile(path, content, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to write file %s: %v\n", path, err)
		return err
	}
	return nil
}

// parseKeyValueFlag parses a comma-separated list of key=value pairs into a map[string]interface{}.
func ParseKeyValueFlag(input string) map[string]any {
	result := make(map[string]any)
	if input == "" {
		return result
	}
	pairs := strings.Split(input, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue // or log warning
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		result[key] = value
	}
	return result
}

func Merge[T any](destination T, source T) error {
	srcRef, _ := getStructReference(source)
	destRef, _ := getStructReference(destination)
	for i := range srcRef.NumField() {
		field := srcRef.Type().Field(i).Name
		if destRef.FieldByName(field).CanSet() && !srcRef.Field(i).IsZero() {
			destRef.FieldByName(field).Set(srcRef.Field(i))
		}
	}
	return nil
}

func getStructReference(i any) (reflect.Value, error) {
	ref := reflect.ValueOf(i)
	if ref.Kind() == reflect.Ptr {
		ref = reflect.Indirect(ref)
	}

	if ref.Kind() == reflect.Interface {
		ref = ref.Elem()
	}

	if ref.Kind() != reflect.Struct {
		return ref, fmt.Errorf("not a struct")
	}

	return ref, nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
