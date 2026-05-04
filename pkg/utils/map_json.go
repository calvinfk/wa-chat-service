package utils

import (
	"encoding/json"
	"reflect"
	"strings"
)

// StructToMap converts a struct to a map[string]any by first marshaling the struct to JSON and then unmarshaling it back into a map.
func StructToMap(v any, omitNull bool) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	err = json.Unmarshal(b, &result)
	// if the value is null
	if omitNull {
		omitNullValues(result)
	}
	return result, err
}

func omitNullValues(m map[string]any) map[string]any {
	// recursively omit null values from the map
	for k, v := range m {
		switch val := v.(type) {
		case map[string]any:
			m[k] = omitNullValues(val)
			if len(m[k].(map[string]any)) == 0 {
				delete(m, k)
			}
		case []any:
			for i, item := range val {
				if itemMap, ok := item.(map[string]any); ok {
					val[i] = omitNullValues(itemMap)
					if len(val[i].(map[string]any)) == 0 {
						val = append(val[:i], val[i+1:]...)
					}
				} else if item == nil {
					val = append(val[:i], val[i+1:]...)
				}
			}
		case nil:
			delete(m, k)
			continue
		}
	}
	return m
}

func MapToStruct(m map[string]any, v any) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func AnyToJsonString(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// getJsonName retrieves the JSON field name for a given struct field, falling back to the lowercase of the field name if no JSON tag is present.
func getJsonName(entity any, fieldName string) string {
	t := reflect.TypeOf(entity)

	// If it's a pointer, get the element
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	// Handle the case where we are validating the wrapper struct
	field, found := t.FieldByName(fieldName)
	if !found {
		return strings.ToLower(fieldName) // Fallback
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag == "" || jsonTag == "-" {
		return strings.ToLower(fieldName)
	}

	// Return only the name part of the tag (before any commas)
	return strings.Split(jsonTag, ",")[0]
}

func JsonStringToMap(jsonString string) (map[string]any, error) {
	var result map[string]any
	err := json.Unmarshal([]byte(jsonString), &result)
	return result, err
}
