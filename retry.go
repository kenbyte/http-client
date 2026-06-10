package Anzar

import (
	"net"
	"net/http"
)

func shouldRetry(statusCode int) bool {
	return statusCode == 429 ||
		statusCode == 502 ||
		statusCode == 503 ||
		statusCode == 504
}

func isRetryableNetworkError(err error) bool {
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout() || netErr.Temporary()
	}
	return false
}

func cloneRequest(req *http.Request) *http.Request {
	return req.Clone(req.Context())
}
