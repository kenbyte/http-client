package main

import (
	"net/http"

	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

type myClient struct {
	c           *http.Client
	baseUrl     string
	apiKey      string
	wsConn      *websocket.Conn
	rateLimiter *rate.Limiter
}
