package mistral

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChatService_Complete(t *testing.T) {
	tests := []struct {
		name           string
		model          Model
		messages       []ChatMessage
		opts           []ChatCompletionRequestOption
		mockResponse   *ChatCompletionResponse
		mockStatusCode int
		wantErr        bool
	}{
		{
			name:  "successful completion",
			model: ModelMistralLargeLatest,
			messages: []ChatMessage{
				{Role: RoleUser, Content: "Hi"},
			},
			mockResponse: &ChatCompletionResponse{
				ID: "chat-123",
				Choices: []ChatChoice{
					{Message: &ChatMessage{Role: RoleAssistant, Content: "Hello!"}},
				},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:  "with temperature",
			model: ModelMistralLargeLatest,
			messages: []ChatMessage{
				{Role: RoleUser, Content: "Hi"},
			},
			opts: []ChatCompletionRequestOption{WithTemperature(0.7)},
			mockResponse: &ChatCompletionResponse{
				ID: "chat-456",
				Choices: []ChatChoice{
					{Message: &ChatMessage{Role: RoleAssistant, Content: "Hi there!"}},
				},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			server := httptest.NewServer(mux)
			defer server.Close()

			client, _ := NewClient(WithAPIKey("test-key"), WithBaseURL(server.URL))

			mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
				var req ChatCompletionRequest
				json.NewDecoder(r.Body).Decode(&req)

				if req.Model != tt.model {
					t.Errorf("expected model %s, got %s", tt.model, req.Model)
				}

				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			})

			resp, _, err := client.Chat.Complete(context.Background(), tt.model, tt.messages, tt.opts...)
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
		model          Model
		messages       []ChatMessage
		mockChunks     []ChatCompletionResponse
		mockStatusCode int
		wantContent    string
		wantErr        bool
	}{
		{
			name:  "successful streaming",
			model: ModelMistralLargeLatest,
			messages: []ChatMessage{
				{Role: RoleUser, Content: "Hi"},
			},
			mockChunks: []ChatCompletionResponse{
				{ID: "1", Choices: []ChatChoice{{Delta: &ChatMessage{Content: "Hello"}}}},
				{ID: "2", Choices: []ChatChoice{{Delta: &ChatMessage{Content: " "}}}},
				{ID: "3", Choices: []ChatChoice{{Delta: &ChatMessage{Content: "World"}}}},
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

			client, _ := NewClient(WithAPIKey("test-key"), WithBaseURL(server.URL))

			mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				w.WriteHeader(tt.mockStatusCode)

				for _, chunk := range tt.mockChunks {
					data, _ := json.Marshal(chunk)
					fmt.Fprintf(w, "data: %s\n\n", data)
				}
				fmt.Fprint(w, "data: [DONE]\n\n")
			})

			dataChan, errChan := client.Chat.Stream(context.Background(), tt.model, tt.messages)

			var received string
			for chunk := range dataChan {
				if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
					received += chunk.Choices[0].Delta.Content
				}
			}

			if err := <-errChan; (err != nil) != tt.wantErr {
				t.Errorf("Stream() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && received != tt.wantContent {
				t.Errorf("expected content %q, got %q", tt.wantContent, received)
			}
		})
	}
}
