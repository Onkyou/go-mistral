package mistral

import (
	"context"
	"net/http"
)

// EmbeddingService provides access to embedding related functions in the
// Mistral API.
//
// Mistral API docs: https://docs.mistral.ai/api/endpoint/embeddings
type EmbeddingService service

// EmbeddingResponse wraps the embedding data response.
type EmbeddingResponse struct {
	Data   []EmbeddingResponseData `json:"data"`
	ID     string                  `json:"id"`
	Model  string                  `json:"model"`
	Object string                  `json:"object"`
	Usage  UsageInfo               `json:"usage"`
}

type EmbeddingResponseData struct {
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
	Object    string    `json:"object"`
}

type EmbeddingEncodingFormat string

const (
	EmbeddingEncodingFormatFloat  EmbeddingEncodingFormat = "float"
	EmbeddingEncodingFormatBase64 EmbeddingEncodingFormat = "base64"
)

type EmbeddingOutputDimension string

const (
	EmbeddingOutputDimensionInteger EmbeddingOutputDimension = "integer"
)

type EmbeddingOutputDataType string

const (
	EmbeddingOutputDataTypeFloat   EmbeddingOutputDataType = "float"
	EmbeddingOutputDataTypeInt8    EmbeddingOutputDataType = "int8"
	EmbeddingOutputDataTypeUint8   EmbeddingOutputDataType = "uint8"
	EmbeddingOutputDataTypeBinary  EmbeddingOutputDataType = "binary"
	EmbeddingOutputDataTypeUbinary EmbeddingOutputDataType = "ubinary"
)

type EmbeddingRequest struct {
	Input           []string                 `json:"input"`
	Model           Model                    `json:"model"`
	EncodingFormat  EmbeddingEncodingFormat  `json:"encoding_format"`
	Metadata        map[string]any           `json:"metadata,omitempty"`
	OutputDataType  EmbeddingOutputDataType  `json:"output_data_type,omitempty"`
	OutputDimension EmbeddingOutputDimension `json:"output_dimension,omitempty"`
}

type EmbeddingRequestOption func(*EmbeddingRequest)

func WithEncodingFormat(encodingFormat EmbeddingEncodingFormat) EmbeddingRequestOption {
	return func(r *EmbeddingRequest) {
		r.EncodingFormat = encodingFormat
	}
}

func WithMetadata(metadata map[string]any) EmbeddingRequestOption {
	return func(r *EmbeddingRequest) {
		r.Metadata = metadata
	}
}

func WithOutputDataType(outputDataType EmbeddingOutputDataType) EmbeddingRequestOption {
	return func(r *EmbeddingRequest) {
		r.OutputDataType = outputDataType
	}
}

func WithOutputDimension(outputDimension EmbeddingOutputDimension) EmbeddingRequestOption {
	return func(r *EmbeddingRequest) {
		r.OutputDimension = outputDimension
	}
}

func (svc *EmbeddingService) Create(ctx context.Context, model Model, inputs []string, opts ...EmbeddingRequestOption) (*EmbeddingResponse, *Response, error) {
	req := &EmbeddingRequest{
		Model: model,
		Input: inputs,
	}
	for _, opt := range opts {
		opt(req)
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
