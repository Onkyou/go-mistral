package mistral_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onkyou/go-mistral/mistral"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *mistral.ClientConfig
		wantErr bool
		check   func(*testing.T, *mistral.Client)
	}{
		{
			name:    "default client with key",
			cfg:     &mistral.ClientConfig{APIKey: "test-key"},
			wantErr: false,
			check: func(t *testing.T, c *mistral.Client) {
				if c.BaseURL.String() != "https://api.mistral.ai/" {
					t.Errorf("expected default base URL, got %+v", c.BaseURL)
				}
			},
		},
		{
			name:    "missing auth",
			cfg:     &mistral.ClientConfig{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := mistral.NewClient(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil {
				tt.check(t, c)
			}
		})
	}
}

func TestNewRequest(t *testing.T) {
	c, _ := mistral.NewClient(&mistral.ClientConfig{APIKey: "test-key"})

	type Body struct {
		Name string `json:"name"`
	}

	tests := []struct {
		name       string
		method     string
		urlStr     string
		body       any
		wantMethod string
		checkBody  bool
	}{
		{
			name:       "basic get",
			method:     http.MethodGet,
			urlStr:     "test",
			body:       nil,
			wantMethod: http.MethodGet,
		},
		{
			name:       "post with body",
			method:     http.MethodPost,
			urlStr:     "v1/chat",
			body:       &Body{Name: "mistral"},
			wantMethod: http.MethodPost,
			checkBody:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := c.NewRequest(tt.method, tt.urlStr, tt.body)
			if err != nil {
				t.Fatalf("NewRequest returned error: %v", err)
			}

			if req.Method != tt.wantMethod {
				t.Errorf("expected method %s, got %s", tt.wantMethod, req.Method)
			}

			if tt.checkBody {
				var b Body
				if err := json.NewDecoder(req.Body).Decode(&b); err != nil {
					t.Fatalf("failed to decode request body: %v", err)
				}
			}

			if got := req.Header.Get("Authorization"); got != "Bearer test-key" {
				t.Errorf("Authorization header is %s, expected Bearer test-key", got)
			}
		})
	}
}

func TestStream_Generic(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	client, _ := mistral.NewClient(&mistral.ClientConfig{APIKey: "test-key", BaseURL: server.URL})

	mux.HandleFunc("/v1/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "data: %s\n\n", `{"id":"1"}`)
		fmt.Fprintf(w, "data: [DONE]\n\n")
	})

	req, _ := client.NewRequest("POST", "v1/stream", nil)
	type dummy struct{ ID string `json:"id"` }

	var received []string
	for chunk, err := range mistral.Stream[*dummy](context.Background(), client, req) {
		if err != nil {
			t.Fatalf("Stream() error = %v", err)
		}
		received = append(received, chunk.ID)
	}

	if len(received) != 1 || received[0] != "1" {
		t.Errorf("expected [1], got %v", received)
	}
}

func TestClient_Do_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			return
		case <-time.After(100 * time.Millisecond):
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	client, _ := mistral.NewClient(&mistral.ClientConfig{APIKey: "test-key", BaseURL: server.URL})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req, _ := client.NewRequest("GET", "v1/test", nil)
	_, err := client.Do(ctx, req, nil)

	if err == nil {
		t.Error("expected error for cancelled context, got nil")
	}
}

func TestClient_Do_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"invalid": "json`)
	}))
	defer server.Close()

	client, _ := mistral.NewClient(&mistral.ClientConfig{APIKey: "test-key", BaseURL: server.URL})

	req, _ := client.NewRequest("GET", "v1/test", nil)
	var v map[string]any
	_, err := client.Do(context.Background(), req, &v)

	if err == nil {
		t.Error("expected error for malformed JSON, got nil")
	}
}
