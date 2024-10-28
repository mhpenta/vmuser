package requests

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// JSONLStreamFetcher represents a fetcher for JSONL streams.
type JSONLStreamFetcher struct {
	PollInterval time.Duration
	URL          string
	StartMessage *StartMessage
	EndMessage   *EndMessage
	HttpClient   *http.Client
}

// JSONLStreamFetcherOption is a function that configures a JSONLStreamFetcher.
type JSONLStreamFetcherOption func(*JSONLStreamFetcher)

// WithPollInterval returns a JSONLStreamFetcherOption that sets the polling interval.
func WithPollInterval(interval time.Duration) JSONLStreamFetcherOption {
	return func(f *JSONLStreamFetcher) {
		f.PollInterval = interval
	}
}

// WithHttpClient returns a JSONLStreamFetcherOption that sets the HTTP client.
func WithHttpClient(client *http.Client) JSONLStreamFetcherOption {
	return func(f *JSONLStreamFetcher) {
		f.HttpClient = client
	}
}

// NewJSONLStreamFetcher creates a new JSONLStreamFetcher with the given URL and options.
func NewJSONLStreamFetcher(url string, options ...JSONLStreamFetcherOption) *JSONLStreamFetcher {
	fetcher := &JSONLStreamFetcher{
		PollInterval: time.Second,
		URL:          url,
		HttpClient:   &http.Client{},
	}

	for _, option := range options {
		option(fetcher)
	}

	return fetcher
}

// FetchJSONLStream fetches the JSONL stream and returns a channel of strings representing the lines.
func (f *JSONLStreamFetcher) FetchJSONLStream(ctx context.Context) <-chan string {
	resultChan := make(chan string)

	go func() {
		defer close(resultChan)

		lastBytePosition := int64(0)

		for {
			req, err := http.NewRequestWithContext(ctx, "GET", f.URL, nil)
			if err != nil {
				slog.Error("Error creating request", "err", err)
				return
			}

			if lastBytePosition > 0 {
				req.Header.Set("Range", fmt.Sprintf("bytes=%d-", lastBytePosition))
			}

			resp, err := f.HttpClient.Do(req)
			if err != nil {
				slog.Error("Error fetching JSONL", "err", err, "url", f.URL)
				return
			}
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					slog.Error("Error closing response body", "err", err)
				}
			}(resp.Body)

			if resp.StatusCode == http.StatusPartialContent {
				scanner := bufio.NewScanner(resp.Body)
				for scanner.Scan() {
					line := scanner.Text()
					resultChan <- line

					if strings.HasPrefix(line, `{"type":"start"`) {
						var startMsg StartMessage
						if err := json.Unmarshal([]byte(line), &startMsg); err == nil {
							slog.Info("Received start of stream", "message", startMsg)
							f.StartMessage = &startMsg
						} else {
							slog.Error("Error parsing start message", "err", err)
						}
					}

					if strings.HasPrefix(line, `{"type":"end"`) {
						var endMsg EndMessage
						if err := json.Unmarshal([]byte(line), &endMsg); err == nil {
							if endMsg.Type == "end" {
								slog.Info("Received end of stream", "message", endMsg)
								f.EndMessage = &endMsg
								return
							}
						} else {
							slog.Error("Error parsing end message", "err", err)
						}
					}
				}

				if err := scanner.Err(); err != nil {
					slog.Error("Error scanning JSONL", "err", err)
					return
				}

				lastBytePosition = resp.ContentLength
			} else if resp.StatusCode == http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					slog.Error("Error reading response body", "err", err)
					return
				}

				resultChan <- string(body)
				return
			} else {
				slog.Error("Unexpected status code", "status_code", resp.StatusCode)
				return
			}

			select {
			case <-time.After(f.PollInterval):
			case <-ctx.Done():
				slog.Info("Context canceled, stopping JSONL stream fetcher")
				return
			}
		}
	}()

	return resultChan
}

// EndMessage represents the structure of the end message in the JSONL stream.
type EndMessage struct {
	Type              string `json:"type"`
	ProcessingEndTime string `json:"processing_end_time,omitempty"`
	Code              int    `json:"code,omitempty"`
	SystemReason      string `json:"system_reason,omitempty"`
	UserReason        string `json:"user_reason,omitempty"`
}

// StartMessage represents the structure of the start message in the JSONL stream.
type StartMessage struct {
	Type                string  `json:"type"`
	ProcessingStartTime string  `json:"processing_start_time,omitempty"`
	AudioURL            *string `json:"audio_url,omitempty"`
	FileFormatVersion   string  `json:"file_format_version,omitempty"`
}
