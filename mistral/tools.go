package mistral

import (
	"encoding/json"
	"reflect"
	"strings"
)

type ToolChoice string

const (
	ToolChoiceAuto     ToolChoice = "auto"
	ToolChoiceNone     ToolChoice = "none"
	ToolChoiceAny      ToolChoice = "any"
	ToolChoiceRequired ToolChoice = "required"
)

type ToolType string

const (
	ToolTypeFunction ToolType = "function"
)

// Tool represents a tool that the model can call.
type Tool struct {
	Type     ToolType `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

// NewFunctionTool creates a new function tool.
func NewFunctionTool(name, description string, parameters any) Tool {

	schema, _ := StructToSchema(parameters)
	return Tool{
		Type: ToolTypeFunction,
		Function: Function{
			Name:        name,
			Description: description,
			Parameters:  schema,
		},
	}
}

// StructToSchema generates a JSON Schema from a Go struct using reflection.
// It uses "json" tags for property names and "mistral" tags for descriptions.
func StructToSchema(v any) (map[string]any, error) {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, nil
	}

	properties := make(map[string]any)
	required := []string{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		name := strings.Split(jsonTag, ",")[0]
		omitEmpty := strings.Contains(jsonTag, "omitempty")

		prop := make(map[string]any)
		switch field.Type.Kind() {
		case reflect.String:
			prop["type"] = "string"
		case reflect.Int, reflect.Int32, reflect.Int64:
			prop["type"] = "integer"
		case reflect.Float32, reflect.Float64:
			prop["type"] = "number"
		case reflect.Bool:
			prop["type"] = "boolean"
		case reflect.Slice:
			prop["type"] = "array"
			// basic simplification for slice items
			itemType := "string"
			switch field.Type.Elem().Kind() {
			case reflect.Int, reflect.Int32, reflect.Int64:
				itemType = "integer"
			}
			prop["items"] = map[string]any{"type": itemType}
		default:
			prop["type"] = "object"
		}

		desc := field.Tag.Get("mistral")
		if desc != "" {
			prop["description"] = desc
		}

		properties[name] = prop
		if !omitEmpty {
			required = append(required, name)
		}
	}

	result := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		result["required"] = required
	}

	return result, nil
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     ToolType     `json:"type"`
	Index    *int         `json:"index,omitempty"`
	Function FunctionCall `json:"function"`
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// It ensures that the Type field defaults to "function" if missing.
func (tc *ToolCall) UnmarshalJSON(data []byte) error {
	type Alias ToolCall
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(tc),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if tc.Type == "" {
		tc.Type = ToolTypeFunction
	}
	return nil
}

type FunctionCall struct {
	Arguments string `json:"arguments"`
	Name      string `json:"name"`
}

// UnmarshalArguments unmarshals the tool call arguments into a provided struct.
func UnmarshalArguments[T any](tc ToolCall) (T, error) {
	var v T
	err := json.Unmarshal([]byte(tc.Function.Arguments), &v)
	return v, err
}
