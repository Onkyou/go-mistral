package mistral_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onkyou/go-mistral/mistral"
)

func TestChatService_Complete(t *testing.T) {
	tests := []struct {
		name           string
		model          mistral.Model
		messages       []*mistral.ChatMessage
		req            *mistral.ChatCompletionRequest
		mockResponse   *mistral.ChatCompletionResponse
		mockStatusCode int
		wantErr        bool
	}{
		{
			name:  "successful completion",
			model: mistral.ModelMistralLargeLatest,
			messages: []*mistral.ChatMessage{
				mistral.NewUserMessage("Hi"),
			},
			mockResponse: &mistral.ChatCompletionResponse{
				ID: "chat-123",
				Choices: []mistral.ChatCompletionChoice{
					{Message: &mistral.ChatMessage{Role: mistral.ChatMessageRoleAssistant, Content: "Hello!"}},
				},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			req: &mistral.ChatCompletionRequest{
				Model:       mistral.ModelMistralLargeLatest,
				Messages:    []*mistral.ChatMessage{mistral.NewUserMessage("Hi")},
				Temperature: func() *float64 { f := 0.7; return &f }(),
			},
			mockResponse: &mistral.ChatCompletionResponse{
				ID: "chat-456",
				Choices: []mistral.ChatCompletionChoice{
					{Message: &mistral.ChatMessage{Role: mistral.ChatMessageRoleAssistant, Content: "Hi there!"}},
				},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			req: &mistral.ChatCompletionRequest{
				Model:    mistral.ModelMistralLargeLatest,
				Messages: []*mistral.ChatMessage{mistral.NewUserMessage("Hi")},
				Guardrails: &mistral.GuardrailConfig{
					BlockOnError: true,
					Moderation: &mistral.ModerationConfig{
						Action:    mistral.ModerationConfigActionBlock,
						ModelName: mistral.ModelMistralModeration2603,
						CustomCategoryThresholds: &mistral.ModerationThresholdConfig{
							Criminal: func() *float64 { f := 0.1; return &f }(),
						},
					},
				},
			},
			mockResponse: &mistral.ChatCompletionResponse{ID: "guardrail-123"},
			wantErr:      false,
		},
		{
			req: &mistral.ChatCompletionRequest{
				Model:    mistral.ModelMistralLargeLatest,
				Messages: []*mistral.ChatMessage{mistral.NewUserMessage("Solve 1+1")},
				ReasoningEffort: func() *mistral.ReasoningEffort {
					r := mistral.ReasoningEffortHigh
					return &r
				}(),
			},
			mockResponse: &mistral.ChatCompletionResponse{ID: "reasoning-123"},
			wantErr:      false,
		},
		{
			req: &mistral.ChatCompletionRequest{
				Model:    mistral.ModelMistralLargeLatest,
				Messages: []*mistral.ChatMessage{mistral.NewUserMessage("output json")},
				ResponseFormat: &mistral.ResponseFormat{
					Type: mistral.ResponseFormatTypeJsonSchema,
					JsonSchema: &mistral.JsonSchema{
						Name:             "output",
						SchemaDefinition: map[string]any{"type": "object"},
						Strict:           true,
					},
				},
			},
			mockResponse: &mistral.ChatCompletionResponse{ID: "format-123"},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			server := httptest.NewServer(mux)
			defer server.Close()

			client, _ := mistral.NewClient(&mistral.ClientConfig{APIKey: "test-key", BaseURL: server.URL})

			mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
				var req mistral.ChatCompletionRequest
				json.NewDecoder(r.Body).Decode(&req)

				if tt.req == nil {
					if req.Model != tt.model {
						t.Errorf("expected model %s, got %s", tt.model, req.Model)
					}
				} else {
					if req.Model != tt.req.Model {
						t.Errorf("expected model %s, got %s", tt.req.Model, req.Model)
					}
				}

				// Basic validation that options were applied
				if tt.req != nil && tt.req.ReasoningEffort != nil && (req.ReasoningEffort == nil || *req.ReasoningEffort != *tt.req.ReasoningEffort) {
					t.Errorf("expected reasoning effort %v, got %v", *tt.req.ReasoningEffort, req.ReasoningEffort)
				}
				if tt.req != nil && tt.req.Guardrails != nil && req.Guardrails == nil {
					t.Errorf("expected guardrails to be set")
				}

				statusCode := tt.mockStatusCode
				if statusCode == 0 {
					statusCode = http.StatusOK
				}
				w.WriteHeader(statusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			})

			if tt.req == nil {
				tt.req = &mistral.ChatCompletionRequest{
					Model:    tt.model,
					Messages: tt.messages,
				}
			}
			resp, _, err := client.Chat.Complete(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Complete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && resp.ID != tt.mockResponse.ID {
				t.Errorf("expected ID %s, got %s", tt.mockResponse.ID, resp.ID)
			}
		})
	}
}

func TestChatService_Stream(t *testing.T) {
	tests := []struct {
		name           string
		model          mistral.Model
		messages       []*mistral.ChatMessage
		mockChunks     []mistral.ChatCompletionResponse
		mockStatusCode int
		wantContent    string
		wantErr        bool
	}{
		{
			name:  "successful streaming",
			model: mistral.ModelMistralLargeLatest,
			messages: []*mistral.ChatMessage{
				mistral.NewUserMessage("Hi"),
			},
			mockChunks: []mistral.ChatCompletionResponse{
				{ID: "1", Choices: []mistral.ChatCompletionChoice{{Delta: &mistral.ChatMessage{Content: "Hello"}}}},
				{ID: "2", Choices: []mistral.ChatCompletionChoice{{Delta: &mistral.ChatMessage{Content: " "}}}},
				{ID: "3", Choices: []mistral.ChatCompletionChoice{{Delta: &mistral.ChatMessage{Content: "World"}}}},
			},
			mockStatusCode: http.StatusOK,
			wantContent:    "Hello World",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			server := httptest.NewServer(mux)
			defer server.Close()

			client, _ := mistral.NewClient(&mistral.ClientConfig{APIKey: "test-key", BaseURL: server.URL})

			mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				w.WriteHeader(tt.mockStatusCode)

				for _, chunk := range tt.mockChunks {
					data, _ := json.Marshal(chunk)
					fmt.Fprintf(w, "data: %s\n\n", data)
				}
				fmt.Fprint(w, "data: [DONE]\n\n")
			})

			req := &mistral.ChatCompletionRequest{
				Model:    tt.model,
				Messages: tt.messages,
			}
			var received string
			var streamErr error
			for chunk, err := range client.Chat.Stream(context.Background(), req) {
				if err != nil {
					streamErr = err
					break
				}
				if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
					received += chunk.Choices[0].Delta.Content
				}
			}

			if (streamErr != nil) != tt.wantErr {
				t.Errorf("Stream() error = %v, wantErr %v", streamErr, tt.wantErr)
			}

			if !tt.wantErr && received != tt.wantContent {
				t.Errorf("expected content %q, got %q", tt.wantContent, received)
			}
		})
	}
}

func TestChatService_Stream_Errors(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	client, _ := mistral.NewClient(&mistral.ClientConfig{APIKey: "test-key", BaseURL: server.URL})

	t.Run("validation error", func(t *testing.T) {
		req := &mistral.ChatCompletionRequest{
			Model:    mistral.Model(""),
			Messages: []*mistral.ChatMessage{mistral.NewUserMessage("Hi")},
		}
		for _, err := range client.Chat.Stream(context.Background(), req) {
			if err == nil {
				t.Error("expected error, got nil")
			}
			return
		}
	})
}

func TestChatCompletionRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     *mistral.ChatCompletionRequest
		wantErr bool
	}{
		{
			name:    "valid basic request",
			req:     &mistral.ChatCompletionRequest{Model: mistral.ModelMistralLargeLatest, Messages: []*mistral.ChatMessage{mistral.NewUserMessage("Hi")}},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name:    "invalid model",
			req:     &mistral.ChatCompletionRequest{Model: mistral.Model("invalid"), Messages: []*mistral.ChatMessage{mistral.NewUserMessage("Hi")}},
			wantErr: true,
		},
		{
			name:    "no messages",
			req:     &mistral.ChatCompletionRequest{Model: mistral.ModelMistralLargeLatest, Messages: []*mistral.ChatMessage{}},
			wantErr: true,
		},
		{
			name: "invalid message in slice",
			req: &mistral.ChatCompletionRequest{
				Model:    mistral.ModelMistralLargeLatest,
				Messages: []*mistral.ChatMessage{mistral.NewUserMessage("")},
			},
			wantErr: true,
		},
		{
			name: "temperature too high",
			req: &mistral.ChatCompletionRequest{
				Model:       mistral.ModelMistralLargeLatest,
				Messages:    []*mistral.ChatMessage{mistral.NewUserMessage("Hi")},
				Temperature: func() *float64 { f := 1.1; return &f }(),
			},
			wantErr: true,
		},
		{
			name: "top_p too low",
			req: &mistral.ChatCompletionRequest{
				Model:    mistral.ModelMistralLargeLatest,
				Messages: []*mistral.ChatMessage{mistral.NewUserMessage("Hi")},
				TopP:     func() *float64 { f := -0.1; return &f }(),
			},
			wantErr: true,
		},
		{
			name: "presence_penalty out of range",
			req: &mistral.ChatCompletionRequest{
				Model:           mistral.ModelMistralLargeLatest,
				Messages:        []*mistral.ChatMessage{mistral.NewUserMessage("Hi")},
				PresencePenalty: func() *float64 { f := 2.5; return &f }(),
			},
			wantErr: true,
		},
		{
			name: "max_tokens zero",
			req: &mistral.ChatCompletionRequest{
				Model:     mistral.ModelMistralLargeLatest,
				Messages:  []*mistral.ChatMessage{mistral.NewUserMessage("Hi")},
				MaxTokens: func() *int { i := 0; return &i }(),
			},
			wantErr: true,
		},
		{
			name: "valid with tools",
			req: &mistral.ChatCompletionRequest{
				Model:    mistral.ModelMistralLargeLatest,
				Messages: []*mistral.ChatMessage{mistral.NewUserMessage("Hi")},
				Tools:    []mistral.Tool{{Type: mistral.ToolTypeFunction, Function: mistral.Function{Name: "test"}}},
			},
			wantErr: false,
		},
		{
			name: "invalid tool",
			req: &mistral.ChatCompletionRequest{
				Model:    mistral.ModelMistralLargeLatest,
				Messages: []*mistral.ChatMessage{mistral.NewUserMessage("Hi")},
				Tools:    []mistral.Tool{{Type: mistral.ToolTypeFunction, Function: mistral.Function{Name: ""}}},
			},
			wantErr: true,
		},
		{
			name: "frequency_penalty out of range",
			req: &mistral.ChatCompletionRequest{
				Model:            mistral.ModelMistralLargeLatest,
				Messages:         []*mistral.ChatMessage{mistral.NewUserMessage("Hi")},
				FrequencyPenalty: func() *float64 { f := -2.1; return &f }(),
			},
			wantErr: true,
		},
		{
			name: "n zero",
			req: &mistral.ChatCompletionRequest{
				Model:    mistral.ModelMistralLargeLatest,
				Messages: []*mistral.ChatMessage{mistral.NewUserMessage("Hi")},
				N:        func() *int { i := 0; return &i }(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("%s: Validate() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestResponseFormat_Validate(t *testing.T) {
	tests := []struct {
		name    string
		rf      *mistral.ResponseFormat
		wantErr bool
	}{
		{"ValidText", &mistral.ResponseFormat{Type: mistral.ResponseFormatTypeText}, false},
		{"ValidJson", &mistral.ResponseFormat{Type: mistral.ResponseFormatTypeJsonObject}, false},
		{"ValidSchema", &mistral.ResponseFormat{Type: mistral.ResponseFormatTypeJsonSchema, JsonSchema: &mistral.JsonSchema{Name: "test", SchemaDefinition: map[string]any{"type": "object"}}}, false},
		{"InvalidType", &mistral.ResponseFormat{Type: mistral.ResponseFormatType("unknown")}, true},
		{"MissingSchema", &mistral.ResponseFormat{Type: mistral.ResponseFormatTypeJsonSchema}, true},
		{"NilFormat", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.rf.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJsonSchema_Validate(t *testing.T) {
	tests := []struct {
		name    string
		js      *mistral.JsonSchema
		wantErr bool
	}{
		{"Valid", &mistral.JsonSchema{Name: "test", SchemaDefinition: map[string]any{"type": "object"}}, false},
		{"MissingName", &mistral.JsonSchema{SchemaDefinition: map[string]any{"type": "object"}}, true},
		{"MissingDefinition", &mistral.JsonSchema{Name: "test"}, true},
		{"NilSchema", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.js.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
