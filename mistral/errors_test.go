package mistral_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/onkyou/go-mistral/mistral"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		apiErr   *mistral.APIError
		expected string
	}{
		{
			name: "with message",
			apiErr: &mistral.APIError{
				HTTPStatusCode: 400,
				Message:        "Invalid parameter",
			},
			expected: "mistral: API error (status=400): Invalid parameter",
		},
		{
			name: "without message, with raw body",
			apiErr: &mistral.APIError{
				HTTPStatusCode: 500,
				RawBody:        []byte("Internal server error details"),
			},
			expected: "mistral: API error (status=500): Internal server error details",
		},
		{
			name: "without message, without raw body",
			apiErr: &mistral.APIError{
				HTTPStatusCode: 404,
			},
			expected: "mistral: API error (status=404): ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.apiErr.Error(); got != tt.expected {
				t.Errorf("APIError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCheckResponse(t *testing.T) {
	tests := []struct {
		name       string
		resp       *http.Response
		wantErr    bool
		checkError func(*testing.T, error)
	}{
		{
			name: "200 OK",
			resp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"status":"ok"}`)),
			},
			wantErr: false,
		},
		{
			name: "204 No Content",
			resp: &http.Response{
				StatusCode: http.StatusNoContent,
				Body:       io.NopCloser(bytes.NewBufferString("")),
			},
			wantErr: false,
		},
		{
			name: "401 Unauthorized",
			resp: &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(bytes.NewBufferString(`{"message":"Unauthorized access","type":"auth_error"}`)),
			},
			wantErr: true,
			checkError: func(t *testing.T, err error) {
				apiErr, ok := err.(*mistral.APIError)
				if !ok {
					t.Fatalf("expected *mistral.APIError, got %T", err)
				}
				if apiErr.HTTPStatusCode != 401 {
					t.Errorf("expected status 401, got %d", apiErr.HTTPStatusCode)
				}
				if apiErr.Message != "Unauthorized access" {
					t.Errorf("expected message 'Unauthorized access', got %s", apiErr.Message)
				}
				if apiErr.Type != "auth_error" {
					t.Errorf("expected type 'auth_error', got %s", apiErr.Type)
				}
			},
		},
		{
			name: "500 Internal Server Error (text)",
			resp: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBufferString(`Something went wrong`)),
			},
			wantErr: true,
			checkError: func(t *testing.T, err error) {
				apiErr, ok := err.(*mistral.APIError)
				if !ok {
					t.Fatalf("expected *mistral.APIError, got %T", err)
				}
				if string(apiErr.RawBody) != "Something went wrong" {
					t.Errorf("expected raw body 'Something went wrong', got %s", string(apiErr.RawBody))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mistral.CheckResponse(tt.resp)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkError != nil {
				tt.checkError(t, err)
			}
		})
	}
}
