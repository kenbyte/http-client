package Anzar

import (
	"net/http"
	"time"
)

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
