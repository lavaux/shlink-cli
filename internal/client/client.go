// Package client provides a typed HTTP client for the Shlink REST API.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/lavaux/shlink-cli/internal/config"
)

// Client is an authenticated HTTP client for the Shlink API.
type Client struct {
	cfg        *config.Config
	httpClient *http.Client
}

// New creates a new API client from config.
func New(cfg *config.Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// do executes an HTTP request against the Shlink API.
func (c *Client) do(method, path string, query url.Values, body interface{}) ([]byte, int, error) {
	endpoint := c.cfg.BaseURL() + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshalling request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, endpoint, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-Api-Key", c.cfg.APIKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr APIError
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Detail != "" {
			return nil, resp.StatusCode, fmt.Errorf("API error %d: %s", resp.StatusCode, apiErr.Detail)
		}
		return nil, resp.StatusCode, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, resp.StatusCode, nil
}

// Get performs a GET request.
func (c *Client) Get(path string, query url.Values) ([]byte, error) {
	b, _, err := c.do(http.MethodGet, path, query, nil)
	return b, err
}

// Post performs a POST request with a JSON body.
func (c *Client) Post(path string, body interface{}) ([]byte, error) {
	b, _, err := c.do(http.MethodPost, path, nil, body)
	return b, err
}

// Patch performs a PATCH request with a JSON body.
func (c *Client) Patch(path string, body interface{}) ([]byte, error) {
	b, _, err := c.do(http.MethodPatch, path, nil, body)
	return b, err
}

// Put performs a PUT request with a JSON body.
func (c *Client) Put(path string, body interface{}) ([]byte, error) {
	b, _, err := c.do(http.MethodPut, path, nil, body)
	return b, err
}

// Delete performs a DELETE request.
func (c *Client) Delete(path string) error {
	_, _, err := c.do(http.MethodDelete, path, nil, nil)
	return err
}

// GetHealth checks the health endpoint (no auth needed).
func (c *Client) GetHealth() ([]byte, error) {
	endpoint := c.cfg.ServerURL + "/rest/health"
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// APIError represents a Shlink API error response.
type APIError struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Status int    `json:"status"`
}
