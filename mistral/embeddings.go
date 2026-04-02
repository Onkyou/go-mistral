package mistral

import (
	"context"
	"fmt"
	"net/http"
)

// --- EmbeddingService ---

// EmbeddingService provides access to embedding related functions in the
// Mistral API.
//
// Mistral API docs: https://docs.mistral.ai/api/endpoint/embeddings
type EmbeddingService service

// Create creates a new embedding from the provided input(s).
func (svc *EmbeddingService) Create(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, *Response, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	uri := "v1/embeddings"
	httpReq, err := svc.client.NewRequest(http.MethodPost, uri, req)
	if err != nil {
		return nil, nil, err
	}

	var resp EmbeddingResponse
	r, err := svc.client.Do(ctx, httpReq, &resp)
	if err != nil {
		return nil, r, err
	}

	return &resp, r, nil
}

// --- Request/Response Models ---

// EmbeddingRequest represents a request for text embedding.
type EmbeddingRequest struct {
	Input           []string                `json:"input"`
	Model           Model                   `json:"model"`
	EncodingFormat  EmbeddingEncodingFormat `json:"encoding_format,omitempty"`
	Metadata        map[string]any          `json:"metadata,omitempty"`
	OutputDataType  EmbeddingOutputDataType `json:"output_data_type,omitempty"`
	OutputDimension *int                    `json:"output_dimension,omitempty"`
}

func (r *EmbeddingRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("embedding request is nil")
	}

	if err := r.Model.Validate(); err != nil {
		return err
	}

	if !r.Model.IsEmbeddingModel() {
		return fmt.Errorf("model %s is not an embedding model", r.Model)
	}

	if len(r.Input) == 0 {
		return fmt.Errorf("input is required")
	}

	if r.EncodingFormat != "" {
		if err := r.EncodingFormat.Validate(); err != nil {
			return err
		}
	}

	if r.OutputDataType != "" {
		if err := r.OutputDataType.Validate(); err != nil {
			return err
		}
	}

	if r.OutputDimension != nil && *r.OutputDimension <= 0 {
		return fmt.Errorf("output dimension must be greater than 0")
	}

	return nil
}

// EmbeddingResponse wraps the embedding data response.
type EmbeddingResponse struct {
	Data   []EmbeddingResponseData `json:"data"`
	ID     string                  `json:"id"`
	Model  string                  `json:"model"`
	Object string                  `json:"object"`
	Usage  UsageInfo               `json:"usage"`
}

// EmbeddingResponseData represents a single embedding unit.
type EmbeddingResponseData struct {
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
	Object    string    `json:"object"`
}

// --- Options and Enums ---

type EmbeddingEncodingFormat string

const (
	EmbeddingEncodingFormatFloat  EmbeddingEncodingFormat = "float"
	EmbeddingEncodingFormatBase64 EmbeddingEncodingFormat = "base64"
)

func (eef *EmbeddingEncodingFormat) Validate() error {
	if eef == nil {
		return fmt.Errorf("embedding encoding format is nil")
	}
	switch *eef {
	case EmbeddingEncodingFormatFloat, EmbeddingEncodingFormatBase64:
		return nil
	default:
		return fmt.Errorf("invalid embedding encoding format: %s", *eef)
	}
}

type EmbeddingOutputDataType string

const (
	EmbeddingOutputDataTypeFloat   EmbeddingOutputDataType = "float"
	EmbeddingOutputDataTypeInt8    EmbeddingOutputDataType = "int8"
	EmbeddingOutputDataTypeUint8   EmbeddingOutputDataType = "uint8"
	EmbeddingOutputDataTypeBinary  EmbeddingOutputDataType = "binary"
	EmbeddingOutputDataTypeUbinary EmbeddingOutputDataType = "ubinary"
)

func (eodt *EmbeddingOutputDataType) Validate() error {
	if eodt == nil {
		return fmt.Errorf("embedding output data type is nil")
	}
	switch *eodt {
	case EmbeddingOutputDataTypeFloat, EmbeddingOutputDataTypeInt8, EmbeddingOutputDataTypeUint8, EmbeddingOutputDataTypeBinary, EmbeddingOutputDataTypeUbinary:
		return nil
	default:
		return fmt.Errorf("invalid embedding output data type: %s", *eodt)
	}
}

func (m *Model) IsEmbeddingModel() bool {
	if m == nil {
		return false
	}
	return *m == ModelMistralEmbed
}
