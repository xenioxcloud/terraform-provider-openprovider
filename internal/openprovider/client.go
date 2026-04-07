package openprovider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const baseURL = "https://api.openprovider.eu/v1beta"

// Client is the Openprovider API client.
type Client struct {
	username   string
	password   string
	httpClient *http.Client

	mu    sync.Mutex
	token string
}

// NewClient creates a new Openprovider API client.
func NewClient(username, password string) *Client {
	return &Client{
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// authenticate obtains a bearer token from the Openprovider API.
func (c *Client) authenticate() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	body := map[string]string{
		"username": c.username,
		"password": c.password,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling auth request: %w", err)
	}

	resp, err := c.httpClient.Post(baseURL+"/auth/login", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("auth failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decoding auth response: %w", err)
	}
	if result.Data.Token == "" {
		return fmt.Errorf("empty token in auth response")
	}

	c.token = result.Data.Token
	return nil
}

// getToken returns the current token, authenticating if needed.
func (c *Client) getToken() (string, error) {
	c.mu.Lock()
	token := c.token
	c.mu.Unlock()

	if token == "" {
		if err := c.authenticate(); err != nil {
			return "", err
		}
		c.mu.Lock()
		token = c.token
		c.mu.Unlock()
	}
	return token, nil
}

// doRequest performs an authenticated HTTP request.
func (c *Client) doRequest(method, path string, body any) ([]byte, error) {
	token, err := c.getToken()
	if err != nil {
		return nil, err
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	// Re-authenticate on 401 and retry once
	if resp.StatusCode == http.StatusUnauthorized {
		if err := c.authenticate(); err != nil {
			return nil, fmt.Errorf("re-authentication failed: %w", err)
		}
		return c.doRequest(method, path, body)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
