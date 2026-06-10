package Anzar

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

type Client struct {
	c      *http.Client
	t      *http.Transport
	cfg    Config
	wsConn *websocket.Conn
}

func NewClient(baseURL string, auth bool) (*Client, error) {
	t := setupTransport()

	r := rate.NewLimiter(
		rate.Every(time.Second),
		5,
	)

	cfg, err := setupConfig(
		baseURL,
		auth,
		r,
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		c:   setupClient(t),
		t:   t,
		cfg: *cfg,
	}, nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	var lastErr error

	if c.cfg.RateLimiter != nil {
		if err := c.cfg.RateLimiter.Wait(req.Context()); err != nil {
			return nil, err
		}
	}

	if c.cfg.AuthEnabled {
		req.Header.Set(
			"Authorization",
			"Bearer "+c.cfg.APIKey,
		)
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	maxAttempts := c.cfg.MaxRetries + 1

	retryAllowed :=
		req.Method == http.MethodGet ||
			req.Method == http.MethodHead

	for attempt := 0; attempt < maxAttempts; attempt++ {
		reqCopy := cloneRequest(req)

		resp, err := c.c.Do(reqCopy)

		if err == nil {
			if resp.StatusCode < 400 {
				return resp, nil
			}

			if !retryAllowed {
				return resp, fmt.Errorf(
					"http error: %d %s",
					resp.StatusCode,
					resp.Status,
				)
			}

			if attempt < maxAttempts-1 &&
				shouldRetry(resp.StatusCode) {
				resp.Body.Close()

				backoff :=
					time.Second *
						time.Duration(1<<attempt)

				select {
				case <-time.After(backoff):
					continue
				case <-req.Context().Done():
					return nil, req.Context().Err()
				}
			}

			return resp, fmt.Errorf(
				"http error: %d %s",
				resp.StatusCode,
				resp.Status,
			)
		}

		lastErr = err

		if !retryAllowed {
			break
		}

		if attempt < maxAttempts-1 &&
			isRetryableNetworkError(err) {
			backoff :=
				time.Second *
					time.Duration(1<<attempt)

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

func (c *Client) Get(
	ctx context.Context,
	path string,
) (*http.Response, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.cfg.BaseURL+path,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

func (c *Client) Post(
	ctx context.Context,
	path string,
	body io.Reader,
) (*http.Response, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.cfg.BaseURL+path,
		body,
	)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}
