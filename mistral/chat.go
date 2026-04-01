package mistral

import (
	"context"
	"net/http"
)

// ChatService provides access to chat completions related functions in the
// Mistral API.
//
// Mistral API docs: https://docs.mistral.ai/api/endpoint/chat
type ChatService service

// ChatMessage represents a single message in the conversation.
type ChatMessage struct {
	Role    Role    `json:"role"`
	Content string  `json:"content"`
	Name    *string `json:"name,omitempty"`

	// Tool related fields
	ToolCallID *string    `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall represents a tool call initiated by the model.
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction represents a function call.
type ToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ResponseFormat represents the format of the response.
type ResponseFormat struct {
	Type string `json:"type"`
}

// ChatCompletionRequest represents the request payload for the chat completion API.
type ChatCompletionRequest struct {
	Model          Model           `json:"model"`
	Messages       []ChatMessage   `json:"messages"`
	Temperature    *float64        `json:"temperature,omitempty"`
	TopP           *float64        `json:"top_p,omitempty"`
	MaxTokens      *int            `json:"max_tokens,omitempty"`
	Stream         *bool           `json:"stream,omitempty"`
	SafePrompt     *bool           `json:"safe_prompt,omitempty"`
	RandomSeed     *int            `json:"random_seed,omitempty"`
	Tools          []Tool          `json:"tools,omitempty"`
	ToolChoice     any             `json:"tool_choice,omitempty"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
}

// Tool represents a tool that the model can call.
type Tool struct {
	Type     string         `json:"type"`
	Function ToolDefinition `json:"function"`
}

// ToolDefinition represents a function definition.
type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters"`
}

// ChatCompletionResponse wraps the chat completion response.
type ChatCompletionResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   Model        `json:"model"`
	Choices []ChatChoice `json:"choices"`
	Usage   UsageInfo    `json:"usage"`
}

// ChatChoice represents a single completion choice.
type ChatChoice struct {
	Index        int          `json:"index"`
	Message      *ChatMessage `json:"message,omitempty"`
	Delta        *ChatMessage `json:"delta,omitempty"`
	FinishReason string       `json:"finish_reason"`
}

// ChatCompletionRequestOption is a functional option for ChatCompletionRequest.
type ChatCompletionRequestOption func(*ChatCompletionRequest)

// WithTemperature sets the temperature for the request.
func WithTemperature(t float64) ChatCompletionRequestOption {
	return func(r *ChatCompletionRequest) {
		r.Temperature = new(t)
	}
}

// WithMaxTokens sets the maximum number of tokens for the request.
func WithMaxTokens(m int) ChatCompletionRequestOption {
	return func(r *ChatCompletionRequest) {
		r.MaxTokens = new(m)
	}
}

// WithTopP sets the top_p for the request.
func WithTopP(p float64) ChatCompletionRequestOption {
	return func(r *ChatCompletionRequest) {
		r.TopP = new(p)
	}
}

// WithStream sets the stream flag for the request.
func WithStream(s bool) ChatCompletionRequestOption {
	return func(r *ChatCompletionRequest) {
		r.Stream = new(s)
	}
}

// WithSafePrompt sets the safe_prompt flag for the request.
func WithSafePrompt(s bool) ChatCompletionRequestOption {
	return func(r *ChatCompletionRequest) {
		r.SafePrompt = new(s)
	}
}

// WithRandomSeed sets the random_seed for the request.
func WithRandomSeed(s int) ChatCompletionRequestOption {
	return func(r *ChatCompletionRequest) {
		r.RandomSeed = new(s)
	}
}

// WithTools sets the tools for the request.
func WithTools(tools []Tool) ChatCompletionRequestOption {
	return func(r *ChatCompletionRequest) {
		r.Tools = tools
	}
}

// WithToolChoice sets the tool_choice for the request.
func WithToolChoice(choice any) ChatCompletionRequestOption {
	return func(r *ChatCompletionRequest) {
		r.ToolChoice = choice
	}
}

// WithResponseFormat sets the response_format for the request.
func WithResponseFormat(format ResponseFormat) ChatCompletionRequestOption {
	return func(r *ChatCompletionRequest) {
		r.ResponseFormat = &format
	}
}

// Complete creates a chat completion.
func (svc *ChatService) Complete(ctx context.Context, model Model, messages []ChatMessage, opts ...ChatCompletionRequestOption) (*ChatCompletionResponse, *Response, error) {
	req := &ChatCompletionRequest{
		Model:    model,
		Messages: messages,
	}
	for _, opt := range opts {
		opt(req)
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

// Stream creates a streaming chat completion. It returns two channels:
// one for completion chunks and one for error signaling.
func (svc *ChatService) Stream(ctx context.Context, model Model, messages []ChatMessage, opts ...ChatCompletionRequestOption) (<-chan *ChatCompletionResponse, <-chan error) {
	req := &ChatCompletionRequest{
		Model:    model,
		Messages: messages,
		Stream:   new(true),
	}
	for _, opt := range opts {
		opt(req)
	}

	uri := "v1/chat/completions"
	httpReq, err := svc.client.NewRequest(http.MethodPost, uri, req)
	if err != nil {
		dataChan := make(chan *ChatCompletionResponse)
		errChan := make(chan error, 1)
		close(dataChan)
		errChan <- err
		close(errChan)
		return dataChan, errChan
	}

	return Stream[*ChatCompletionResponse](ctx, svc.client, httpReq)
}
