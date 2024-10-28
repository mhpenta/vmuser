package requests

import (
	"context"
	"io"
	"net/url"
)

type Fetcher interface {
	GetContentsAsBytes(url string) ([]byte, error)
}

type FetcherReader interface {
	// GetContentsAsReader returns a reader for the contents of the URL.
	// Note: In the future we will want this to also return the size of the content
	GetContentsAsReader(url string) (io.Reader, error)
}

type FetcherWithContext interface {
	GetContentsAsBytesWithContext(ctx context.Context, url string) ([]byte, error)
}

type FetcherWithContextFromRedirect interface {
	GetContentsAsBytesWithContextAndFinalURL(ctx context.Context, url string) ([]byte, url.URL, error)
}

type FetcherBytesAndReader interface {
	GetContentsAsBytes(url string) ([]byte, error)
	GetContentsAsReader(url string) (io.Reader, error)
}
