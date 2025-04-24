package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
)

// CreateDirectory creates the given directory path with the given permissions.
// If the directory already exists, nothing is done and nil is returned.
// If the directory does not exist, it is created.
// If there is an error creating the directory, an error is returned.
// The permissions are set to 0755, which is the default for mkdir.
func CreateDirectory(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error creating directory %s: %s\n", path, err)
		return err
	}
	return nil
}

// WriteFile writes the given content to the given file path.
// If the file already exists, it is overwritten.
// If the file does not exist, it is created.
// The function returns an error if there is a problem writing the file.
func WriteFile(path string, content []byte) error {
	if err := os.WriteFile(path, content, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to write file %s: %v\n", path, err)
		return err
	}
	return nil
}

// ParseKeyValueFlag parses a comma-separated string of key-value pairs into a map.
// The string is split by commas, and then each pair is split by the first equals sign.
// Leading and trailing whitespace on the keys and values is trimmed.
// If a pair does not have a key or value, it is skipped.
// The function returns the resulting map.
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

// Merge takes two structs of the same type and merges the source into the destination, overwriting any non-zero fields in the destination with the value from the source.
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

// getStructReference takes an arbitrary input and returns the underlying struct reference
// and error. It dereferences pointers, extracts the value from interfaces, and
// returns an error if the underlying type is not a struct.
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

// FileExists returns true if the file at path exists and false otherwise.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// PrintErrorAndExit prints the given error to stderr and exits with status code 1.
func PrintErrorAndExit(err error) {
	fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
	os.Exit(1)
}

func HasLabels(source, labels map[string]any) bool {
	for key, val := range labels {
		if source[key] != val {
			return false
		}
	}
	return true
}

func ToMap[T any](input T) (map[string]any, error) {
	data, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	var output map[string]any
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, err
	}
	return output, nil
}
