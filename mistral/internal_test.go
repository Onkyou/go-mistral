package mistral

import (
	"testing"
)

func TestNewClient_Internal(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *ClientConfig
		wantErr bool
		check   func(*testing.T, *Client)
	}{
		{
			name:    "api key is set correctly",
			cfg:     &ClientConfig{APIKey: "test-key"},
			wantErr: false,
			check: func(t *testing.T, c *Client) {
				if c.apiKey != "test-key" {
					t.Errorf("expected api key to be set")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewClient() error = %v", err)
			}
			if tt.check != nil {
				tt.check(t, c)
			}
		})
	}
}
