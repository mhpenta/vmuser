package urlext

import (
	"fmt"
	"net/url"
	"strings"
)

func ExtractSubdomain(urlString string) (string, error) {
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %v", err)
	}

	parts := strings.Split(parsedURL.Hostname(), ".")

	if len(parts) > 2 {
		return parts[0], nil
	}

	return "", fmt.Errorf("no subdomain found")
}
