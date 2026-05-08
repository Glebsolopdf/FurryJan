package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	BaseURL         = "https://e621.net"
	DefaultTimeout  = 60 * time.Second
	DownloadTimeout = 30 * time.Minute
	UserAgentFormat = "Furryjan (by %s on e621)"
)

// Client is an HTTP client for e621 API
type Client struct {
	username     string
	apiKey       string
	userAgent    string
	httpClient   *http.Client
	rateLimitMS  int
	lastRequest  time.Time
	retryCount   int
	retryDelayMs int
}

func NewClient(username, apiKey string, rateLimitMS int) *Client {
	return NewClientWithTimeout(username, apiKey, rateLimitMS, DefaultTimeout)
}

func NewClientWithTimeout(username, apiKey string, rateLimitMS int, timeout time.Duration) *Client {
	return &Client{
		username:     username,
		apiKey:       apiKey,
		userAgent:    fmt.Sprintf(UserAgentFormat, username),
		httpClient:   &http.Client{Timeout: timeout},
		rateLimitMS:  rateLimitMS,
		lastRequest:  time.Now().Add(-time.Duration(rateLimitMS) * time.Millisecond),
		retryCount:   3,
		retryDelayMs: 2000,
	}
}

func (c *Client) GetPosts(tags []string, limit, page int) ([]Post, error) {
	query := url.Values{}

	if len(tags) > 0 {
		query.Set("tags", strings.Join(tags, " "))
	}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	} else {
		query.Set("limit", "320")
	}
	if page > 0 {
		query.Set("page", fmt.Sprintf("%d", page))
	}

	urlStr := fmt.Sprintf("%s/posts.json?%s", BaseURL, query.Encode())
	var resp PostsResponse

	err := c.doRequest("GET", urlStr, nil, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Posts, nil
}

func (c *Client) GetPost(postID int) (*Post, error) {
	urlStr := fmt.Sprintf("%s/posts/%d.json", BaseURL, postID)
	var resp PostResponse

	err := c.doRequest("GET", urlStr, nil, &resp)
	if err != nil {
		return nil, err
	}

	return &resp.Post, nil
}

func (c *Client) DownloadFile(fileURL string, writer io.Writer) (int64, error) {
	req, err := http.NewRequest("GET", fileURL, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("User-Agent", c.userAgent)

	c.applyRateLimit()
	resp, err := c.httpClient.Do(req)
	c.lastRequest = time.Now()
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	written, err := io.Copy(writer, resp.Body)
	return written, err
}

func (c *Client) DownloadFileWithProgress(fileURL string, expectedSize int) (*http.Response, error) {
	req, err := http.NewRequest("GET", fileURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)

	c.applyRateLimit()
	resp, err := c.httpClient.Do(req)
	c.lastRequest = time.Now()
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return resp, nil
}

func (c *Client) doRequest(method, urlStr string, body io.Reader, result interface{}) error {
	var lastErr error

	for attempt := 0; attempt < c.retryCount; attempt++ {
		req, err := http.NewRequest(method, urlStr, body)
		if err != nil {
			return err
		}

		req.Header.Set("User-Agent", c.userAgent)
		req.Header.Set("Authorization", c.basicAuth())

		c.applyRateLimit()
		resp, err := c.httpClient.Do(req)
		c.lastRequest = time.Now()
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		switch resp.StatusCode {
		case http.StatusOK, http.StatusNoContent:
			if len(responseBody) > 0 {
				if err := json.Unmarshal(responseBody, result); err != nil {
					return fmt.Errorf("failed to unmarshal response: %w", err)
				}
			}
			return nil

		case http.StatusUnauthorized:
			return fmt.Errorf("HTTP 401: Unauthorized - check your API credentials")

		case http.StatusForbidden:
			return fmt.Errorf("HTTP 403: Forbidden - check User-Agent header")

		case http.StatusTooManyRequests:
			// Rate limited - wait and retry
			if attempt < c.retryCount-1 {
				time.Sleep(5 * time.Second)
				continue
			}
			lastErr = fmt.Errorf("HTTP 429: Rate limited after retries")

		case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
			// Server error - wait and retry
			if attempt < c.retryCount-1 {
				time.Sleep(time.Duration(c.retryDelayMs) * time.Millisecond)
				continue
			}
			lastErr = fmt.Errorf("HTTP %d: Server error after retries", resp.StatusCode)

		default:
			var errResp ErrorResponse
			if err := json.Unmarshal(responseBody, &errResp); err == nil && !errResp.Success {
				return fmt.Errorf("HTTP %d: %s", resp.StatusCode, errResp.Reason)
			}
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(responseBody))
		}
	}

	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("request failed after %d attempts", c.retryCount)
}

func (c *Client) basicAuth() string {
	auth := fmt.Sprintf("%s:%s", c.username, c.apiKey)
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

func (c *Client) applyRateLimit() {
	elapsed := time.Since(c.lastRequest)
	waitTime := time.Duration(c.rateLimitMS)*time.Millisecond - elapsed
	if waitTime > 0 {
		time.Sleep(waitTime)
	}
}
