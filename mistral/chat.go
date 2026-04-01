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

type ResponseFormatType string

const (
	ResponseFormatTypeText       ResponseFormatType = "text"
	ResponseFormatTypeJsonObject ResponseFormatType = "json_object"
	ResponseFormatTypeJsonSchema ResponseFormatType = "json_schema"
)

// ResponseFormat represents the format of the response.
type ResponseFormat struct {
	Type       ResponseFormatType `json:"type"`
	JsonSchema *JsonSchema        `json:"json_schema,omitempty"`
}

type JsonSchema struct {
	Name             string         `json:"name"`
	Description      string         `json:"description,omitempty"`
	SchemaDefinition map[string]any `json:"schema_definition"`
	Strict           bool           `json:"strict"`
}

// ChatCompletionRequest represents the request payload for the chat completion API.
type ChatCompletionRequest struct {
	// The frequency_penalty penalizes the repetition of words based on their frequency in the generated text. A higher frequency penalty discourages the model from repeating words that have already appeared frequently in the output, promoting diversity and reducing repetition.
	FrequencyPenalty float64 `json:"frequency_penalty"`

	// Guardrails are used to control the output of the model.
	Guardrails *GuardrailConfig `json:"guardrail,omitempty"`

	// The maximum number of tokens to generate in the completion. The token count of your prompt plus max_tokens cannot exceed the model's context length.
	MaxTokens *int `json:"max_tokens,omitempty"`

	// The prompt(s) to generate completions for, encoded as a list of dict with role and content.
	Messages []ChatMessage `json:"messages"`

	// A dictionary of metadata to attach to the request.
	Metadata map[string]string `json:"metadata,omitempty"`

	// ID of the model to use.
	Model Model `json:"model"`

	// Number of completions to return for each request, input tokens are only billed once.
	N *int `json:"n,omitempty"`

	// Whether to enable parallel function calling during tool use, when enabled the model can call multiple tools in parallel.
	ParallelToolCalls *bool `json:"parallel_tool_calls,omitempty"`

	// The presence_penalty determines how much the model penalizes the repetition of words or phrases. A higher presence penalty encourages the model to use a wider variety of words and phrases, making the output more diverse and creative.
	PresencePenalty float64 `json:"presence_penalty"`

	// The seed to use for random sampling. If set, different calls will generate deterministic results.
	RandomSeed *int `json:"random_seed,omitempty"`

	// Controls the reasoning effort level for reasoning models. "high" enables comprehensive reasoning traces, "none" disables reasoning effort.
	ReasoningEffort *ReasoningEffort `json:"reasoning_effort,omitempty"`

	// Specify the format that the model must output. By default it will use \{ "type": "text" \}. Setting to \{ "type": "json_object" \} enables JSON mode, which guarantees the message the model generates is in JSON. When using JSON mode you MUST also instruct the model to produce JSON yourself with a system or a user message. Setting to \{ "type": "json_schema" \} enables JSON schema mode, which guarantees the message the model generates is in JSON and follows the schema you provide.
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`

	// What sampling temperature to use, we recommend between 0.0 and 0.7. Higher values like 0.7 will make the output more random, while lower values like 0.2 will make it more focused and deterministic. We generally recommend altering this or top_p but not both. The default value varies depending on the model you are targeting. Call the /models endpoint to retrieve the appropriate value.
	Temperature *float64 `json:"temperature,omitempty"`

	// Nucleus sampling, where the model considers the results of the tokens with top_p probability mass. So 0.1 means only the tokens comprising the top 10% probability mass are considered. We generally recommend altering this or temperature but not both.
	TopP *float64 `json:"top_p,omitempty"`

	// Whether to stream back partial progress. If set, tokens will be sent as data-only server-side events as they become available, with the stream terminated by a data: [DONE] message. Otherwise, the server will hold the request open until the timeout or until completion, with the response containing the full result as JSON.
	Stream bool `json:"stream,omitempty"`

	// Whether to inject a safety prompt before all conversations.
	SafePrompt bool `json:"safe_prompt"`

	// Stop generation if this token is detected. Or if one of these tokens is detected when providing an array
	Stop []string `json:"stop,omitempty"`

	// Controls which (if any) tool is called by the model. none means the model will not call any tool and instead generates a message. auto means the model can pick between generating a message or calling one or more tools. any or required means the model must call one or more tools. Specifying a particular tool forces the model to call that tool.
	ToolChoice ToolChoice `json:"tool_choice,omitempty"`

	// A list of tools the model may call.
	Tools []Tool `json:"tools,omitempty"`
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

type FinishReason string

const (
	FinishReasonStop        FinishReason = "stop"
	FinishReasonLength      FinishReason = "length"
	FinishReasonModelLength FinishReason = "model_length"
	FinishReasonError       FinishReason = "error"
	FinishReasonToolCalls   FinishReason = "tool_calls"
)

// ChatCompletionChoice represents a single completion choice.
type ChatCompletionChoice struct {
	Index   int          `json:"index"`
	Message *ChatMessage `json:"message,omitempty"`

	// Delta can be used by consumers to get the partial results in a streaming completion.
	Delta        *ChatMessage `json:"delta,omitempty"`
	FinishReason FinishReason `json:"finish_reason"`
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

// WithSafePrompt sets the safe_prompt flag for the request.
func WithSafePrompt(s bool) ChatCompletionRequestOption {
	return func(r *ChatCompletionRequest) {
		r.SafePrompt = s
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
func WithToolChoice(choice ToolChoice) ChatCompletionRequestOption {
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
		Stream:   true,
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
