// Package requests provides utilities to make HTTP requests with features like retries, rate limiting,
// and handling of different content types.
package requests

import (
	"sync"
	"time"
	"xbrlgo/ext/httpext/headers"
)

// SECRequestInstallerRobuster wraps the RetryRequest struct to provide specific configurations suitable for
// SEC-related requests, plus a network down wait and retry policy. See SECRequest for more details.
type SECRequestInstallerRobuster struct {
	*RetryRequest
}

// singleton instance and initialization syncing
var (
	instanceSECInstaller *SECRequestInstallerRobuster
	onceSECInstaller     sync.Once
)

const (
	secRequestBackoffOn429Retry = time.Duration(601) * time.Second // 10 minutes and 1 second
)

// NewSECRequestInstallerRequest provides a global access point to the NewSECRequestInstallerRequest which has
// pre-configured settings suitable for SEC-related requests, plus a robust network unavailable retry policy.
//
// SECRequestInstallerRobuster is like SECRequest but with an additional total network down retry policy.
//
// See SECRequest for more details. SECRequestInstallerRobuster is designed to be used in long-running installation
// processes.
func NewSECRequestInstallerRequest() *SECRequestInstallerRobuster {
	onceSECInstaller.Do(func() {
		instanceSECInstaller = &SECRequestInstallerRobuster{
			NewRetryRequest(
				WithHeaders(headers.SECBotHeaders()),                                                       // SetWithBucket headers specific to SEC.
				WithAttemptsAndBackoff(Attempts, Backoff),                                                  // Configure retry attempts and backoff delay.
				WithRateLimiting(SECAttemptsPerSecond, SECBurstSize),                                       // Configure SEC policy rate limiting settings.
				WithNetworkRetryPolicy(DefaultNetworkUnavailableBackOff, DefaultNetworkUnavailableMaxWait), // Retry on major network errors.
				WithLongBackOffOn429(secRequestBackoffOn429Retry),                                          // Long backoff on 429, 10 minutes
				WithNoRetry404(),                                                                           // Break on 404, do not retry - let's not annoy the SEC
			),
		}
	})
	return instanceSECInstaller
}
