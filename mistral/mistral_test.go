package mistral

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		opts      []ClientOption
		wantErr   bool
		checkFunc func(*testing.T, *Client)
	}{
		{
			name:    "default client with key",
			opts:    []ClientOption{WithAPIKey("test-key")},
			wantErr: false,
			checkFunc: func(t *testing.T, c *Client) {
				if c.BaseURL.String() != defaultBaseURL {
					t.Errorf("expected default base URL, got %s", c.BaseURL)
				}
				if c.apiKey != "test-key" {
					t.Errorf("expected api key to be set")
				}
			},
		},
		{
			name:    "missing auth",
			opts:    []ClientOption{},
			wantErr: true,
		},
		{
			name:    "custom base URL",
			opts:    []ClientOption{WithAPIKey("test-key"), WithBaseURL("https://custom.mistral.ai/v1")},
			wantErr: false,
			checkFunc: func(t *testing.T, c *Client) {
				expected := "https://custom.mistral.ai/v1/"
				if c.BaseURL.String() != expected {
					t.Errorf("expected %s, got %s", expected, c.BaseURL)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, c)
			}
		})
	}
}

func TestNewRequest(t *testing.T) {
	c, _ := NewClient(WithAPIKey("test-key"))

	type Body struct {
		Name string `json:"name"`
	}

	tests := []struct {
		name       string
		method     string
		urlStr     string
		body       any
		wantURL    string
		wantMethod string
		checkBody  bool
	}{
		{
			name:       "basic get",
			method:     http.MethodGet,
			urlStr:     "test",
			body:       nil,
			wantURL:    defaultBaseURL + "test",
			wantMethod: http.MethodGet,
		},
		{
			name:       "post with body",
			method:     http.MethodPost,
			urlStr:     "v1/chat",
			body:       &Body{Name: "mistral"},
			wantURL:    defaultBaseURL + "v1/chat",
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

			if req.URL.String() != tt.wantURL {
				t.Errorf("expected URL %s, got %s", tt.wantURL, req.URL.String())
			}

			if tt.checkBody {
				var b Body
				if err := json.NewDecoder(req.Body).Decode(&b); err != nil {
					t.Fatalf("failed to decode request body: %v", err)
				}
				if b.Name != "mistral" {
					t.Errorf("expected body name mistral, got %s", b.Name)
				}
			}

			if got := req.Header.Get("Authorization"); got != "Bearer test-key" {
				t.Errorf("Authorization header is %s, expected Bearer test-key", got)
			}
		})
	}
}
