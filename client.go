package main

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

type myClient struct {
	c           *http.Client
	t           *http.Transport
	baseUrl     string
	apiKey      string
	wsConn      *websocket.Conn
	rateLimiter *rate.Limiter
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

func newClient(baseURL, apiKey string) *myClient {
	t := setupTransport()

	return &myClient{
		c:           setupClient(t),
		t:           t,
		baseUrl:     baseURL,
		apiKey:      apiKey,
		rateLimiter: rate.NewLimiter(rate.Every(time.Second), 5),
	}
}
