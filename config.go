package Anzar

import (
	"fmt"
	"os"

	"golang.org/x/time/rate"
)

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type Config struct {
	BaseURL     string
	APIKey      string
	AuthEnabled bool
	RateLimiter *rate.Limiter
	MaxRetries  int
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
