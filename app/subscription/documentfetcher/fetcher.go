package documentfetcher

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/app/subscription"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

type FetcherOptions interface{}

type Fetcher interface {
	DownloadDocument(ctx context.Context, source *subscription.ImportSource, opts ...FetcherOptions) ([]byte, error)
}

var knownFetcher = make(map[string]Fetcher)

func RegisterFetcher(name string, fetcher Fetcher) error {
	if _, found := knownFetcher[name]; found {
		return newError("fetcher ", name, " already registered")
	}
	knownFetcher[name] = fetcher
	return nil
}

func GetFetcher(name string) (Fetcher, error) {
	if fetcher, found := knownFetcher[name]; found {
		return fetcher, nil
	}
	return nil, newError("fetcher ", name, " not found")
}
