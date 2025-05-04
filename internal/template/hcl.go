package template

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

func RenderToHCL(data interface{}) string {
	normalizedData := render(data)

	root, ok := normalizedData.(map[string]interface{})
	if !ok {
		panic("rendered data is not a map[string]interface{}")
	}

	hclFile := hclwrite.NewEmptyFile()
	body := hclFile.Body()

	// Separate dependencies
	if dependenciesRaw, ok := root["dependencies"]; ok {
		dependenciesList, ok := dependenciesRaw.([]interface{})
		if !ok {
			panic("dependencies is not a []interface{}")
		}

		WriteDependencyBlocks(body, dependenciesList)

		// Prevent duplicate rendering
		delete(root, "dependencies")
	}

	writeMapToBody(body, root)

	return SanitizeExpressions(string(hclFile.Bytes()))
}

func render(data interface{}) interface{} {
	value := reflect.ValueOf(data)

	switch value.Kind() {
	case reflect.Map:
		result := map[string]interface{}{}
		for _, key := range value.MapKeys() {
			result[fmt.Sprint(key.Interface())] = render(value.MapIndex(key).Interface())
		}
		return result

	case reflect.Slice:
		result := []interface{}{}
		for i := 0; i < value.Len(); i++ {
			result = append(result, render(value.Index(i).Interface()))
		}
		return result

	case reflect.Struct:
		return structToMap(data)

	default:
		return data
	}
}

func structToMap(in interface{}) map[string]interface{} {
	out := make(map[string]interface{})

	val := reflect.ValueOf(in)
	typ := reflect.TypeOf(in)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if !field.CanInterface() {
			continue
		}

		key := fieldType.Name
		if jsonTag := fieldType.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			key = jsonTag
		}

		out[key] = render(field.Interface())
	}

	return out
}

func ctyList(vals []interface{}) cty.Value {
	var out []cty.Value
	for _, v := range vals {
		switch x := v.(type) {
		case string:
			out = append(out, cty.StringVal(x))
		case int:
			out = append(out, cty.NumberIntVal(int64(x)))
		case int64:
			out = append(out, cty.NumberIntVal(x))
		case float64:
			out = append(out, cty.NumberFloatVal(x))
		case bool:
			out = append(out, cty.BoolVal(x))
		}
	}

	if len(out) == 0 {
		// 🛡️ Provide an explicit type (defaulting to string for safety)
		return cty.ListValEmpty(cty.String) // You can change this as needed
	}
	return cty.ListVal(out)
}

func mapToCtyObject(data map[string]interface{}) cty.Value {
	obj := make(map[string]cty.Value)
	for k, v := range data {
		switch x := v.(type) {
		case string:
			obj[k] = cty.StringVal(x)
		case int:
			obj[k] = cty.NumberIntVal(int64(x))
		case int64:
			obj[k] = cty.NumberIntVal(x)
		case float64:
			obj[k] = cty.NumberFloatVal(x)
		case bool:
			obj[k] = cty.BoolVal(x)
		case map[string]interface{}:
			obj[k] = mapToCtyObject(x)
		case []interface{}:
			obj[k] = ctyList(x)
		default:
			fmt.Printf("⚠️ Skipping unsupported nested type %s\n", k)
		}
	}
	return cty.ObjectVal(obj)
}

func ctyListOfObjects(data []map[string]interface{}) cty.Value {
	var list []cty.Value
	for _, item := range data {
		list = append(list, mapToCtyObject(item))
	}
	return cty.ListVal(list)
}

func writeMapToBody(body *hclwrite.Body, data map[string]interface{}) {
	for key, value := range data {
		switch val := value.(type) {
		case string:
			body.SetAttributeValue(key, cty.StringVal(val))
		case int:
			body.SetAttributeValue(key, cty.NumberIntVal(int64(val)))
		case int64:
			body.SetAttributeValue(key, cty.NumberIntVal(val))
		case float64:
			body.SetAttributeValue(key, cty.NumberFloatVal(val))
		case bool:
			body.SetAttributeValue(key, cty.BoolVal(val))
		case []interface{}:
			if len(val) > 0 {
				if _, ok := val[0].(map[string]interface{}); ok {
					var maps []map[string]interface{}
					for _, item := range val {
						if m, ok := item.(map[string]interface{}); ok {
							maps = append(maps, m)
						}
					}
					body.SetAttributeValue(key, ctyListOfObjects(maps))
					continue
				}
			}
			body.SetAttributeValue(key, ctyList(val))
		case map[string]interface{}:
			body.SetAttributeValue(key, mapToCtyObject(val))
		default:
			fmt.Printf("⚠️ Unsupported type for key %s: %T\n", key, val)
		}
	}
}

func WriteDependencyBlocks(body *hclwrite.Body, dependencies []interface{}) {
	for _, raw := range dependencies {
		// Ensure it's a map
		depMap, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract the block label (service)
		serviceName, ok := depMap["service"].(string)
		if !ok {
			continue
		}

		// Copy all other keys except "service"
		blockData := map[string]interface{}{}
		for k, v := range depMap {
			if k != "service" {
				blockData[k] = v
			}
		}

		// Create Terragrunt-style block
		block := body.AppendNewBlock("dependency", []string{serviceName})
		writeMapToBody(block.Body(), blockData)
		body.AppendNewline()
	}
}

func RenderTerraformAttrs(tf map[string]interface{}) string {
	file := hclwrite.NewEmptyFile()
	body := file.Body()
	writeMapToBody(body, tf) // already exists
	return strings.TrimSpace(string(file.Bytes()))
}

// SanitizeExpressions removes "__" prefix and unquotes dependency expressions
// SanitizeExpressions handles expression unquoting and $$ escape fix
func SanitizeExpressions(input string) string {
	// Fix escaped dollar signs (e.g., $${foo} → ${foo})
	input = strings.ReplaceAll(input, "$${", "${")

	// Unquote __expression markers: "__dependency.foo.bar" → dependency.foo.bar
	re := regexp.MustCompile(`= *"__([^"]+)"`)
	input = re.ReplaceAllStringFunc(input, func(match string) string {
		matches := re.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match
		}
		expr := matches[1]
		return "= " + expr
	})

	return input
}
