package httpclient

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a thin wrapper around net/http used by HTTP-calling widgets.
type Client struct {
	httpClient *http.Client
	userAgent  string
}

// New returns a Client with sensible defaults (3s timeout, ccw user-agent).
func New() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 3 * time.Second},
		userAgent:  "ccw/0.1",
	}
}

// Get fetches the given URL and returns the response body bytes.
// It returns an error for non-200 status codes and network failures.
func (c *Client) Get(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
