package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
}

func NewClient(serverURL, apiToken string) *Client {
	return &Client{
		baseURL:    serverURL + "/api/plugins/inventory-monitor",
		apiToken:   apiToken,
		httpClient: &http.Client{},
	}
}

func (c *Client) doRequest(method, path string, body any) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Token "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, resp.StatusCode, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, resp.StatusCode, nil
}

func (c *Client) Create(path string, body any, result any) error {
	respBody, _, err := c.doRequest("POST", path, body)
	if err != nil {
		return err
	}
	return json.Unmarshal(respBody, result)
}

func (c *Client) Read(path string, result any) error {
	respBody, _, err := c.doRequest("GET", path, nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(respBody, result)
}

func (c *Client) Update(path string, body any, result any) error {
	respBody, _, err := c.doRequest("PATCH", path, body)
	if err != nil {
		return err
	}
	return json.Unmarshal(respBody, result)
}

func (c *Client) Delete(path string) error {
	_, status, err := c.doRequest("DELETE", path, nil)
	if err != nil && status != 404 {
		return err
	}
	return nil
}

// nestedID is a helper used when the API returns nested objects like {"id": 1, ...}
type nestedID struct {
	ID int64 `json:"id"`
}
