package mistral

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// --- Models ---

type Model string

const (
	ModelMistralLargeLatest    Model = "mistral-large-latest"
	ModelMistralMediumLatest   Model = "mistral-medium-latest"
	ModelMistralSmallLatest    Model = "mistral-small-latest"
	ModelCodestralLatest       Model = "codestral-latest"
	ModelOpenMixtral8x7b       Model = "open-mixtral-8x7b"
	ModelOpenMixtral8x22b      Model = "open-mixtral-8x22b"
	ModelOpenMistral7b         Model = "open-mistral-7b"
	ModelMistralLarge2402      Model = "mistral-large-2402"
	ModelMistralMedium2312     Model = "mistral-medium-2312"
	ModelMistralSmall2402      Model = "mistral-small-2402"
	ModelMistralSmall2312      Model = "mistral-small-2312"
	ModelMistralTiny           Model = "mistral-tiny-2312"
	ModelMistralEmbed          Model = "mistral-embed"
	ModelMistralModeration2603 Model = "mistral-moderation-2603"
	ModelMistralModerationLatest Model = "mistral-moderation-latest"
)

func (m *Model) Validate() error {
	if m == nil {
		return fmt.Errorf("model is nil")
	}
	switch *m {
	case ModelMistralLargeLatest, ModelMistralMediumLatest, ModelMistralSmallLatest, ModelCodestralLatest, ModelOpenMixtral8x7b, ModelOpenMixtral8x22b, ModelOpenMistral7b, ModelMistralLarge2402, ModelMistralMedium2312, ModelMistralSmall2402, ModelMistralSmall2312, ModelMistralTiny, ModelMistralEmbed, ModelMistralModeration2603, ModelMistralModerationLatest:
		return nil
	default:
		return fmt.Errorf("invalid model: %s", *m)
	}
}

// --- Usage ---

type UsageInfo struct {
	CompletionTokens    int                  `json:"completion_tokens"`
	NumCachedTokens     *int                 `json:"num_cached_tokens,omitempty"`
	PromptAudioSeconds  *int                 `json:"prompt_audio_seconds,omitempty"`
	PromptTokenDetails  *PromptTokenDetails  `json:"prompt_token_details,omitempty"`
	PromptTokens        int                  `json:"prompt_tokens"`
	PromptTokensDetails *PromptTokensDetails `json:"prompt_tokens_details,omitempty"`
	TotalTokens         int                  `json:"total_tokens"`
}

type PromptTokenDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

type PromptTokensDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

// --- Chat Messages ---

type ChatMessageRole string

const (
	ChatMessageRoleUser      ChatMessageRole = "user"
	ChatMessageRoleAssistant ChatMessageRole = "assistant"
	ChatMessageRoleSystem    ChatMessageRole = "system"
	ChatMessageRoleTool      ChatMessageRole = "tool"
)

func (r *ChatMessageRole) Validate() error {
	if r == nil {
		return fmt.Errorf("chat message role is nil")
	}
	switch *r {
	case ChatMessageRoleUser, ChatMessageRoleAssistant, ChatMessageRoleSystem, ChatMessageRoleTool:
		return nil
	default:
		return fmt.Errorf("invalid chat message role: %s", *r)
	}
}

// ChatMessage represents a single message in a conversation.
type ChatMessage struct {
	Role    ChatMessageRole `json:"role"`
	Content string          `json:"content"`
	Name    *string         `json:"name,omitempty"`

	// ToolCalls is only used when Role is ChatMessageRoleAssistant.
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// ToolCallID is only used when Role is ChatMessageRoleTool.
	ToolCallID *string `json:"tool_call_id,omitempty"`

	// Prefix is only used when Role is ChatMessageRoleAssistant for model conditioning.
	Prefix *bool `json:"prefix,omitempty"`
}

func (m *ChatMessage) Validate() error {
	if m == nil {
		return fmt.Errorf("chat message is nil")
	}

	if err := m.Role.Validate(); err != nil {
		return err
	}

	switch m.Role {
	case ChatMessageRoleUser, ChatMessageRoleSystem:
		if m.Content == "" {
			return fmt.Errorf("%s message must have content", m.Role)
		}
	case ChatMessageRoleAssistant:
		if m.Content == "" && len(m.ToolCalls) == 0 {
			return fmt.Errorf("assistant message must have content or tool calls")
		}
	case ChatMessageRoleTool:
		if m.ToolCallID == nil || *m.ToolCallID == "" {
			return fmt.Errorf("tool message must have a tool_call_id")
		}
	}

	for _, tc := range m.ToolCalls {
		if err := tc.Validate(); err != nil {
			return fmt.Errorf("invalid tool call: %w", err)
		}
	}

	return nil
}

func (m *ChatMessage) IsUserMessage() bool      { return m != nil && m.Role == ChatMessageRoleUser }
func (m *ChatMessage) IsAssistantMessage() bool { return m != nil && m.Role == ChatMessageRoleAssistant }
func (m *ChatMessage) IsSystemMessage() bool    { return m != nil && m.Role == ChatMessageRoleSystem }
func (m *ChatMessage) IsToolMessage() bool      { return m != nil && m.Role == ChatMessageRoleTool }

type MessageOption func(*ChatMessage)

func WithName(name string) MessageOption {
	return func(m *ChatMessage) { m.Name = &name }
}

func WithToolCalls(calls []ToolCall) MessageOption {
	return func(m *ChatMessage) {
		if len(calls) == 0 {
			m.ToolCalls = nil
		} else {
			m.ToolCalls = calls
		}
	}
}

func WithToolCallID(id string) MessageOption {
	return func(m *ChatMessage) { m.ToolCallID = &id }
}

func WithPrefix(prefix bool) MessageOption {
	return func(m *ChatMessage) { m.Prefix = &prefix }
}

func NewUserMessage(content string, opts ...MessageOption) *ChatMessage {
	m := &ChatMessage{Role: ChatMessageRoleUser, Content: content}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func NewSystemMessage(content string, opts ...MessageOption) *ChatMessage {
	m := &ChatMessage{Role: ChatMessageRoleSystem, Content: content}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func NewAssistantMessage(content string, opts ...MessageOption) *ChatMessage {
	m := &ChatMessage{Role: ChatMessageRoleAssistant, Content: content}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func NewToolMessage(content string, opts ...MessageOption) *ChatMessage {
	m := &ChatMessage{Role: ChatMessageRoleTool, Content: content}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// --- Tools ---

type ToolChoice string

const (
	ToolChoiceAuto     ToolChoice = "auto"
	ToolChoiceNone     ToolChoice = "none"
	ToolChoiceAny      ToolChoice = "any"
	ToolChoiceRequired ToolChoice = "required"
)

func (tc *ToolChoice) Validate() error {
	if tc == nil {
		return fmt.Errorf("tool choice is nil")
	}
	switch *tc {
	case ToolChoiceAuto, ToolChoiceNone, ToolChoiceAny, ToolChoiceRequired:
		return nil
	default:
		if *tc == "" {
			return fmt.Errorf("empty tool choice")
		}
		return nil
	}
}

func (tc *ToolChoice) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	// Try unmarshaling as a string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*tc = ToolChoice(s)
		return nil
	}

	// Try unmarshaling as an object: {"type": "function", "function": {"name": "my_func"}}
	var obj struct {
		Type     string   `json:"type"`
		Function struct {
			Name string `json:"name"`
		} `json:"function"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}

	if obj.Type != "function" {
		return fmt.Errorf("invalid tool choice type: %s", obj.Type)
	}
	if obj.Function.Name == "" {
		return fmt.Errorf("tool choice function name is required")
	}

	*tc = ToolChoice(obj.Function.Name)
	return nil
}

type ToolType string

const (
	ToolTypeFunction ToolType = "function"
)

func (tt *ToolType) Validate() error {
	if tt == nil {
		return fmt.Errorf("tool type is nil")
	}
	if *tt != ToolTypeFunction {
		return fmt.Errorf("invalid tool type: %s", *tt)
	}
	return nil
}

type Tool struct {
	Type     ToolType `json:"type"`
	Function Function `json:"function"`
}

func (t *Tool) Validate() error {
	if t == nil {
		return fmt.Errorf("tool is nil")
	}
	if err := t.Type.Validate(); err != nil {
		return err
	}
	return t.Function.Validate()
}

type Function struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

func (f *Function) Validate() error {
	if f == nil {
		return fmt.Errorf("function is nil")
	}
	if f.Name == "" {
		return fmt.Errorf("function name is required")
	}
	return nil
}

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

func (tc *ToolCall) Validate() error {
	if tc == nil {
		return fmt.Errorf("tool call is nil")
	}
	if tc.ID == "" {
		return fmt.Errorf("tool call id is required")
	}
	if err := tc.Type.Validate(); err != nil {
		return err
	}
	return tc.Function.Validate()
}

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

func (fc *FunctionCall) Validate() error {
	if fc == nil {
		return fmt.Errorf("function call is nil")
	}
	if fc.Name == "" {
		return fmt.Errorf("function call name is required")
	}
	if fc.Arguments == "" {
		return fmt.Errorf("function call arguments are required")
	}
	return nil
}

func UnmarshalArguments[T any](tc ToolCall) (T, error) {
	var v T
	err := json.Unmarshal([]byte(tc.Function.Arguments), &v)
	return v, err
}

// --- Reasoning & Guardrails ---

type ReasoningEffort string

const (
	ReasoningEffortHigh   ReasoningEffort = "high"
	ReasoningEffortMedium ReasoningEffort = "medium"
	ReasoningEffortLow    ReasoningEffort = "low"
	ReasoningEffortNone   ReasoningEffort = "none"
)

func (re *ReasoningEffort) Validate() error {
	if re == nil {
		return fmt.Errorf("reasoning effort is nil")
	}
	switch *re {
	case ReasoningEffortHigh, ReasoningEffortMedium, ReasoningEffortLow, ReasoningEffortNone:
		return nil
	default:
		return fmt.Errorf("invalid reasoning effort: %s", *re)
	}
}

type GuardrailConfig struct {
	BlockOnError bool              `json:"block_on_error"`
	Moderation   *ModerationConfig `json:"moderation_llm_v2,omitempty"`
}

func (cfg *GuardrailConfig) Validate() error {
	if cfg == nil {
		return fmt.Errorf("guardrail config is nil")
	}
	if cfg.Moderation != nil {
		return cfg.Moderation.Validate()
	}
	return nil
}

type ModerationConfigAction string

const (
	ModerationConfigActionBlock ModerationConfigAction = "block"
	ModerationConfigActionNone  ModerationConfigAction = "none"
)

func (action ModerationConfigAction) Validate() error {
	switch action {
	case ModerationConfigActionBlock, ModerationConfigActionNone:
		return nil
	default:
		return fmt.Errorf("invalid action: %s", action)
	}
}

type ModerationConfig struct {
	Action                   ModerationConfigAction     `json:"action"`
	IgnoreOtherCategories    bool                       `json:"ignore_other_categories"`
	ModelName                Model                      `json:"model_name"`
	CustomCategoryThresholds *ModerationThresholdConfig `json:"custom_category_thresholds,omitempty"`
}

func (cfg *ModerationConfig) Validate() error {
	if cfg == nil {
		return fmt.Errorf("moderation config is nil")
	}
	if err := cfg.Action.Validate(); err != nil {
		return err
	}
	if cfg.ModelName == "" {
		return fmt.Errorf("moderation model name is required")
	}
	if err := cfg.ModelName.Validate(); err != nil {
		return err
	}
	if cfg.CustomCategoryThresholds != nil {
		return cfg.CustomCategoryThresholds.Validate()
	}
	return nil
}

type ModerationThresholdConfig struct {
	Criminal              *float64 `json:"criminal,omitempty"`
	Dangerous             *float64 `json:"dangerous,omitempty"`
	Financial             *float64 `json:"financial,omitempty"`
	HateAndDiscrimination *float64 `json:"hate_and_discrimination,omitempty"`
	Health                *float64 `json:"health,omitempty"`
	Jailbreaking          *float64 `json:"jailbreaking,omitempty"`
	Law                   *float64 `json:"law,omitempty"`
	Pii                   *float64 `json:"pii,omitempty"`
	Selfharm              *float64 `json:"selfharm,omitempty"`
	Sexual                *float64 `json:"sexual,omitempty"`
	ViolenceAndThreats    *float64 `json:"violence_and_threats,omitempty"`
}

func (cfg *ModerationThresholdConfig) Validate() error {
	if cfg == nil {
		return fmt.Errorf("moderation category thresholds is nil")
	}
	thresholds := []*float64{
		cfg.Criminal, cfg.Dangerous, cfg.Financial, cfg.HateAndDiscrimination,
		cfg.Health, cfg.Jailbreaking, cfg.Law, cfg.Pii, cfg.Selfharm,
		cfg.Sexual, cfg.ViolenceAndThreats,
	}
	for _, t := range thresholds {
		if t != nil && (*t < 0 || *t > 1) {
			return fmt.Errorf("invalid threshold value: %f (must be between 0 and 1)", *t)
		}
	}
	return nil
}
