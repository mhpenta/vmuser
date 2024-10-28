package requests

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// IsPossibleNetworkOrDNSIssueErr analyzes the error and logs a specific warning if it detects a network or DNS resolution issue.
func IsPossibleNetworkOrDNSIssueErr(err error, url string) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	if strings.Contains(errMsg, "dial tcp") && strings.Contains(errMsg, "i/o timeout") {
		slog.Warn("Network or DNS resolution issue detected", "error", err, "url", url)
		return true
	}
	return false
}

// IsNetworkUnavailable tries to determine if a network or DNS issue might be indicating a broader internet outage.
func IsNetworkUnavailable(err error, url string) bool {
	if !IsPossibleNetworkOrDNSIssueErr(err, url) {
		return false
	}
	return isNetworkAvailableCheck()
}

// closeResponseBody safely closes the HTTP response body.
func closeResponseBody(body io.ReadCloser) {
	if err := body.Close(); err != nil {
		slog.Warn("Failed to close response body", "error", err)
	}
}

func isNetworkAvailableCheck() bool {
	urls := []string{
		"https://www.google.com",
		"https://wikipedia.org",
		"https://twitter.com/home",
		"https://www.facebook.com",
	}

	responses := make(chan bool, len(urls))

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for _, url := range urls {
		go func(url string) {
			req, _ := http.NewRequestWithContext(context.Background(), "GET", url, nil)
			resp, err := client.Do(req)
			if err == nil {
				closeResponseBody(resp.Body)
			}
			responses <- err == nil
		}(url)
	}

	for range urls {
		if <-responses {
			return true // If any request succeeds, return true immediately
		}
	}
	return false // If all requests failed, return false
}
