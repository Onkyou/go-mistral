package mistral

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmbeddingService_Create(t *testing.T) {
	tests := []struct {
		name           string
		model          Model
		input          string
		opts           []EmbeddingRequestOption
		wantInput      []string
		wantEncoding   EmbeddingEncodingFormat
		mockResponse   *EmbeddingResponse
		mockStatusCode int
		wantErr        bool
	}{
		{
			name:      "successful single embedding",
			model:     ModelMistralEmbed,
			input:     "test input",
			wantInput: []string{"test input"},
			mockResponse: &EmbeddingResponse{
				ID: "emb-123",
				Data: []EmbeddingResponseData{
					{Embedding: []float64{0.1, 0.2}, Index: 0},
				},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:         "with options",
			model:        ModelMistralEmbed,
			input:        "test input",
			opts:         []EmbeddingRequestOption{WithEncodingFormat(EmbeddingEncodingFormatBase64)},
			wantInput:    []string{"test input"},
			wantEncoding: EmbeddingEncodingFormatBase64,
			mockResponse: &EmbeddingResponse{
				ID: "emb-456",
				Data: []EmbeddingResponseData{
					{Embedding: []float64{0.3, 0.4}, Index: 0},
				},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "api error",
			model:          ModelMistralEmbed,
			input:          "invalid",
			wantInput:      []string{"invalid"},
			mockStatusCode: http.StatusBadRequest,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			server := httptest.NewServer(mux)
			defer server.Close()

			client, _ := NewClient(WithAPIKey("test-key"), WithBaseURL(server.URL))

			mux.HandleFunc("/v1/embeddings", func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected method %s, got %s", http.MethodPost, r.Method)
				}

				var req EmbeddingRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("failed to decode request: %v", err)
				}

				// Validate request
				if req.Model != tt.model {
					t.Errorf("expected model %s, got %s", tt.model, req.Model)
				}
				if len(req.Input) != len(tt.wantInput) {
					t.Errorf("expected %d inputs, got %d", len(tt.wantInput), len(req.Input))
				}
				if tt.wantEncoding != "" && req.EncodingFormat != tt.wantEncoding {
					t.Errorf("expected encoding %s, got %s", tt.wantEncoding, req.EncodingFormat)
				}

				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			})

			resp, _, err := client.Embedding.Create(context.Background(), tt.model, []string{tt.input}, tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && resp.ID != tt.mockResponse.ID {
				t.Errorf("expected ID %s, got %s", tt.mockResponse.ID, resp.ID)
			}
		})
	}
}

func TestEmbeddingService_CreateBatch(t *testing.T) {
	tests := []struct {
		name           string
		model          Model
		inputs         []string
		mockResponse   *EmbeddingResponse
		mockStatusCode int
		wantErr        bool
	}{
		{
			name:   "successful batch embedding",
			model:  ModelMistralEmbed,
			inputs: []string{"input 1", "input 2"},
			mockResponse: &EmbeddingResponse{
				ID: "emb-batch",
				Data: []EmbeddingResponseData{
					{Embedding: []float64{0.1}, Index: 0},
					{Embedding: []float64{0.2}, Index: 1},
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

			mux.HandleFunc("/v1/embeddings", func(w http.ResponseWriter, r *http.Request) {
				var req EmbeddingRequest
				json.NewDecoder(r.Body).Decode(&req)
				if len(req.Input) != len(tt.inputs) {
					t.Errorf("expected %d inputs, got %d", len(tt.inputs), len(req.Input))
				}
				w.WriteHeader(tt.mockStatusCode)
				json.NewEncoder(w).Encode(tt.mockResponse)
			})

			resp, _, err := client.Embedding.Create(context.Background(), tt.model, tt.inputs)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(resp.Data) != len(tt.inputs) {
				t.Errorf("expected %d results, got %d", len(tt.inputs), len(resp.Data))
			}
		})
	}
}
