package mistral

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// APIError represents an error returned by the Mistral API.
type APIError struct {
	HTTPStatusCode int
	Message        string  `json:"message"`
	Type           string  `json:"type"`
	Param          *string `json:"param"`
	Code           *string `json:"code"`
	RawBody        []byte  `json:"-"`
}

func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" && len(e.RawBody) > 0 {
		msg = string(e.RawBody)
	}
	return fmt.Sprintf("mistral: API error (status=%d): %s", e.HTTPStatusCode, msg)
}

// CheckResponse checks the API response for errors, and returns them if present.
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}

	apiErr := &APIError{HTTPStatusCode: r.StatusCode}
	data, err := io.ReadAll(r.Body)
	if err == nil && len(data) > 0 {
		apiErr.RawBody = data
		json.Unmarshal(data, apiErr)
	}

	return apiErr
}
