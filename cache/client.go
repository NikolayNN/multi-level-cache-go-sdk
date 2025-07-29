package cache

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	endpointGetAll   = "/api/v1/cache/get_all"
	endpointPutAll   = "/api/v1/cache/put_all"
	endpointEvictAll = "/api/v1/cache/evict_all"
)

// Client is an HTTP client for the multi-level cache service.
type Client struct {
	baseURL       string
	httpClient    *http.Client
	gzipThreshold int
	getTimeout    time.Duration
	putTimeout    time.Duration
	evictTimeout  time.Duration
}

// New creates a client with default options.
func New(baseURL string) *Client {
	return NewWithThreshold(baseURL, 0)
}

// NewWithThreshold creates a client with gzip threshold in bytes.
func NewWithThreshold(baseURL string, gzipThreshold int) *Client {
	return NewWithOptions(baseURL, gzipThreshold,
		5*time.Second, 5*time.Second, 5*time.Second, &http.Client{})
}

// NewWithOptions creates a client with full customization.
func NewWithOptions(baseURL string, gzipThreshold int,
	getTimeout, putTimeout, evictTimeout time.Duration, httpClient *http.Client) *Client {
	if strings.HasSuffix(baseURL, "/") {
		baseURL = strings.TrimSuffix(baseURL, "/")
	}
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &Client{
		baseURL:       baseURL,
		httpClient:    httpClient,
		gzipThreshold: gzipThreshold,
		getTimeout:    getTimeout,
		putTimeout:    putTimeout,
		evictTimeout:  evictTimeout,
	}
}

// GetAll fetches multiple entries from the cache and unmarshals the values into
// the provided generic type.
func GetAll[T any](c *Client, ctx context.Context, ids []CacheId) ([]CacheEntryHit[T], error) {
	body, err := json.Marshal(ids)
	if err != nil {
		return nil, err
	}
	resp, err := c.sendRequest(ctx, endpointGetAll, c.getTimeout, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var hits []CacheEntryHit[T]
	if err := json.NewDecoder(resp.Body).Decode(&hits); err != nil {
		return nil, err
	}
	return hits, nil
}

// PutAll stores multiple entries in the cache.
func (c *Client) PutAll(ctx context.Context, entries []CacheEntry[any]) error {
	body, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	resp, err := c.sendRequest(ctx, endpointPutAll, c.putTimeout, body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// EvictAll removes multiple entries from the cache.
func (c *Client) EvictAll(ctx context.Context, ids []CacheId) error {
	body, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	resp, err := c.sendRequest(ctx, endpointEvictAll, c.evictTimeout, body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) sendRequest(ctx context.Context, endpoint string, timeout time.Duration, body []byte) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	url := c.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	payload := body
	if c.gzipThreshold > 0 && len(body) >= c.gzipThreshold {
		req.Header.Set("Content-Encoding", "gzip")
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		if _, err := gw.Write(body); err != nil {
			gw.Close()
			return nil, err
		}
		if err := gw.Close(); err != nil {
			return nil, err
		}
		payload = buf.Bytes()
	}

	req.Body = io.NopCloser(bytes.NewReader(payload))
	req.ContentLength = int64(len(payload))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected response code %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}

	return resp, nil
}
