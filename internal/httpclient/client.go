package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HTTPClient is a generic client for making REST requests
type HTTPClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewHTTPClient initializes a new REST client
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Get makes an HTTP GET request and decodes the response
func (c *HTTPClient) Get(endpoint string, response any) error {
	url := fmt.Sprintf("%s%s", c.BaseURL, endpoint)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to make GET request: [%w]", err)
	}
	defer resp.Body.Close()

	return c.parseResponse(resp, response)
}

// Post makes an HTTP POST request with a JSON payload
func (c *HTTPClient) Post(endpoint string, payload any, response any) error {
	url := fmt.Sprintf("%s%s", c.BaseURL, endpoint)
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: [%w]", err)
	}

	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to make POST request: [%w]", err)
	}
	defer resp.Body.Close()

	return c.parseResponse(resp, response)
}

// parseResponse handles the HTTP response and decodes the JSON
func (c *HTTPClient) parseResponse(resp *http.Response, response any) error {
	if resp.StatusCode >= 400 {
		var apiError APIError
		if err := json.NewDecoder(resp.Body).Decode(&apiError); err != nil {
			return fmt.Errorf("error response, status: [%d]", resp.StatusCode)
		}
		return fmt.Errorf("API error: [%s]", apiError.Error)
	}

	return json.NewDecoder(resp.Body).Decode(response)
}
