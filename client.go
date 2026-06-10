package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

type Config struct {
	BaseURL     string
	APIKey      string
	AuthEnabled bool
	RateLimiter *rate.Limiter

	MaxRetries int
}

func shouldRetry(statusCode int) bool {
	return statusCode == 429 || statusCode == 502 || statusCode == 503 || statusCode == 504
}
func isRetryableNetworkError(err error) bool {
	if netErr, ok := err.(net.Error); ok {
		return netErr.Temporary() || netErr.Timeout()
	}
	return false
}

type myClient struct {
	c      *http.Client
	t      *http.Transport
	cfg    Config
	wsConn *websocket.Conn
}

func setupTransport() *http.Transport {
	return &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		DisableKeepAlives:   false,
	}
}

func setupClient(t *http.Transport) *http.Client {
	return &http.Client{
		Transport: t,
		Timeout:   30 * time.Second,
	}
}

func setupConfig(baseURL string, auth bool, r *rate.Limiter) (*Config, error) {
	cfg := &Config{
		BaseURL:     baseURL,
		AuthEnabled: auth,
		RateLimiter: r,
		MaxRetries:  5,
	}

	if auth {
		api := os.Getenv("APIKEY")
		if api == "" {
			return nil, fmt.Errorf("APIKEY required but not set")
		}
		cfg.APIKey = api
	}

	return cfg, nil
}

func newClient(baseURL string, auth bool) (*myClient, error) {
	t := setupTransport()
	r := rate.NewLimiter(rate.Every(time.Second), 5)

	cfg, err := setupConfig(baseURL, auth, r)
	if err != nil {
		return nil, err
	}

	return &myClient{
		c:   setupClient(t),
		t:   t,
		cfg: *cfg,
	}, nil
}

func (c *myClient) Do(req *http.Request) (*http.Response, error) {
	var lastErr error

	if c.cfg.RateLimiter != nil {
		if err := c.cfg.RateLimiter.Wait(req.Context()); err != nil {
			return nil, err
		}
	}

	if c.cfg.AuthEnabled {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}

	req.Header.Set("Content-Type", "application/json")

	maxAttempts := c.cfg.MaxRetries + 1

	for attempt := 0; attempt < maxAttempts; attempt++ {
		resp, err := c.c.Do(req)

		if err == nil {
			if resp.StatusCode < 400 {
				return resp, nil
			}

			if attempt < maxAttempts-1 && shouldRetry(resp.StatusCode) {
				resp.Body.Close()

				backoff := time.Second * time.Duration(1<<attempt)

				select {
				case <-time.After(backoff):
					continue
				case <-req.Context().Done():
					return nil, req.Context().Err()
				}
			}

			return resp, fmt.Errorf("http error: %d %s", resp.StatusCode, resp.Status)
		}

		lastErr = err

		if attempt < maxAttempts-1 && isRetryableNetworkError(err) {
			backoff := time.Second * time.Duration(1<<attempt)

			select {
			case <-time.After(backoff):
				continue
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}
		}

		break
	}

	return nil, fmt.Errorf(
		"request failed after %d attempts: %w",
		maxAttempts,
		lastErr,
	)
}

func (c *myClient) Get(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.BaseURL+path, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *myClient) Post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+path, body)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}
