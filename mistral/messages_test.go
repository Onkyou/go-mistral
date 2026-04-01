package mistral

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestChatMessage_Helpers(t *testing.T) {
	tests := []struct {
		name     string
		msg      ChatMessage
		check    func(ChatMessage) bool
		expected bool
	}{
		{"UserIsUser", NewUserMessage("hi"), ChatMessage.IsUserMessage, true},
		{"UserIsSystem", NewUserMessage("hi"), ChatMessage.IsSystemMessage, false},
		{"SystemIsSystem", NewSystemMessage("hi"), ChatMessage.IsSystemMessage, true},
		{"AssistantIsAssistant", NewAssistantMessage("hi"), ChatMessage.IsAssistantMessage, true},
		{"ToolIsTool", NewToolMessage("res", WithToolCallID("id")), ChatMessage.IsToolMessage, true},
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
		msg      ChatMessage
		contains string
	}{
		{
			name:     "UserMessage",
			msg:      NewUserMessage("hello"),
			contains: `"role":"user","content":"hello"`,
		},
		{
			name: "AssistantMessageWithToolCalls",
			msg: NewAssistantMessage("", WithToolCalls([]ToolCall{
				{ID: "call_1", Type: "function", Function: FunctionCall{Name: "test", Arguments: "{}"}},
			})),
			contains: `"tool_calls":[{"id":"call_1"`,
		},
		{
			name:     "ToolMessage",
			msg:      NewToolMessage("result", WithToolCallID("call_1")),
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
