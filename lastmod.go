package bulk

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	ErrRequestFailed = errors.New("http request failed")
)

// FetchLastModDatesForURLs fetches the last modification date for multiple urls at once.
func FetchLastModDatesForURLs(
	ctx context.Context, executor Executor, modifyRequest func(r *http.Request) error, urls ...string,
) ([]time.Time, error) {
	if len(urls) > 0 {
		initialContext, cancel := context.WithCancel(ctx)
		defer cancel()

		results := executor.AddRequestsWithInterceptor(initialContext, modifyRequest, urls...)

		times := make([]time.Time, len(urls))
		for i, result := range results {
			lastModified, err := handleResponse(<-result)
			if err != nil {
				return nil, err
			}

			times[i] = lastModified
		}
	}

	return []time.Time{}, nil
}

func handleResponse(r Result) (time.Time, error) {
	if r.Err() != nil {
		return time.Time{}, r.Err()
	}

	defer r.Res().Body.Close()
	if _, err := ioutil.ReadAll(r.Res().Body); err != nil {
		return time.Time{}, err
	}

	if r.Res().StatusCode == http.StatusNotFound {
		return time.Unix(0, 0), nil
	}

	if r.Res().StatusCode != http.StatusOK && r.Res().StatusCode != http.StatusNotModified {
		return time.Time{}, fmt.Errorf("failed to get last-modified date for %s: %w", r.URL(), ErrRequestFailed)
	}

	return time.Parse(time.RFC1123, r.Res().Header.Get("last-modified"))
}
