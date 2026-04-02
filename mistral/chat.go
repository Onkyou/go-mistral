package mistral

import (
	"context"
	"fmt"
	"iter"
	"net/http"
)

// --- ChatService ---

// ChatService provides access to chat completion related functions in the
// Mistral API.
type ChatService service

// Complete creates a chat completion.
func (svc *ChatService) Complete(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, *Response, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	uri := "v1/chat/completions"
	httpReq, err := svc.client.NewRequest(http.MethodPost, uri, req)
	if err != nil {
		return nil, nil, err
	}

	var resp ChatCompletionResponse
	r, err := svc.client.Do(ctx, httpReq, &resp)
	if err != nil {
		return nil, r, err
	}

	return &resp, r, nil
}

// Stream creates a streaming chat completion. It returns an iterator for
// completion chunks and error signaling.
func (svc *ChatService) Stream(ctx context.Context, req *ChatCompletionRequest) iter.Seq2[*ChatCompletionResponse, error] {
	return func(yield func(*ChatCompletionResponse, error) bool) {
		req.Stream = true
		if err := req.Validate(); err != nil {
			yield(nil, err)
			return
		}

		uri := "v1/chat/completions"
		httpReq, err := svc.client.NewRequest(http.MethodPost, uri, req)
		if err != nil {
			yield(nil, err)
			return
		}

		for chunk, err := range Stream[*ChatCompletionResponse](ctx, svc.client, httpReq) {
			if !yield(chunk, err) {
				return
			}
		}
	}
}

// --- Request/Response Models ---

// ChatCompletionRequest represents a request for a chat completion.
type ChatCompletionRequest struct {
	Model    Model         `json:"model"`
	Messages []*ChatMessage `json:"messages"`

	// Temperature specifies the sampling temperature.
	// Range: [0, 1]
	Temperature *float64 `json:"temperature,omitempty"`

	// TopP specifies the nucleus sampling probability.
	// Range: [0, 1]
	TopP *float64 `json:"top_p,omitempty"`

	// MaxTokens specifies the maximum number of tokens to generate.
	MaxTokens *int `json:"max_tokens,omitempty"`

	// Stream specifies whether to stream the response.
	Stream bool `json:"stream,omitempty"`

	// Stop specifies up to 4 stop sequences.
	Stop any `json:"stop,omitempty"`

	// RandomSeed specifies a seed for deterministic generation.
	RandomSeed *int `json:"random_seed,omitempty"`

	// ResponseFormat specifies the format of the response.
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`

	// Tools specifies a list of tools the model may call.
	Tools []Tool `json:"tools,omitempty"`

	// ToolChoice specifies the tool selection strategy.
	ToolChoice ToolChoice `json:"tool_choice,omitempty"`

	// PresencePenalty specifies the presence penalty.
	// Range: [-2, 2]
	PresencePenalty *float64 `json:"presence_penalty,omitempty"`

	// FrequencyPenalty specifies the frequency penalty.
	// Range: [-2, 2]
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`

	// N specifies the number of completions to generate.
	N *int `json:"n,omitempty"`

	// Prediction specifies the prediction configuration. (Alpha)
	Prediction any `json:"prediction,omitempty"`

	// SafePrompt specifies whether to use a safe prompt.
	SafePrompt bool `json:"safe_prompt,omitempty"`

	// Guardrails specifies the guardrail configuration.
	Guardrails *GuardrailConfig `json:"guardrails,omitempty"`

	// ReasoningEffort specifies the reasoning effort.
	ReasoningEffort *ReasoningEffort `json:"reasoning_effort,omitempty"`
}

func (r *ChatCompletionRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("chat completion request is nil")
	}

	if err := r.Model.Validate(); err != nil {
		return err
	}

	if len(r.Messages) == 0 {
		return fmt.Errorf("messages are required")
	}

	for i := range r.Messages {
		if err := r.Messages[i].Validate(); err != nil {
			return fmt.Errorf("invalid message at index %d: %w", i, err)
		}
	}

	if r.Temperature != nil && (*r.Temperature < 0 || *r.Temperature > 1) {
		return fmt.Errorf("temperature must be between 0 and 1, got %f", *r.Temperature)
	}

	if r.TopP != nil && (*r.TopP < 0 || *r.TopP > 1) {
		return fmt.Errorf("top_p must be between 0 and 1, got %f", *r.TopP)
	}

	if r.PresencePenalty != nil && (*r.PresencePenalty < -2 || *r.PresencePenalty > 2) {
		return fmt.Errorf("presence_penalty must be between -2 and 2, got %f", *r.PresencePenalty)
	}

	if r.FrequencyPenalty != nil && (*r.FrequencyPenalty < -2 || *r.FrequencyPenalty > 2) {
		return fmt.Errorf("frequency_penalty must be between -2 and 2, got %f", *r.FrequencyPenalty)
	}

	if r.MaxTokens != nil && *r.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be greater than 0")
	}

	if r.N != nil && *r.N <= 0 {
		return fmt.Errorf("n must be greater than 0")
	}

	if r.ResponseFormat != nil {
		if err := r.ResponseFormat.Validate(); err != nil {
			return fmt.Errorf("invalid response_format: %w", err)
		}
	}

	if r.ToolChoice != "" {
		if err := r.ToolChoice.Validate(); err != nil {
			return fmt.Errorf("invalid tool_choice: %w", err)
		}
	}

	for i, tool := range r.Tools {
		if err := tool.Validate(); err != nil {
			return fmt.Errorf("invalid tool at index %d: %w", i, err)
		}
	}

	if r.Guardrails != nil {
		if err := r.Guardrails.Validate(); err != nil {
			return fmt.Errorf("invalid guardrails: %w", err)
		}
	}

	if r.ReasoningEffort != nil {
		if err := r.ReasoningEffort.Validate(); err != nil {
			return fmt.Errorf("invalid reasoning_effort: %w", err)
		}
	}

	return nil
}

// ChatCompletionResponse wraps the chat completion response.
type ChatCompletionResponse struct {
	Choices []ChatCompletionChoice `json:"choices"`
	Created int                    `json:"created"`
	ID      string                 `json:"id"`
	Model   Model                  `json:"model"`
	Object  string                 `json:"object"`
	Usage   UsageInfo              `json:"usage"`
}

// ChatCompletionChoice represents a single completion choice.
type ChatCompletionChoice struct {
	Index   int          `json:"index"`
	Message *ChatMessage `json:"message,omitempty"`

	// Delta can be used by consumers to get the partial results in a streaming completion.
	Delta        *ChatMessage `json:"delta,omitempty"`
	FinishReason FinishReason `json:"finish_reason"`
}

type ResponseFormatType string

const (
	ResponseFormatTypeText       ResponseFormatType = "text"
	ResponseFormatTypeJsonObject ResponseFormatType = "json_object"
	ResponseFormatTypeJsonSchema ResponseFormatType = "json_schema"
)

func (rft *ResponseFormatType) Validate() error {
	if rft == nil {
		return fmt.Errorf("response format type is nil")
	}
	switch *rft {
	case ResponseFormatTypeText, ResponseFormatTypeJsonObject, ResponseFormatTypeJsonSchema:
		return nil
	default:
		return fmt.Errorf("invalid response format type: %s", *rft)
	}
}

// ResponseFormat specifies the format of the response.
type ResponseFormat struct {
	Type       ResponseFormatType `json:"type"`
	JsonSchema *JsonSchema        `json:"json_schema,omitempty"`
}

func (rf *ResponseFormat) Validate() error {
	if rf == nil {
		return fmt.Errorf("response format is nil")
	}
	if err := rf.Type.Validate(); err != nil {
		return err
	}
	if rf.Type == ResponseFormatTypeJsonSchema {
		if rf.JsonSchema == nil {
			return fmt.Errorf("json_schema is required for type json_schema")
		}
		return rf.JsonSchema.Validate()
	}
	return nil
}

// JsonSchema represents a JSON Schema.
type JsonSchema struct {
	Name             string         `json:"name"`
	Description      string         `json:"description,omitempty"`
	SchemaDefinition map[string]any `json:"schema"`
	Strict           bool           `json:"strict,omitempty"`
}

func (js *JsonSchema) Validate() error {
	if js == nil {
		return fmt.Errorf("json schema is nil")
	}
	if js.Name == "" {
		return fmt.Errorf("json schema name is required")
	}
	if js.SchemaDefinition == nil {
		return fmt.Errorf("json schema definition is required")
	}
	return nil
}

type FinishReason string

const (
	FinishReasonStop        FinishReason = "stop"
	FinishReasonLength      FinishReason = "length"
	FinishReasonModelLength FinishReason = "model_length"
	FinishReasonError       FinishReason = "error"
	FinishReasonToolCalls   FinishReason = "tool_calls"
)
