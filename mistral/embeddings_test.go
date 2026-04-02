package mistral_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onkyou/go-mistral/mistral"
)

func TestEmbeddingRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     *mistral.EmbeddingRequest
		wantErr bool
	}{
		{
			name:    "valid basic request",
			req:     &mistral.EmbeddingRequest{Model: mistral.ModelMistralEmbed, Input: []string{"test"}},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name:    "invalid model",
			req:     &mistral.EmbeddingRequest{Model: mistral.Model("invalid"), Input: []string{"test"}},
			wantErr: true,
		},
		{
			name:    "empty input",
			req:     &mistral.EmbeddingRequest{Model: mistral.ModelMistralEmbed, Input: []string{}},
			wantErr: true,
		},
		{
			name:    "invalid encoding format",
			req:     &mistral.EmbeddingRequest{Model: mistral.ModelMistralEmbed, Input: []string{"test"}, EncodingFormat: mistral.EmbeddingEncodingFormat("invalid")},
			wantErr: true,
		},
		{
			name:    "invalid output data type",
			req:     &mistral.EmbeddingRequest{Model: mistral.ModelMistralEmbed, Input: []string{"test"}, OutputDataType: mistral.EmbeddingOutputDataType("invalid")},
			wantErr: true,
		},
		{
			name:    "invalid output dimension",
			req:     &mistral.EmbeddingRequest{Model: mistral.ModelMistralEmbed, Input: []string{"test"}, OutputDimension: intPtr(0)},
			wantErr: true,
		},
		{
			name:    "valid output dimension",
			req:     &mistral.EmbeddingRequest{Model: mistral.ModelMistralEmbed, Input: []string{"test"}, OutputDimension: intPtr(512)},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmbeddingEncodingFormat_Validate(t *testing.T) {
	tests := []struct {
		name    string
		format  mistral.EmbeddingEncodingFormat
		wantErr bool
	}{
		{"ValidFloat", mistral.EmbeddingEncodingFormatFloat, false},
		{"ValidBase64", mistral.EmbeddingEncodingFormatBase64, false},
		{"Invalid", mistral.EmbeddingEncodingFormat("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.format.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmbeddingOutputDataType_Validate(t *testing.T) {
	tests := []struct {
		name    string
		odt     mistral.EmbeddingOutputDataType
		wantErr bool
	}{
		{"ValidFloat", mistral.EmbeddingOutputDataTypeFloat, false},
		{"ValidInt8", mistral.EmbeddingOutputDataTypeInt8, false},
		{"ValidUint8", mistral.EmbeddingOutputDataTypeUint8, false},
		{"ValidBinary", mistral.EmbeddingOutputDataTypeBinary, false},
		{"ValidUbinary", mistral.EmbeddingOutputDataTypeUbinary, false},
		{"Invalid", mistral.EmbeddingOutputDataType("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.odt.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func intPtr(i int) *int { return &i }

func TestEmbeddingService_CreateWithFields(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	client, _ := mistral.NewClient(&mistral.ClientConfig{APIKey: "test-key", BaseURL: server.URL})

	mux.HandleFunc("/v1/embeddings", func(w http.ResponseWriter, r *http.Request) {
		var req mistral.EmbeddingRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.EncodingFormat != mistral.EmbeddingEncodingFormatBase64 {
			t.Errorf("expected encoding base64, got %s", req.EncodingFormat)
		}
		if req.OutputDataType != mistral.EmbeddingOutputDataTypeInt8 {
			t.Errorf("expected output data type int8, got %s", req.OutputDataType)
		}
		if req.OutputDimension == nil || *req.OutputDimension != 512 {
			t.Errorf("expected output dimension 512, got %v", req.OutputDimension)
		}

		if req.Metadata == nil || req.Metadata["user_id"] != "123" {
			t.Errorf("expected metadata user_id=123, got %v", req.Metadata)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&mistral.EmbeddingResponse{ID: "opt-123"})
	})

	req := &mistral.EmbeddingRequest{
		Model:           mistral.ModelMistralEmbed,
		Input:           []string{"test"},
		EncodingFormat:  mistral.EmbeddingEncodingFormatBase64,
		OutputDataType:  mistral.EmbeddingOutputDataTypeInt8,
		OutputDimension: intPtr(512),
		Metadata:        map[string]any{"user_id": "123"},
	}

	_, _, err := client.Embedding.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
}

func TestEmbeddingService_Create(t *testing.T) {
	tests := []struct {
		name           string
		model          mistral.Model
		input          string
		wantInput      []string
		wantEncoding   mistral.EmbeddingEncodingFormat
		mockResponse   *mistral.EmbeddingResponse
		mockStatusCode int
		wantErr        bool
	}{
		{
			name:      "successful single embedding",
			model:     mistral.ModelMistralEmbed,
			input:     "test input",
			wantInput: []string{"test input"},
			mockResponse: &mistral.EmbeddingResponse{
				ID: "emb-123",
				Data: []mistral.EmbeddingResponseData{
					{Embedding: []float64{0.1, 0.2}, Index: 0},
				},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:         "with options",
			model:        mistral.ModelMistralEmbed,
			input:        "test input",
			wantInput:    []string{"test input"},
			wantEncoding: mistral.EmbeddingEncodingFormatBase64,
			mockResponse: &mistral.EmbeddingResponse{
				ID: "emb-456",
				Data: []mistral.EmbeddingResponseData{
					{Embedding: []float64{0.3, 0.4}, Index: 0},
				},
			},
			mockStatusCode: http.StatusOK,
			wantErr:        false,
		},
		{
			name:      "api error",
			model:     mistral.ModelMistralEmbed,
			input:     "invalid",
			wantInput: []string{"invalid"},
			mockStatusCode: http.StatusBadRequest,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			server := httptest.NewServer(mux)
			defer server.Close()

			client, _ := mistral.NewClient(&mistral.ClientConfig{APIKey: "test-key", BaseURL: server.URL})

			mux.HandleFunc("/v1/embeddings", func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected method %s, got %s", http.MethodPost, r.Method)
				}

				var req mistral.EmbeddingRequest
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

			req := &mistral.EmbeddingRequest{
				Model: tt.model,
				Input: []string{tt.input},
			}
			if tt.name == "with options" {
				req.EncodingFormat = mistral.EmbeddingEncodingFormatBase64
			}

			resp, _, err := client.Embedding.Create(context.Background(), req)
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
		model          mistral.Model
		inputs         []string
		mockResponse   *mistral.EmbeddingResponse
		mockStatusCode int
		wantErr        bool
	}{
		{
			name:   "successful batch embedding",
			model:  mistral.ModelMistralEmbed,
			inputs: []string{"input 1", "input 2"},
			mockResponse: &mistral.EmbeddingResponse{
				ID: "emb-batch",
				Data: []mistral.EmbeddingResponseData{
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

			client, _ := mistral.NewClient(&mistral.ClientConfig{APIKey: "test-key", BaseURL: server.URL})

			mux.HandleFunc("/v1/embeddings", func(w http.ResponseWriter, r *http.Request) {
				var req mistral.EmbeddingRequest
				json.NewDecoder(r.Body).Decode(&req)
				if len(req.Input) != len(tt.inputs) {
					t.Errorf("expected %d inputs, got %d", len(tt.inputs), len(req.Input))
				}
				w.WriteHeader(tt.mockStatusCode)
				json.NewEncoder(w).Encode(tt.mockResponse)
			})

			req := &mistral.EmbeddingRequest{
				Model: tt.model,
				Input: tt.inputs,
			}
			resp, _, err := client.Embedding.Create(context.Background(), req)
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
