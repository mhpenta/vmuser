// Package requests provides utilities to make HTTP requests with features like retries, rate limiting,
// and handling of different content types.
package requests

import (
	"sync"
	"time"
	"xbrlgo/ext/httpext/headers"
)

// SECRequest wraps the RetryRequest struct to provide specific configurations suitable for SEC-related requests.
// It encapsulates settings such as headers specific to the SEC, rate limiting, and S.E.C. related retry strategies.
type SECRequest struct {
	*RetryRequest
}

// singleton instance and initialization syncing
var (
	instance *SECRequest
	once     sync.Once
)

// Constants used for SEC request configurations.
const (
	SECAttemptsPerSecond = 10               // Number of retry attempts allowed per second.
	SECBurstSize         = 10               // Maximum number of events taken out of the rate limiter in a single burst.
	Attempts             = 7                // Maximum number of attempts to make for a request.
	Backoff              = 10 * time.Second // Duration to wait between retry attempts.
)

// NewSECRequest provides a global access point to the SECRequest which has pre-configured settings suitable for SEC-related requests.
// It initializes a singleton instance of the SECRequest struct if it hasn't already been initialized.
// It sets specific headers for the SEC, sets the number of retry attempts, backoff delay, and rate limiting configurations.
// As of July 27, 2021, the SEC limits automated searches to a total of no more than 10 requests per second.
func NewSECRequest() *SECRequest {
	once.Do(func() {
		instance = &SECRequest{
			NewRetryRequest(
				WithHeaders(headers.SECBotHeaders()),                 // SetWithBucket headers specific to SEC.
				WithAttemptsAndBackoff(Attempts, Backoff),            // Configure retry attempts and backoff delay.
				WithRateLimiting(SECAttemptsPerSecond, SECBurstSize), // Configure SEC policy rate limiting settings.
				WithLongBackOffOn429(secRequestBackoffOn429Retry),    // Long backoff on 429, 10 minutes
				WithNoRetry404(),                                     // Break on 404, do not retry - let's not annoy the SEC
			),
		}
	})
	return instance
}
