package mistral_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onkyou/go-mistral/mistral"
)

func TestClassifiersService_Moderate(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	client, _ := mistral.NewClient(&mistral.ClientConfig{APIKey: "test-key", BaseURL: server.URL})

	mux.HandleFunc("/v1/moderations", func(w http.ResponseWriter, r *http.Request) {
		var req mistral.ModerateRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Model != mistral.ModelMistralModerationLatest {
			t.Errorf("expected model %s, got %s", mistral.ModelMistralModerationLatest, req.Model)
		}

		resp := mistral.ModerationResponse{
			ID:    "mod-123",
			Model: req.Model,
			Results: []mistral.ModerationResult{
				{
					Categories:     map[string]bool{"sexual": false},
					CategoryScores: map[string]float64{"sexual": 0.01},
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	})

	req := &mistral.ModerateRequest{
		Input: "I love coding in Go!",
		Model: mistral.ModelMistralModerationLatest,
	}

	resp, _, err := client.Classifiers.Moderate(context.Background(), req)
	if err != nil {
		t.Fatalf("Moderate() error = %v", err)
	}

	if resp.ID != "mod-123" {
		t.Errorf("expected ID mod-123, got %s", resp.ID)
	}
}

func TestClassifiersService_ModerateChat(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	client, _ := mistral.NewClient(&mistral.ClientConfig{APIKey: "test-key", BaseURL: server.URL})

	mux.HandleFunc("/v1/chat/moderations", func(w http.ResponseWriter, r *http.Request) {
		var req mistral.ModerateChatRequest
		json.NewDecoder(r.Body).Decode(&req)

		if len(req.Input) != 1 || req.Input[0].Content != "Hello" {
			t.Errorf("unexpected input messages")
		}

		resp := mistral.ModerationResponse{
			ID: "mod-chat-123",
			Results: []mistral.ModerationResult{
				{Categories: map[string]bool{"hate": false}},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	})

	req := &mistral.ModerateChatRequest{
		Input: []*mistral.ChatMessage{mistral.NewUserMessage("Hello")},
		Model: mistral.ModelMistralModerationLatest,
	}

	resp, _, err := client.Classifiers.ModerateChat(context.Background(), req)
	if err != nil {
		t.Fatalf("ModerateChat() error = %v", err)
	}

	if resp.ID != "mod-chat-123" {
		t.Errorf("expected ID mod-chat-123, got %s", resp.ID)
	}
}

func TestClassifiersService_Classify(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	client, _ := mistral.NewClient(&mistral.ClientConfig{APIKey: "test-key", BaseURL: server.URL})

	mux.HandleFunc("/v1/classifications", func(w http.ResponseWriter, r *http.Request) {
		resp := mistral.ClassificationResponse{
			ID: "class-123",
			Results: []mistral.ClassificationResult{
				{Target: "positive", Score: 0.99},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	})

	req := &mistral.ClassifyRequest{
		Input: "This is great!",
		Model: mistral.ModelMistralModerationLatest,
	}

	resp, _, err := client.Classifiers.Classify(context.Background(), req)
	if err != nil {
		t.Fatalf("Classify() error = %v", err)
	}

	if resp.Results[0].Target != "positive" {
		t.Errorf("expected target positive, got %s", resp.Results[0].Target)
	}
}

func TestClassifiersService_ClassifyChat(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	client, _ := mistral.NewClient(&mistral.ClientConfig{APIKey: "test-key", BaseURL: server.URL})

	mux.HandleFunc("/v1/chat/classifications", func(w http.ResponseWriter, r *http.Request) {
		var req mistral.ClassifyChatRequest
		json.NewDecoder(r.Body).Decode(&req)

		if len(req.Input.Messages) == 0 {
			t.Errorf("expected messages in input")
		}

		resp := mistral.ClassificationResponse{
			ID: "class-chat-123",
			Results: []mistral.ClassificationResult{
				{Target: "greeting", Score: 0.8},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	})

	req := &mistral.ClassifyChatRequest{
		Input: mistral.ClassifyChatInput{
			Messages: []*mistral.ChatMessage{mistral.NewUserMessage("Hi")},
		},
		Model: mistral.ModelMistralModerationLatest,
	}

	resp, _, err := client.Classifiers.ClassifyChat(context.Background(), req)
	if err != nil {
		t.Fatalf("ClassifyChat() error = %v", err)
	}

	if resp.ID != "class-chat-123" {
		t.Errorf("expected ID class-chat-123, got %s", resp.ID)
	}
}

func TestModerateRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     *mistral.ModerateRequest
		wantErr bool
	}{
		{"Valid", &mistral.ModerateRequest{Input: "test", Model: mistral.ModelMistralModerationLatest}, false},
		{"EmptyInput", &mistral.ModerateRequest{Input: "", Model: mistral.ModelMistralModerationLatest}, true},
		{"MissingModel", &mistral.ModerateRequest{Input: "test", Model: ""}, true},
		{"NilRequest", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestModerateChatRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     *mistral.ModerateChatRequest
		wantErr bool
	}{
		{"Valid", &mistral.ModerateChatRequest{Input: []*mistral.ChatMessage{mistral.NewUserMessage("test")}, Model: mistral.ModelMistralModerationLatest}, false},
		{"EmptyInput", &mistral.ModerateChatRequest{Input: []*mistral.ChatMessage{}, Model: mistral.ModelMistralModerationLatest}, true},
		{"InvalidMessage", &mistral.ModerateChatRequest{Input: []*mistral.ChatMessage{mistral.NewUserMessage("")}, Model: mistral.ModelMistralModerationLatest}, true},
		{"NilRequest", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClassifyRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     *mistral.ClassifyRequest
		wantErr bool
	}{
		{"Valid", &mistral.ClassifyRequest{Input: "test", Model: mistral.ModelMistralModerationLatest}, false},
		{"EmptyInput", &mistral.ClassifyRequest{Input: "", Model: mistral.ModelMistralModerationLatest}, true},
		{"NilRequest", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClassifyChatRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     *mistral.ClassifyChatRequest
		wantErr bool
	}{
		{"Valid", &mistral.ClassifyChatRequest{Input: mistral.ClassifyChatInput{Messages: []*mistral.ChatMessage{mistral.NewUserMessage("test")}}, Model: mistral.ModelMistralModerationLatest}, false},
		{"EmptyInput", &mistral.ClassifyChatRequest{Input: mistral.ClassifyChatInput{}, Model: mistral.ModelMistralModerationLatest}, true},
		{"NilRequest", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
