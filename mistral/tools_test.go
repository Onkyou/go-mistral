package mistral

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTool_JSON(t *testing.T) {
	tool := Tool{
		Type: ToolTypeFunction,
		Function: Function{
			Name:        "test",
			Description: "desc",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"arg1": map[string]any{"type": "string"},
				},
			},
		},
	}

	data, err := json.Marshal(tool)
	if err != nil {
		t.Fatal(err)
	}

	expected := `"type":"function","function":{"name":"test","description":"desc"`
	if !strings.Contains(string(data), expected) {
		t.Errorf("expected JSON to contain %s, got %s", expected, string(data))
	}
}

func TestUnmarshalArguments(t *testing.T) {
	type args struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tc := ToolCall{
		ID:   "call_123",
		Type: "function",
		Function: FunctionCall{
			Name:      "test",
			Arguments: `{"name":"John","age":30}`,
		},
	}

	v, err := UnmarshalArguments[args](tc)
	if err != nil {
		t.Fatalf("UnmarshalArguments failed: %v", err)
	}

	if v.Name != "John" || v.Age != 30 {
		t.Errorf("expected {John 30}, got %+v", v)
	}
}
func TestStructToSchema(t *testing.T) {
	type testArgs struct {
		Name     string `json:"name" mistral:"the name"`
		Age      int    `json:"age,omitempty"`
		IsActive bool   `json:"is_active" mistral:"status"`
	}

	schema, err := StructToSchema(testArgs{})
	if err != nil {
		t.Fatal(err)
	}

	if schema["type"] != "object" {
		t.Errorf("expected type object, got %v", schema["type"])
	}

	props := schema["properties"].(map[string]any)
	if props["name"].(map[string]any)["type"] != "string" {
		t.Errorf("expected name to be string")
	}
	if props["name"].(map[string]any)["description"] != "the name" {
		t.Errorf("expected name description to be 'the name'")
	}

	required := schema["required"].([]string)
	if len(required) != 2 { // Name and IsActive are required, Age is omitempty
		t.Errorf("expected 2 required fields, got %v", len(required))
	}
}
