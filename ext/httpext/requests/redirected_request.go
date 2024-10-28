package requests

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// RedirectedRequest embeds RetryRequest and adds functionality to track redirects.
type RedirectedRequest struct {
	retryRequest *RetryRequest
	finalURL     url.URL
}

// NewRedirectedRequest creates a new RedirectedRequest instance.
func NewRedirectedRequest(options ...RetryRequestOption) *RedirectedRequest {
	rr := &RedirectedRequest{
		retryRequest: NewRetryRequest(options...),
	}

	originalCheckRedirect := rr.retryRequest.client.CheckRedirect
	rr.retryRequest.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		rr.finalURL = *req.URL
		if originalCheckRedirect != nil {
			return originalCheckRedirect(req, via)
		}
		if len(via) >= 10 {
			return http.ErrUseLastResponse
		}
		return nil
	}

	return rr
}

func (rr *RedirectedRequest) GetContentsAsBytesWithContextAndFinalURL(ctx context.Context, urlStr string) ([]byte, url.URL, error) {
	return rr.getContentsAsBytesWithContextAndFinalURL(ctx, urlStr, true)
}

func (rr *RedirectedRequest) getContentsAsBytesWithContextAndFinalURL(ctx context.Context, urlStr string, checkForJavaRedirect bool) ([]byte, url.URL, error) {

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, url.URL{}, fmt.Errorf("failed to parse input URL %s: %w", urlStr, err)
	}

	resp, cancel, err := rr.retryRequest.GetResponse(ctx, urlStr)
	if cancel != nil {
		defer cancel()
	}
	if err != nil {
		// TODO when this errors here, I want it to still return a url.URL based on the urlStr, if possible - can I do that?
		return nil, *parsedURL, fmt.Errorf("failed to get a response for the URL %s: %w", urlStr, err)
	}
	if resp == nil {
		return nil, *parsedURL, fmt.Errorf("failed to get a response (nil) for the URL %s", urlStr)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			slog.Error("Error closing reader", "err", err)
		}
	}(resp.Body)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, url.URL{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if checkForJavaRedirect {
		finalURLStr, found := extractJavaScriptRedirect(string(bodyBytes))
		if found {
			return rr.getContentsAsBytesWithContextAndFinalURL(ctx, finalURLStr, false)
		}
	}

	return bodyBytes, *resp.Request.URL, nil
}

func extractJavaScriptRedirect(content string) (string, bool) {

	re := regexp.MustCompile(`content="0;URL=(.+?)"`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.Trim(matches[1], "\""), true
	}

	re = regexp.MustCompile(`location\.replace\("(.+?)"\)`)
	matches = re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.Trim(matches[1], "\""), true
	}

	return "", false
}
