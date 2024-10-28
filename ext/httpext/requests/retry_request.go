package requests

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"golang.org/x/net/html/charset"
	"golang.org/x/time/rate"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrNetworkUnavailableAfterMaxWait = errors.New("network unavailable after max wait")

type StatusCodeError struct {
	StatusCode int
	URL        string
	Message    string
}

func (e *StatusCodeError) Error() string {
	return fmt.Sprintf("%s: %s", e.Message, e.URL)
}

var (
	ErrNotFound            = &StatusCodeError{StatusCode: http.StatusNotFound, Message: "404 Not Found"}
	ErrUnprocessableEntity = &StatusCodeError{StatusCode: http.StatusUnprocessableEntity, Message: "422 Unprocessable Entity"}
	ErrNotFoundNoRetry     = errors.New("404 Not Found, not retrying due to NoRetry404 option")
)

// Constants used for default configurations.
const (
	DefaultUserAgent                 = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:106.0) Gecko/20100101 Firefox/106.0"
	DefaultMaxRetries                = 5
	DefaultBackoffFactor             = 3 * time.Second
	DefaultRequestTimeout            = 60 * time.Second
	DefaultNetworkUnavailableBackOff = 5 * time.Minute
	DefaultNetworkUnavailableMaxWait = 6 * time.Hour
)

// RetryRequest struct encapsulates configuration for making HTTP requests with retry and rate limiting functionality.
type RetryRequest struct {
	headers          http.Header
	maxRetries       int
	backoffFactor    time.Duration
	client           *http.Client
	limiter          *rate.Limiter
	isRateLimited    bool
	requestTimeout   time.Duration
	noRetry404       bool
	noRetry422       bool
	longBackOffOn429 time.Duration

	resolveNetworkUnavailable bool
	networkUnavailableBackOff time.Duration
	networkUnavailableMaxWait time.Duration
}

// RetryRequestOption represents a functional option type for configuring the RetryRequest.
type RetryRequestOption func(*RetryRequest)

// WithHeaders provides custom headers for the HTTP request.
func WithHeaders(headers http.Header) RetryRequestOption {
	return func(r *RetryRequest) {
		r.headers = headers
	}
}

// WithRequestTimeout provides custom request timeout for the HTTP request.
func WithRequestTimeout(requestTimeout time.Duration) RetryRequestOption {
	return func(r *RetryRequest) {
		r.requestTimeout = requestTimeout
	}
}

// WithAttemptsAndBackoff configures the maximum number of retry attempts and backoff delay.
func WithAttemptsAndBackoff(attempts int, backoff time.Duration) RetryRequestOption {
	return func(r *RetryRequest) {
		r.maxRetries = attempts
		r.backoffFactor = backoff
	}
}

// WithRateLimiting configures rate limiting for the HTTP requests.
func WithRateLimiting(limit rate.Limit, burst int) RetryRequestOption {
	return func(r *RetryRequest) {
		r.limiter = rate.NewLimiter(limit, burst)
		r.isRateLimited = true
	}
}

// WithNoRetry404 configures the request to not retry on 404 Not Found errors.
func WithNoRetry404() RetryRequestOption {
	return func(r *RetryRequest) {
		r.noRetry404 = true
	}
}

// WithNoRetry422 configures the request to not retry on 422 Unprocessable Entity errors.
func WithNoRetry422() RetryRequestOption {
	return func(r *RetryRequest) {
		r.noRetry422 = true
	}
}

// WithNetworkRetryPolicy configures the backoff delay and maximum wait time for retrying requests when
// the network is completely unavailable. This is useful for scenarios where network connectivity is intermittent
// or unreliable, allowing the application to intelligently delay retries until the maxWaitTime is reached.
func WithNetworkRetryPolicy(networkUnavailableBackOff time.Duration, maxWaitTime time.Duration) RetryRequestOption {
	return func(r *RetryRequest) {
		r.resolveNetworkUnavailable = true
		r.networkUnavailableBackOff = networkUnavailableBackOff
		r.networkUnavailableMaxWait = maxWaitTime
	}
}

// WithLongBackOffOn429 configures the backoff delay for retrying requests when a 429 Too Many Requests status code is received.
func WithLongBackOffOn429(backoff time.Duration) RetryRequestOption {
	return func(r *RetryRequest) {
		r.longBackOffOn429 = backoff
	}
}

// WithLoggedRedirects configures the request to log redirects using slog.
func WithLoggedRedirects() RetryRequestOption {
	return func(r *RetryRequest) {
		r.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			slog.Info("Redirecting request", "url", req.URL.String())
			return nil
		}
	}
}

// NewRetryRequest initializes a new RetryRequest instance using provided options.
func NewRetryRequest(options ...RetryRequestOption) *RetryRequest {
	r := &RetryRequest{
		headers:        make(http.Header),
		maxRetries:     DefaultMaxRetries,
		backoffFactor:  DefaultBackoffFactor,
		requestTimeout: DefaultRequestTimeout,
		client:         &http.Client{},
	}

	r.headers.Set("User-Agent", DefaultUserAgent)

	for _, opt := range options {
		opt(r)
	}

	return r
}

func (r *RetryRequest) createRequestAndGetResponse(ctx context.Context, url string) (*http.Response, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(ctx, r.requestTimeout)
	req, reqErr := http.NewRequestWithContext(ctx, "GET", url, nil)
	if reqErr != nil {
		cancel()
		return nil, nil, reqErr
	}
	req.Header = r.headers
	resp, err := r.client.Do(req)
	return resp, cancel, err
}

// GetResponse sends an HTTP GET request to the specified URL with retries on failures.
func (r *RetryRequest) GetResponse(ctx context.Context, url string) (*http.Response, context.CancelFunc, error) {
	// Note, this rate limiter is at the start of the request. This works as a general rule so long as the backoff
	// time is less than the rate limiter time.
	if r.isRateLimited {
		err := r.limiter.Wait(ctx)
		if err != nil {
			return nil, nil, err
		}
	}

	var resp *http.Response
	var err error
	var cancel context.CancelFunc
	for i := 0; i < r.maxRetries; i++ {
		resp, cancel, err = r.createRequestAndGetResponse(ctx, url)
		if err == nil {
			if resp.StatusCode == http.StatusNotFound && r.noRetry404 {
				return resp, cancel, fmt.Errorf("%w: %s", ErrNotFoundNoRetry, url)
			}
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				// done, return response
				return resp, cancel, nil
			}
		}

		if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
			cancel()
		}

		if resp != nil {
			closeErr := resp.Body.Close()
			if closeErr != nil {
				slog.Error("Failed to close response body, potential leak, continuing", "err", closeErr)
			}
		}

		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, nil, context.Canceled
		}

		if r.resolveNetworkUnavailable && i == r.maxRetries-1 {
			// if it is the last attempt, check network if WithNetworkRetryPolicy is set
			if IsNetworkUnavailable(err, url) {
				start := time.Now()
				for {
					remainingTime := r.networkUnavailableMaxWait - time.Since(start)
					if remainingTime <= 0 {
						return nil, nil, ErrNetworkUnavailableAfterMaxWait
					}

					sleepDuration := min(remainingTime, r.networkUnavailableBackOff)
					time.Sleep(sleepDuration)

					resp, cancel, err = r.createRequestAndGetResponse(ctx, url)
					if err == nil {
						if resp.StatusCode == http.StatusNotFound && r.noRetry404 {
							return resp, cancel, &StatusCodeError{
								StatusCode: resp.StatusCode,
								URL:        url,
								Message:    ErrNotFound.Message,
							}
						}
						if resp.StatusCode == http.StatusUnprocessableEntity && r.noRetry422 {
							return resp, cancel, &StatusCodeError{
								StatusCode: resp.StatusCode,
								URL:        url,
								Message:    ErrUnprocessableEntity.Message,
							}
						}
						if resp.StatusCode >= 200 && resp.StatusCode < 300 {
							// done, return response
							return resp, cancel, nil
						}
					}

					cancel()
					if resp != nil {
						closeErr := resp.Body.Close()
						if closeErr != nil {
							slog.Error("Failed to close response body, potential leak, continuing", "err", closeErr)
						}
					}

					if err != nil {
						// If the new error is not a network or DNS issue, return immediately
						if !IsPossibleNetworkOrDNSIssueErr(err, url) {
							return nil, nil, err
						}
					}
				}
			}
			continue
		}

		// Wait for exponential backoff
		//time.Sleep(r.backoffFactor * time.Duration(1<<i))
		//if resp != nil {
		//	slog.Info("Retrying request", "url", url, "attempt", i+1, "maxRetries", r.maxRetries, "lastError", err, "responseStatusCode", resp.StatusCode, "responseStatus", resp.Status, "responseHeader", resp.Header)
		//} else {
		//	slog.Info("Retrying request", "url", url, "attempt", i+1, "maxRetries", r.maxRetries, "lastError", err)
		//}
		if err := r.backoff(ctx, i, url, err, resp); err != nil {
			return nil, nil, err
		}
	}

	// If here, all retries failed
	return nil, nil, fmt.Errorf("max retries reached: last error: %w", err)
}

func (r *RetryRequest) fetchContentsAsBytes(ctx context.Context, url string) ([]byte, error) {
	var bodyBytes []byte
	var err error

	for attempt := 0; attempt < r.maxRetries; attempt++ {
		bodyBytes, err = r.attemptFetchContents(ctx, url)
		if err == nil {
			return bodyBytes, nil
		}

		if strings.Contains(err.Error(), "stream error") {
			slog.Info("Encountered stream error, will retry",
				"url", url,
				"attempt", attempt+1,
				"maxRetries", r.maxRetries,
				"error", err)

			if err := r.backoff(ctx, attempt, url, err, nil); err != nil {
				return nil, err
			}
			continue
		}
		return nil, err
	}
	return nil, fmt.Errorf("max retries reached: last error: %w", err)
}

func (r *RetryRequest) attemptFetchContents(ctx context.Context, url string) ([]byte, error) {
	resp, cancel, err := r.GetResponse(ctx, url)
	if cancel != nil {
		defer cancel()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get a response for the URL %s: %w", url, err)
	}
	if resp == nil {
		return nil, fmt.Errorf("failed to get a response (nil) for the URL %s", url)
	}
	defer func() {
		if resp.Body != nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				slog.Error("Failed to close response body", "err", closeErr)
			}
		}
	}()

	var reader io.Reader = resp.Body

	// Handle gzip encoding if present
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, gzipReaderError := gzip.NewReader(resp.Body)
		if gzipReaderError != nil {
			slog.Error("Failed to create gzip reader", "err", gzipReaderError)
			return nil, gzipReaderError
		}
		defer func() {
			if gzipReader != nil {
				if errLeak := gzipReader.Close(); errLeak != nil {
					slog.Error("Failed to close gzip reader, potential leak", "err", errLeak)
				}
			}
		}()
		reader = gzipReader
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "text/") || strings.Contains(contentType, "json") || strings.Contains(contentType, "xml") {
		decodedReader, err := charset.NewReader(reader, contentType)
		if err != nil {
			slog.Error("Failed to decode response content", "err", err)
			return nil, err
		}
		return io.ReadAll(decodedReader)
	} else {
		// For binary data, read raw bytes directly
		return io.ReadAll(reader)
	}
}

// GetContents sends an HTTP GET request to retrieve content from the specified URL, handling gzip encoding if present.
func (r *RetryRequest) GetContents(url string) (string, error) {
	bodyBytes, err := r.fetchContentsAsBytes(context.Background(), url)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

// GetContentsAsBytes sends an HTTP GET request to retrieve content from the specified URL, handling gzip encoding if present.
func (r *RetryRequest) GetContentsAsBytes(url string) ([]byte, error) {
	bodyBytes, err := r.fetchContentsAsBytes(context.Background(), url)
	if err != nil {
		return nil, err
	}
	return bodyBytes, nil
}

// GetContentsAsBytesWithContext sends an HTTP GET request to retrieve content from the specified URL, handling gzip encoding if present.
func (r *RetryRequest) GetContentsAsBytesWithContext(ctx context.Context, url string) ([]byte, error) {
	bodyBytes, err := r.fetchContentsAsBytes(ctx, url)
	if err != nil {
		return nil, err
	}
	return bodyBytes, nil
}

// GetContentFromURL sends an HTTP GET request to retrieve content from the specified url.URL,
// handling gzip encoding if present. We immediately convert the url to a string because that is required for
// http.NewRequestWithContext where it is subsequently (and unfortunately) converted back to a url.URL.
func (r *RetryRequest) GetContentFromURL(url *url.URL) ([]byte, error) {
	bodyBytes, err := r.fetchContentsAsBytes(context.Background(), url.String())
	if err != nil {
		return nil, err
	}
	return bodyBytes, nil
}

// PostContentsAsBytes sends an HTTP Post request to retrieve content from the specified URL, handling gzip encoding if present.
func (r *RetryRequest) PostContentsAsBytes(url string, reader io.Reader) ([]byte, error) {
	bodyBytes, err := r.fetchContentsAsBytesPost(url, reader)
	if err != nil {
		return nil, err
	}
	return bodyBytes, nil
}

// GetCSV sends an HTTP GET request to retrieve CSV content from the specified URL.
func (r *RetryRequest) GetCSV(url string) (string, error) {
	resp, cancel, err := r.GetResponse(context.Background(), url)
	defer cancel()
	if err != nil || resp == nil {
		return "", fmt.Errorf("failed to get a csv response for the URL: %w", err)
	}
	defer func(Body io.ReadCloser) {
		closeErr := Body.Close()
		if closeErr != nil {
			slog.Error("Failed to close response body", "err", closeErr)
		}
	}(resp.Body)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read response content", "err", err)
		return "", err
	}

	return string(bodyBytes), nil
}

// SendPostRequest sends an HTTP POST request to the specified URL with retries on failures.
// The body parameter is the data to be sent in the POST request.
func (r *RetryRequest) SendPostRequest(url string, body io.Reader) (*http.Response, context.CancelFunc, error) {
	if r.isRateLimited {
		err := r.limiter.Wait(context.Background())
		if err != nil {
			return nil, nil, err
		}
	}

	var resp *http.Response
	var err error

	for i := 0; i < r.maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), r.requestTimeout)
		req, reqErr := http.NewRequestWithContext(ctx, "POST", url, body)
		if reqErr != nil {
			cancel()
			return nil, nil, reqErr
		}

		req.Header = r.headers
		resp, err = r.client.Do(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Successful request
			return resp, cancel, nil
		}
		cancel()

		if resp != nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				slog.Error("Failed to close response body, potential leak", "error", closeErr)
			}
		}

		// Delay for exponential backoff
		time.Sleep(r.backoffFactor * time.Duration(1<<i))
		slog.Info("Retrying POST request", "url", url, "attempt", i+1, "maxRetries", r.maxRetries)
	}

	// If reached here, all retries failed
	return nil, nil, fmt.Errorf("failed after max retries: last error: %w", err)
}

// fetchContentsAsBytes sends an HTTP GET request to retrieve content from the specified URL,
// handling gzip encoding if present, and returns content as bytes.
func (r *RetryRequest) fetchContentsAsBytesPost(url string, body io.Reader) ([]byte, error) {
	resp, cancel, err := r.SendPostRequest(url, body)
	if cancel != nil {
		defer cancel()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get a response for the URL: %w", err)
	}
	if resp == nil {
		return nil, fmt.Errorf("received a nil response for the URL")
	}
	defer func() {
		if resp.Body != nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				slog.Error("Failed to close response body", "err", closeErr)
			}
		}
	}()

	var reader io.Reader = resp.Body
	var bodyBytes []byte

	// Handle gzip encoding if present
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, gzipReaderError := gzip.NewReader(resp.Body)
		if gzipReaderError != nil {
			slog.Error("Failed to create gzip reader", "err", err)
			return nil, gzipReaderError
		}
		defer func() {
			if gzipReader != nil {
				if errLeak := gzipReader.Close(); errLeak != nil {
					slog.Error("Failed to close gzip reader, potential leak", "err", errLeak)
				}
			}
		}()
		reader = gzipReader
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "text/") || strings.Contains(contentType, "json") || strings.Contains(contentType, "xml") {
		decodedReader, err := charset.NewReader(reader, contentType)
		if err != nil {
			slog.Error("Failed to decode response content", "err", err)
			return nil, err
		}
		bodyBytes, err = io.ReadAll(decodedReader)
		if err != nil {
			slog.Error("Failed to read response content", "err", err)
			return nil, err
		}
	} else {
		// For binary data, read raw bytes directly
		bodyBytes, err = io.ReadAll(reader)
		if err != nil {
			slog.Error("Failed to read response content", "err", err)
			return nil, err
		}
	}

	return bodyBytes, nil
}

// GetContentsAsReader sends an HTTP GET request to retrieve content from the specified URL and returns an io.Reader
// Note: In the future, we will want to have this return the content size from the response
func (r *RetryRequest) GetContentsAsReader(url string) (io.Reader, error) {
	reader, err := r.fetchContentsAsReader(url)
	if err != nil {
		return nil, err
	}
	return reader, nil
}

func (r *RetryRequest) fetchContentsAsReader(url string) (io.Reader, error) {
	resp, _, err := r.GetResponse(context.Background(), url)
	if err != nil {
		return nil, fmt.Errorf("failed to get a response for the URL %s: %w", url, err)
	}
	if resp == nil {
		return nil, fmt.Errorf("failed to get a response (nil) for the URL %s", url)
	}

	var reader io.Reader = resp.Body

	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, gzipReaderError := gzip.NewReader(resp.Body)
		if gzipReaderError != nil {
			slog.Error("Failed to create gzip reader", "err", gzipReaderError)
			return nil, gzipReaderError
		}
		reader = gzipReader
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "text/") || strings.Contains(contentType, "json") || strings.Contains(contentType, "xml") {
		decodedReader, err := charset.NewReader(reader, contentType)
		if err != nil {
			slog.Error("Failed to decode response content", "err", err)
			return nil, err
		}
		reader = decodedReader
	}

	return reader, nil
}

func (r *RetryRequest) backoff(
	ctx context.Context,
	attempt int,
	url string,
	lastError error,
	resp *http.Response) error {

	backoffDuration := r.backoffFactor * time.Duration(1<<attempt)

	logMessage := "Retrying request after backoff"

	if resp != nil && resp.StatusCode == http.StatusTooManyRequests && r.longBackOffOn429 > backoffDuration {
		backoffDuration = r.longBackOffOn429
		logMessage = "Retrying request after long backoff on 429"
	}

	// Log before waiting
	if resp != nil {
		slog.Info(logMessage,
			"url", url,
			"attempt", attempt+1,
			"maxRetries", r.maxRetries,
			"backoffDuration", backoffDuration,
			"lastError", lastError,
			"responseStatusCode", resp.StatusCode,
			"responseStatus", resp.Status,
			"responseHeader", resp.Header)
	} else {
		slog.Info(logMessage,
			"url", url,
			"attempt", attempt+1,
			"maxRetries", r.maxRetries,
			"backoffDuration", backoffDuration,
			"lastError", lastError)
	}

	timer := time.NewTimer(backoffDuration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// fetchContentsAsBytes sends an HTTP GET request to retrieve content from the specified URL,
// handling gzip encoding if present, and returns content as bytes.
func (r *RetryRequest) fetchContentsAsBytesV1(ctx context.Context, url string) ([]byte, error) {
	resp, cancel, err := r.GetResponse(ctx, url)
	if cancel != nil {
		defer cancel()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get a response for the URL %s: %w", url, err)
	}
	if resp == nil {
		return nil, fmt.Errorf("failed to get a response (nil) for the URL %s", url)
	}
	defer func() {
		if resp.Body != nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				slog.Error("Failed to close response body", "err", closeErr)
			}
		}
	}()

	var reader io.Reader = resp.Body
	var bodyBytes []byte

	// Handle gzip encoding if present
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, gzipReaderError := gzip.NewReader(resp.Body)
		if gzipReaderError != nil {
			slog.Error("Failed to create gzip reader", "err", gzipReaderError)
			return nil, gzipReaderError
		}
		defer func() {
			if gzipReader != nil {
				if errLeak := gzipReader.Close(); errLeak != nil {
					slog.Error("Failed to close gzip reader, potential leak", "err", errLeak)
				}
			}
		}()
		reader = gzipReader
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "text/") || strings.Contains(contentType, "json") || strings.Contains(contentType, "xml") {
		decodedReader, err := charset.NewReader(reader, contentType)
		if err != nil {
			slog.Error("Failed to decode response content", "err", err)
			return nil, err
		}
		bodyBytes, err = io.ReadAll(decodedReader)
		if err != nil {
			return nil, fmt.Errorf("failed to read response content: %w", err)
		}
	} else {
		// For binary data, read raw bytes directly
		bodyBytes, err = io.ReadAll(reader)
		if err != nil {
			slog.Error("Failed to read response content", "err", err)
			return nil, err
		}
	}

	return bodyBytes, nil
}
