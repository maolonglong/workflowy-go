package workflowy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	defaultBaseURL   = "https://workflowy.com/api/v1"
	defaultUserAgent = "workflowy-go/0.1.0"
)

// Client is the top-level WorkFlowy API client.
type Client struct {
	Nodes   *NodesService
	Targets *TargetsService

	httpClient *http.Client
	baseURL    string
	apiKey     string
	userAgent  string
}

// NewClient creates a new WorkFlowy client with the given options.
func NewClient(opts ...Option) (*Client, error) {
	c := &Client{
		httpClient: http.DefaultClient,
		baseURL:    defaultBaseURL,
		userAgent:  defaultUserAgent,
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	s := &service{client: c}
	c.Nodes = &NodesService{s: s}
	c.Targets = &TargetsService{s: s}

	return c, nil
}

// Option configures a Client.
type Option func(*Client)

// WithAPIKey sets the API key for authentication.
func WithAPIKey(key string) Option {
	return func(c *Client) { c.apiKey = key }
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// WithBaseURL overrides the default API base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// WithUserAgent sets a custom User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *Client) { c.userAgent = ua }
}

// service is shared infrastructure for all sub-services.
type service struct {
	client *Client
}

func (s *service) newRequest(method, path string, body any) (*http.Request, error) {
	u, err := url.JoinPath(s.client.baseURL, path)
	if err != nil {
		return nil, fmt.Errorf("workflowy: invalid URL: %w", err)
	}

	var req *http.Request
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("workflowy: marshal request body: %w", err)
		}
		req, err = http.NewRequest(method, u, bytes.NewReader(buf))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, u, nil)
		if err != nil {
			return nil, err
		}
	}

	req.Header.Set("Authorization", "Bearer "+s.client.apiKey)
	req.Header.Set("User-Agent", s.client.userAgent)
	return req, nil
}

func (s *service) do(ctx context.Context, req *http.Request, v any) error {
	resp, err := s.client.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("workflowy: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return parseAPIError(resp)
	}

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return fmt.Errorf("workflowy: decode response: %w", err)
		}
	}
	return nil
}
