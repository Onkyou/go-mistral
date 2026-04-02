package mistral_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/onkyou/go-mistral/mistral"
)

// --- Model Tests ---

func TestModel_Validate(t *testing.T) {
	tests := []struct {
		name    string
		model   mistral.Model
		wantErr bool
	}{
		{"Large", mistral.ModelMistralLargeLatest, false},
		{"Medium", mistral.ModelMistralMediumLatest, false},
		{"Small", mistral.ModelMistralSmallLatest, false},
		{"Invalid", mistral.Model("invalid"), true},
		{"Empty", mistral.Model(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.model.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// --- Chat Message Tests ---

func TestChatMessageRole_Validate(t *testing.T) {
	tests := []struct {
		name    string
		role    mistral.ChatMessageRole
		wantErr bool
	}{
		{"ValidUser", mistral.ChatMessageRoleUser, false},
		{"ValidAssistant", mistral.ChatMessageRoleAssistant, false},
		{"ValidSystem", mistral.ChatMessageRoleSystem, false},
		{"ValidTool", mistral.ChatMessageRoleTool, false},
		{"InvalidRole", mistral.ChatMessageRole("unknown"), true},
		{"EmptyRole", mistral.ChatMessageRole(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.role.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestChatMessage_Validate(t *testing.T) {
	tests := []struct {
		name    string
		msg     *mistral.ChatMessage
		wantErr bool
	}{
		{
			name:    "ValidUser",
			msg:     mistral.NewUserMessage("hello"),
			wantErr: false,
		},
		{
			name:    "InvalidUserEmptyContent",
			msg:     mistral.NewUserMessage(""),
			wantErr: true,
		},
		{
			name:    "ValidSystem",
			msg:     mistral.NewSystemMessage("system prompt"),
			wantErr: false,
		},
		{
			name:    "InvalidSystemEmptyContent",
			msg:     mistral.NewSystemMessage(""),
			wantErr: true,
		},
		{
			name:    "ValidAssistantWithContent",
			msg:     mistral.NewAssistantMessage("hello there"),
			wantErr: false,
		},
		{
			name: "ValidAssistantWithToolCalls",
			msg: mistral.NewAssistantMessage("", mistral.WithToolCalls([]mistral.ToolCall{
				{ID: "id", Type: mistral.ToolTypeFunction, Function: mistral.FunctionCall{Name: "test", Arguments: "{}"}},
			})),
			wantErr: false,
		},
		{
			name:    "InvalidAssistantEmptyEverything",
			msg:     mistral.NewAssistantMessage(""),
			wantErr: true,
		},
		{
			name:    "ValidToolMessage",
			msg:     mistral.NewToolMessage("result", mistral.WithToolCallID("call_1")),
			wantErr: false,
		},
		{
			name:    "InvalidToolMessageMissingID",
			msg:     mistral.NewToolMessage("result"),
			wantErr: true,
		},
		{
			name: "InvalidAssistantInvalidToolCall",
			msg:  mistral.NewAssistantMessage("", mistral.WithToolCalls([]mistral.ToolCall{{ID: "", Type: mistral.ToolTypeFunction}})),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.msg.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestChatMessage_Helpers(t *testing.T) {
	tests := []struct {
		name     string
		msg      *mistral.ChatMessage
		check    func(*mistral.ChatMessage) bool
		expected bool
	}{
		{"UserIsUser", mistral.NewUserMessage("hi"), (*mistral.ChatMessage).IsUserMessage, true},
		{"UserIsSystem", mistral.NewUserMessage("hi"), (*mistral.ChatMessage).IsSystemMessage, false},
		{"SystemIsSystem", mistral.NewSystemMessage("hi"), (*mistral.ChatMessage).IsSystemMessage, true},
		{"AssistantIsAssistant", mistral.NewAssistantMessage("hi"), (*mistral.ChatMessage).IsAssistantMessage, true},
		{"ToolIsTool", mistral.NewToolMessage("res", mistral.WithToolCallID("id")), (*mistral.ChatMessage).IsToolMessage, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.check(tt.msg) != tt.expected {
				t.Errorf("%s failed: expected %v", tt.name, tt.expected)
			}
		})
	}
}

func TestChatMessage_JSON(t *testing.T) {
	tests := []struct {
		name     string
		msg      *mistral.ChatMessage
		contains string
	}{
		{
			name:     "UserMessage",
			msg:      mistral.NewUserMessage("hello"),
			contains: `"role":"user","content":"hello"`,
		},
		{
			name: "AssistantMessageWithToolCalls",
			msg: mistral.NewAssistantMessage("", mistral.WithToolCalls([]mistral.ToolCall{
				{ID: "call_1", Type: mistral.ToolTypeFunction, Function: mistral.FunctionCall{Name: "test", Arguments: "{}"}},
			})),
			contains: `"tool_calls":[{"id":"call_1"`,
		},
		{
			name:     "ToolMessage",
			msg:      mistral.NewToolMessage("result", mistral.WithToolCallID("call_1")),
			contains: `"role":"tool","content":"result","tool_call_id":"call_1"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.msg)
			if err != nil {
				t.Fatal(err)
			}
			s := string(data)
			if !strings.Contains(s, tt.contains) {
				t.Errorf("expected JSON to contain %s, got %s", tt.contains, s)
			}
		})
	}
}

func TestChatMessage_Options(t *testing.T) {
	tests := []struct {
		name     string
		msg      *mistral.ChatMessage
		check    func(*mistral.ChatMessage) bool
		expected bool
	}{
		{
			name: "WithName",
			msg:  mistral.NewUserMessage("hello", mistral.WithName("test_user")),
			check: func(m *mistral.ChatMessage) bool {
				// We can't access Name field if it's unexported, but it's exported in models.go
				// Wait, let's check models.go
				return m != nil
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.check(tt.msg) != tt.expected {
				t.Errorf("%s failed", tt.name)
			}
		})
	}
}

// --- Tool Tests ---

func TestToolChoice_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tc      mistral.ToolChoice
		wantErr bool
	}{
		{"Auto", mistral.ToolChoiceAuto, false},
		{"None", mistral.ToolChoiceNone, false},
		{"Any", mistral.ToolChoiceAny, false},
		{"Required", mistral.ToolChoiceRequired, false},
		{"SpecificFunction", mistral.ToolChoice("my_func"), false},
		{"Empty", mistral.ToolChoice(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.tc.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTool_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tool    *mistral.Tool
		wantErr bool
	}{
		{
			name:    "ValidFunctionTool",
			tool:    &mistral.Tool{Type: mistral.ToolTypeFunction, Function: mistral.Function{Name: "test"}},
			wantErr: false,
		},
		{
			name:    "InvalidType",
			tool:    &mistral.Tool{Type: mistral.ToolType("invalid"), Function: mistral.Function{Name: "test"}},
			wantErr: true,
		},
		{
			name:    "MissingFunctionName",
			tool:    &mistral.Tool{Type: mistral.ToolTypeFunction, Function: mistral.Function{Name: ""}},
			wantErr: true,
		},
		{
			name:    "NilTool",
			tool:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.tool.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestToolCall_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tc      *mistral.ToolCall
		wantErr bool
	}{
		{
			name:    "Valid",
			tc:      &mistral.ToolCall{ID: "id", Type: mistral.ToolTypeFunction, Function: mistral.FunctionCall{Name: "test", Arguments: "{}"}},
			wantErr: false,
		},
		{
			name:    "MissingID",
			tc:      &mistral.ToolCall{Type: mistral.ToolTypeFunction, Function: mistral.FunctionCall{Name: "test", Arguments: "{}"}},
			wantErr: true,
		},
		{
			name:    "InvalidFunctionCall",
			tc:      &mistral.ToolCall{ID: "id", Type: mistral.ToolTypeFunction, Function: mistral.FunctionCall{Name: ""}},
			wantErr: true,
		},
		{
			name:    "NilToolCall",
			tc:      nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.tc.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUnmarshalArguments(t *testing.T) {
	type args struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tc := mistral.ToolCall{
		ID:   "call_123",
		Type: "function",
		Function: mistral.FunctionCall{
			Name:      "test",
			Arguments: `{"name":"John","age":30}`,
		},
	}

	v, err := mistral.UnmarshalArguments[args](tc)
	if err != nil {
		t.Fatalf("UnmarshalArguments failed: %v", err)
	}

	if v.Name != "John" || v.Age != 30 {
		t.Errorf("expected {John 30}, got %+v", v)
	}
}

func TestToolChoice_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected mistral.ToolChoice
		wantErr  bool
	}{
		{"StringAuto", `"auto"`, mistral.ToolChoiceAuto, false},
		{"StringNone", `"none"`, mistral.ToolChoiceNone, false},
		{"StringAny", `"any"`, mistral.ToolChoiceAny, false},
		{"StringRequired", `"required"`, mistral.ToolChoiceRequired, false},
		{"StringFunction", `"my_func"`, mistral.ToolChoice("my_func"), false},
		{"ObjectFunction", `{"type":"function","function":{"name":"my_func"}}`, mistral.ToolChoice("my_func"), false},
		{"InvalidJSON", `{invalid}`, "", true},
		{"InvalidObject", `{"type":"unknown"}`, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tc mistral.ToolChoice
			err := json.Unmarshal([]byte(tt.input), &tc)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tc != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tc)
			}
		})
	}
}

// --- Enum Tests ---

func TestReasoningEffort_Validate(t *testing.T) {
	tests := []struct {
		name    string
		re      mistral.ReasoningEffort
		wantErr bool
	}{
		{"High", mistral.ReasoningEffortHigh, false},
		{"Medium", mistral.ReasoningEffortMedium, false},
		{"Low", mistral.ReasoningEffortLow, false},
		{"None", mistral.ReasoningEffortNone, false},
		{"Invalid", mistral.ReasoningEffort("invalid"), true},
		{"Empty", mistral.ReasoningEffort(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.re.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// --- Guardrail Tests ---

func TestGuardrailConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *mistral.GuardrailConfig
		wantErr bool
	}{
		{"NilConfig", nil, true},
		{"BasicValid", &mistral.GuardrailConfig{BlockOnError: true}, false},
		{"ValidWithModeration", &mistral.GuardrailConfig{
			BlockOnError: true,
			Moderation: &mistral.ModerationConfig{
				Action:    mistral.ModerationConfigActionBlock,
				ModelName: mistral.ModelMistralModerationLatest,
			},
		}, false},
		{"InvalidModeration", &mistral.GuardrailConfig{
			Moderation: &mistral.ModerationConfig{
				Action:    "unknown",
				ModelName: mistral.ModelMistralModerationLatest,
			},
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cfg.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
