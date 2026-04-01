package mistral

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	Version          = "1.0.0"
	defaultBaseURL   = "https://api.mistral.ai/"
	defaultUserAgent = "go-mistral/" + Version
)

// Client manages communication with the Mistral API.
type Client struct {
	client    *http.Client
	BaseURL   *url.URL
	UserAgent string
	apiKey    string

	// Services
	Chat      *ChatService
	Embedding *EmbeddingService
}

type service struct {
	client *Client
}

// Option is a functional option for configuring the Client.
type ClientOption func(*Client) error

// NewClient returns a new Mistral API client.
func NewClient(opts ...ClientOption) (*Client, error) {
	baseURL, _ := url.Parse(defaultBaseURL)
	c := &Client{
		client:    &http.Client{},
		BaseURL:   baseURL,
		UserAgent: defaultUserAgent,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	// Validate that at least an API key is provided
	// (Unless a custom transport is used that handles auth)
	if c.apiKey == "" && c.client.Transport == nil {
		return nil, errors.New("cannot create client, no authentication has been provided")
	}

	c.Chat = &ChatService{client: c}
	c.Embedding = &EmbeddingService{client: c}

	return c, nil
}

// WithAPIKey sets the API key for the client.
func WithAPIKey(key string) ClientOption {
	return func(c *Client) error {
		c.apiKey = key
		return nil
	}
}

// WithBaseURL sets the base URL for the client.
func WithBaseURL(base string) ClientOption {
	return func(c *Client) error {
		if !strings.HasSuffix(base, "/") {
			base += "/"
		}
		u, err := url.Parse(base)
		if err != nil {
			return err
		}
		c.BaseURL = u
		return nil
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) error {
		if client == nil {
			return errors.New("mistral: http client is nil")
		}
		c.client = client
		return nil
	}
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client.
func (c *Client) NewRequest(method, urlStr string, body any) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		if err := enc.Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	return req, nil
}

// Response is a Mistral API response. This wraps the standard http.Response
// and provides convenient access to common fields.
type Response struct {
	*http.Response
}

// Do sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by v.
func (c *Client) Do(ctx context.Context, req *http.Request, v any) (*Response, error) {
	req = req.WithContext(ctx)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response := &Response{Response: resp}

	err = CheckResponse(resp)
	if err != nil {
		return response, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			decErr := json.NewDecoder(resp.Body).Decode(v)
			if decErr == io.EOF {
				decErr = nil // ignore EOF errors caused by empty response body
			}
			if decErr != nil {
				err = decErr
			}
		}
	}

	return response, err
}

// Stream performs a streaming API request and returns two channels for data and errors.
func Stream[T any](ctx context.Context, c *Client, req *http.Request) (<-chan T, <-chan error) {
	dataChan := make(chan T)
	errChan := make(chan error, 1)

	go func() {
		defer close(dataChan)
		defer close(errChan)

		req = req.WithContext(ctx)
		resp, err := c.client.Do(req)
		if err != nil {
			errChan <- err
			return
		}
		defer resp.Body.Close()

		if err := CheckResponse(resp); err != nil {
			errChan <- err
			return
		}

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					errChan <- err
				}
				return
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}

			var chunk T
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				errChan <- err
				return
			}

			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			case dataChan <- chunk:
			}
		}
	}()

	return dataChan, errChan
}

// APIError represents an error returned by the Mistral API.
type APIError struct {
	HTTPStatusCode int
	Message        string  `json:"message"`
	Type           string  `json:"type"`
	Param          *string `json:"param"`
	Code           *string `json:"code"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("mistral: API error (status=%d): %s", e.HTTPStatusCode, e.Message)
}

// CheckResponse checks the API response for errors, and returns them if present.
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}

	apiErr := &APIError{HTTPStatusCode: r.StatusCode}
	data, err := io.ReadAll(r.Body)
	if err == nil && len(data) > 0 {
		json.Unmarshal(data, apiErr)
	}

	return apiErr
}
