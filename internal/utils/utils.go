package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
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
		if key == "" || value == "" {
			continue
		}
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

func SanitizePath(input string) string {
	// Normalize line endings and split into path parts
	lines := strings.FieldsFunc(input, func(r rune) bool {
		return r == '\n' || r == '\r' || r == '\t'
	})
	// Split again by "/" and trim all fragments
	var cleanParts []string
	for _, line := range lines {
		parts := strings.Split(line, "/")
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				cleanParts = append(cleanParts, trimmed)
			}
		}
	}
	// Rejoin into clean slash-delimited path
	return strings.Join(cleanParts, "/")
}

func findEditor() (editor string, args []string) {
	if editor := os.Getenv("VISUAL"); editor != "" {
		parts := strings.Fields(editor)
		return parts[0], parts[1:]
	}

	if editor := os.Getenv("EDITOR"); editor != "" {
		parts := strings.Fields(editor)
		return parts[0], parts[1:]
	}

	// OS-specific fallbacks
	switch runtime.GOOS {
	case "windows":
		return "notepad", nil
	case "darwin":
		// On macOS, 'open -t' opens with the default text editor,
		// or 'vi'/'nano' are common command-line fallbacks.
		// Let's prefer 'vi' as a common default if 'open' isn't desired.
		return "vi", nil // or "nano", or check for TextEdit via 'open -a TextEdit'
	default: // linux, bsd, etc.
		return "vi", nil // or "nano"
	}
}

// editFileContent opens the content of the given file path in the user's
// preferred editor and returns the potentially modified content as a byte slice.
// It uses a temporary file to avoid modifying the original directly during editing.
func EditFile(filePath string, existingContent []byte) ([]byte, error) {
	editor, eArgs := findEditor()
	if editor == "" {
		return nil, fmt.Errorf("no suitable editor found (checked VISUAL, EDITOR env vars)")
	}

	// 2. Create a temporary file
	// Use the original filename's extension for syntax highlighting if possible.
	pattern := "edit-*" + filepath.Ext(filePath)
	tempFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	defer os.Remove(tempFile.Name()) // Ensure cleanup

	// 3. Write original content to the temporary file
	if _, err := tempFile.Write(existingContent); err != nil {
		tempFile.Close() // Close before attempting remove on error path
		return nil, fmt.Errorf("failed to write to temporary file '%s': %w", tempFile.Name(), err)
	}

	// 4. Close the file handle before launching the editor
	if err := tempFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temporary file '%s': %w", tempFile.Name(), err)
	}

	args := []string{tempFile.Name()}

	if eArgs != nil && len(eArgs) > 0 {
		args = append(args, eArgs...)
	}

	// 5. Prepare and run the editor command
	cmd := exec.Command(strings.TrimSpace(editor), args...)
	// Connect the editor to the Go program's stdin, stdout, and stderr
	// This is crucial for interactive use in the same terminal.
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// This error means the editor command itself failed (e.g., editor not found,
		// or the editor exited with a non-zero status).
		return nil, fmt.Errorf("editor command '%s %s' failed: %w", editor, tempFile.Name(), err)
	}

	// 6. Read the content from the temporary file AFTER the editor has closed
	editedContent, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read edited content from temporary file '%s': %w", tempFile.Name(), err)
	}

	// 7. Temporary file is removed by defer. Return the content.
	return editedContent, nil
}

func ToYAML(i any) ([]byte, error) {
	data, err := yaml.Marshal(i)
	if err != nil {
		return nil, err
	}
	return data, nil
}
