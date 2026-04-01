package mistral

// ChatMessage represents a single message in a conversation.
type ChatMessage struct {
	Role    Role    `json:"role"`
	Content string  `json:"content"`
	Name    *string `json:"name,omitempty"`

	// ToolCalls is only used when Role is RoleAssistant.
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// ToolCallID is only used when Role is RoleTool.
	ToolCallID *string `json:"tool_call_id,omitempty"`

	// Prefix is only used when Role is RoleAssistant for model conditioning.
	Prefix *bool `json:"prefix,omitempty"`
}

// IsUserMessage returns true if the message is from a user.
func (m ChatMessage) IsUserMessage() bool {
	return m.Role == RoleUser
}

// IsAssistantMessage returns true if the message is from the assistant.
func (m ChatMessage) IsAssistantMessage() bool {
	return m.Role == RoleAssistant
}

// IsSystemMessage returns true if the message is a system message.
func (m ChatMessage) IsSystemMessage() bool {
	return m.Role == RoleSystem
}

// IsToolMessage returns true if the message is from a tool.
func (m ChatMessage) IsToolMessage() bool {
	return m.Role == RoleTool
}

// MessageOption is a functional option for ChatMessage.
type MessageOption func(*ChatMessage)

// WithName sets the name for the message.
func WithName(name string) MessageOption {
	return func(m *ChatMessage) {
		m.Name = &name
	}
}

// WithToolCalls sets the tool calls for an assistant message.
func WithToolCalls(calls []ToolCall) MessageOption {
	return func(m *ChatMessage) {
		if len(calls) == 0 {
			m.ToolCalls = nil
		} else {
			m.ToolCalls = calls
		}
	}
}

// WithToolCallID sets the tool call ID for a tool message.
func WithToolCallID(id string) MessageOption {
	return func(m *ChatMessage) {
		m.ToolCallID = &id
	}
}

// WithPrefix sets the prefix flag for an assistant message.
func WithPrefix(prefix bool) MessageOption {
	return func(m *ChatMessage) {
		m.Prefix = &prefix
	}
}

// NewUserMessage creates a new message from a user.
func NewUserMessage(content string, opts ...MessageOption) ChatMessage {
	m := ChatMessage{
		Role:    RoleUser,
		Content: content,
	}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

// NewSystemMessage creates a new system message.
func NewSystemMessage(content string, opts ...MessageOption) ChatMessage {
	m := ChatMessage{
		Role:    RoleSystem,
		Content: content,
	}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

// NewAssistantMessage creates a new message from the assistant.
func NewAssistantMessage(content string, opts ...MessageOption) ChatMessage {
	m := ChatMessage{
		Role:    RoleAssistant,
		Content: content,
	}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

// NewToolMessage creates a new message from a tool.
func NewToolMessage(content string, opts ...MessageOption) ChatMessage {
	m := ChatMessage{
		Role:    RoleTool,
		Content: content,
	}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}
