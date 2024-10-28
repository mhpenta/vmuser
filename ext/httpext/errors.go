package httpext

import (
	"errors"
	"log/slog"
	"net"
	"os"
	"strings"
	"syscall"
)

const InternalServerError = "internal server error"
const BadRequestError = "bad request error"

// IsDialError determines if the given error is related to network dialing or connectivity issues.
// It checks for various types of network errors, including:
//   - Timeout errors (net.Error with Timeout() == true)
//   - Dial and read operation errors (net.OpError)
//   - Specific system errors like connection refused, host unreachable, and network unreachable
//   - DNS lookup timeout errors (net.DNSError)
//   - Generic timeout errors (detected by os.IsTimeout)
//   - String matching for common network error messages
//
// This function is useful for determining if an error is likely due to network issues
// and may be resolved by retrying the operation after a delay.
//
// Returns true if the error is identified as a network dialing or connectivity issue,
// false otherwise or if the input error is nil.
func IsDialError(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return true
		}
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if opErr.Op == "dial" || opErr.Op == "read" {
			return true
		}

		var sysErr syscall.Errno
		if errors.As(opErr.Err, &sysErr) {
			switch sysErr {
			case syscall.ECONNREFUSED, syscall.EHOSTUNREACH, syscall.ENETUNREACH, syscall.ETIMEDOUT:
				return true
			}
		}
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		slog.Warn("DNS lookup error encountered",
			"error", dnsErr,
			"isTimeout", dnsErr.IsTimeout,
			"name", dnsErr.Name)
		return dnsErr.IsTimeout
	}

	if os.IsTimeout(err) {
		return true
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "network is unreachable") ||
		strings.Contains(errMsg, "no such host") ||
		strings.Contains(errMsg, "i/o timeout")
}
