package requests

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

func SimpleFetchBytes(urlRequestPath string) ([]byte, error) {
	parsedURL, err := url.ParseRequestURI(urlRequestPath)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	resp, err := http.Get(parsedURL.String())
	if err != nil {
		return nil, fmt.Errorf("error fetching content: %w", err)
	}
	defer func(Body io.ReadCloser) {
		errCloser := Body.Close()
		if errCloser != nil {
			slog.Error("Failed to close response body", "err", errCloser)
		}
	}(resp.Body)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	return data, nil
}
