package mistral

import (
	"context"
	"fmt"
	"net/http"
)

// --- ClassifiersService ---

// ClassifiersService provides access to moderation and classification related functions.
type ClassifiersService service

// Moderate performs text moderation.
func (svc *ClassifiersService) Moderate(ctx context.Context, req *ModerateRequest) (*ModerationResponse, *Response, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	uri := "v1/moderations"
	httpReq, err := svc.client.NewRequest(http.MethodPost, uri, req)
	if err != nil {
		return nil, nil, err
	}

	var resp ModerationResponse
	r, err := svc.client.Do(ctx, httpReq, &resp)
	if err != nil {
		return nil, r, err
	}

	return &resp, r, nil
}

// ModerateChat performs chat moderation.
func (svc *ClassifiersService) ModerateChat(ctx context.Context, req *ModerateChatRequest) (*ModerationResponse, *Response, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	uri := "v1/chat/moderations"
	httpReq, err := svc.client.NewRequest(http.MethodPost, uri, req)
	if err != nil {
		return nil, nil, err
	}

	var resp ModerationResponse
	r, err := svc.client.Do(ctx, httpReq, &resp)
	if err != nil {
		return nil, r, err
	}

	return &resp, r, nil
}

// Classify performs text classification.
func (svc *ClassifiersService) Classify(ctx context.Context, req *ClassifyRequest) (*ClassificationResponse, *Response, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	uri := "v1/classifications"
	httpReq, err := svc.client.NewRequest(http.MethodPost, uri, req)
	if err != nil {
		return nil, nil, err
	}

	var resp ClassificationResponse
	r, err := svc.client.Do(ctx, httpReq, &resp)
	if err != nil {
		return nil, r, err
	}

	return &resp, r, nil
}

// ClassifyChat performs chat classification.
func (svc *ClassifiersService) ClassifyChat(ctx context.Context, req *ClassifyChatRequest) (*ClassificationResponse, *Response, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	uri := "v1/chat/classifications"
	httpReq, err := svc.client.NewRequest(http.MethodPost, uri, req)
	if err != nil {
		return nil, nil, err
	}

	var resp ClassificationResponse
	r, err := svc.client.Do(ctx, httpReq, &resp)
	if err != nil {
		return nil, r, err
	}

	return &resp, r, nil
}

// --- Request/Response Models ---

// ModerateRequest represents a request for text moderation.
type ModerateRequest struct {
	Input any   `json:"input"` // string or []string
	Model Model `json:"model"`
}

func (r *ModerateRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("moderate request is nil")
	}
	if r.Input == nil {
		return fmt.Errorf("input is required")
	}
	switch v := r.Input.(type) {
	case string:
		if v == "" {
			return fmt.Errorf("input cannot be empty")
		}
	case []string:
		if len(v) == 0 {
			return fmt.Errorf("input cannot be empty")
		}
	}
	return r.Model.Validate()
}

// ModerationResponse represents the response from a moderation or classification request.
type ModerationResponse struct {
	ID      string             `json:"id"`
	Model   Model              `json:"model"`
	Results []ModerationResult `json:"results"`
}

// ModerationResult represents the result for a single input in a moderation request.
type ModerationResult struct {
	Categories     map[string]bool    `json:"categories"`
	CategoryScores map[string]float64 `json:"category_scores"`
}

// ModerateChatRequest represents a request for chat moderation.
type ModerateChatRequest struct {
	Input []*ChatMessage `json:"input"`
	Model Model          `json:"model"`
}

func (r *ModerateChatRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("moderate chat request is nil")
	}
	if len(r.Input) == 0 {
		return fmt.Errorf("input messages are required")
	}
	for i, msg := range r.Input {
		if err := msg.Validate(); err != nil {
			return fmt.Errorf("invalid message at index %d: %w", i, err)
		}
	}
	return r.Model.Validate()
}

// ClassifyRequest represents a request for text classification.
type ClassifyRequest struct {
	Input any   `json:"input"` // string or []string
	Model Model `json:"model"`
}

func (r *ClassifyRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("classify request is nil")
	}
	if r.Input == nil {
		return fmt.Errorf("input is required")
	}
	switch v := r.Input.(type) {
	case string:
		if v == "" {
			return fmt.Errorf("input cannot be empty")
		}
	case []string:
		if len(v) == 0 {
			return fmt.Errorf("input cannot be empty")
		}
	}
	return r.Model.Validate()
}

// ClassificationResponse represents the response from a classification request.
type ClassificationResponse struct {
	ID      string                 `json:"id"`
	Model   Model                  `json:"model"`
	Results []ClassificationResult `json:"results"`
}

// ClassificationResult represents the result for a single input in a classification request.
type ClassificationResult struct {
	Target string  `json:"target,omitempty"`
	Score  float64 `json:"score,omitempty"`
}

// ClassifyChatRequest represents a request for chat classification.
type ClassifyChatRequest struct {
	Input ClassifyChatInput `json:"input"`
	Model Model             `json:"model"`
}

type ClassifyChatInput struct {
	Messages []*ChatMessage `json:"messages"`
}

func (r *ClassifyChatRequest) Validate() error {
	if r == nil {
		return fmt.Errorf("classify chat request is nil")
	}
	if len(r.Input.Messages) == 0 {
		return fmt.Errorf("input messages are required")
	}
	for i, msg := range r.Input.Messages {
		if err := msg.Validate(); err != nil {
			return fmt.Errorf("invalid message at index %d: %w", i, err)
		}
	}
	return r.Model.Validate()
}
