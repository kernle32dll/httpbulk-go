package bulk

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// FetchLastModDatesForUrls fetches the last modification date for multiple urls at once.
func FetchLastModDatesForUrls(executor Executor, modifyRequest func(r *http.Request) error, urls ...string) ([]time.Time, error) {
	if urls != nil && len(urls) > 0 {
		initialContext, cancel := context.WithCancel(context.Background())
		results := executor.AddRequestsWithInterceptor(initialContext, modifyRequest, urls...)

		times := make([]time.Time, len(urls))
		for i, result := range results {
			lastModified, err := handleResponse(<-result)
			if err != nil {
				cancel()
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

	if r.Res().StatusCode == 404 {
		return time.Unix(0, 0), nil
	}

	if r.Res().StatusCode != 200 && r.Res().StatusCode != 304 {
		return time.Time{}, fmt.Errorf("failed to get last-modified date for %s", r.Url())
	}

	return time.Parse(time.RFC1123, r.Res().Header.Get("last-modified"))
}
