package mistral

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	Version          = "1.0.0"
	defaultBaseURL   = "https://api.mistral.ai/"
	defaultUserAgent = "go-mistral/" + Version
	defaultTimeOut   = 5 * time.Minute
)

// Client manages communication with the Mistral API.
type Client struct {
	client    *http.Client
	BaseURL   *url.URL
	UserAgent string
	apiKey    string

	// Services
	Chat        *ChatService
	Embedding   *EmbeddingService
	Classifiers *ClassifiersService
}

type service struct {
	client *Client
}

// ClientConfig is used to configure the Mistral API client.
type ClientConfig struct {
	// APIKey is the Mistral API key.
	APIKey string

	// BaseURL is the base URL for the Mistral API.
	// Defaults to https://api.mistral.ai/
	BaseURL string

	// HTTPClient is the HTTP client used to make requests.
	// Defaults to a new http.Client.
	HTTPClient *http.Client

	// UserAgent is the User-Agent header sent with requests.
	// Defaults to go-mistral/VERSION.
	UserAgent string
}

func (cfg *ClientConfig) Validate() error {
	if cfg == nil {
		return fmt.Errorf("mistral: config is nil")
	}

	if cfg.APIKey == "" && (cfg.HTTPClient == nil || (cfg.HTTPClient != nil && cfg.HTTPClient.Transport == nil)) {
		return errors.New("cannot create client, no authentication has been provided")
	}

	return nil
}

func (cfg *ClientConfig) initializeWithDefaults() {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if !strings.HasSuffix(cfg.BaseURL, "/") {
		cfg.BaseURL += "/"
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{
			Timeout: defaultTimeOut,
		}
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = defaultUserAgent
	}
}

// NewClient returns a new Mistral API client using the provided configuration.
func NewClient(cfg *ClientConfig) (*Client, error) {
	if cfg == nil {
		cfg = &ClientConfig{}
	}

	cfg.initializeWithDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	u, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("mistral: invalid base URL: %w", err)
	}

	c := &Client{
		client:    cfg.HTTPClient,
		BaseURL:   u,
		UserAgent: cfg.UserAgent,
		apiKey:    cfg.APIKey,
	}

	c.Chat = &ChatService{client: c}
	c.Embedding = &EmbeddingService{client: c}
	c.Classifiers = &ClassifiersService{client: c}

	return c, nil
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

// Stream performs a streaming API request and returns an iterator for data and errors.
func Stream[T any](ctx context.Context, c *Client, req *http.Request) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		req = req.WithContext(ctx)
		resp, err := c.client.Do(req)
		if err != nil {
			var zero T
			yield(zero, err)
			return
		}
		defer resp.Body.Close()

		if err := CheckResponse(resp); err != nil {
			var zero T
			yield(zero, err)
			return
		}

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					var zero T
					yield(zero, err)
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
				var zero T
				yield(zero, err)
				return
			}

			if !yield(chunk, nil) {
				return
			}

			select {
			case <-ctx.Done():
				var zero T
				yield(zero, ctx.Err())
				return
			default:
			}
		}
	}
}


