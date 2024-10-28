package requests

import "time"

func GetTwitterShortURLFetcher() *RedirectedRequest {
	return NewRedirectedRequest(
		WithLoggedRedirects(),
		WithAttemptsAndBackoff(3, 5*time.Second),
		WithNoRetry404(),
		WithNoRetry422(),
		WithLongBackOffOn429(1*time.Minute))
}
