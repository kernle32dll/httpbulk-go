package bulk

import (
	"net/http"
	"time"
)

// Result is a simple data holder for bulk request results.
type Result struct {
	url string
	res *http.Response
	dur time.Duration
	err error
}

// Url returns the originally requested url. If you want to know the final URL, look at the HTTP response.
func (r Result) Url() string {
	return r.url
}

// Err returns an error, if any occurred.
func (r Result) Err() error {
	return r.err
}

// Res returns the http response, if no error occurred.
func (r Result) Res() http.Response {
	return *r.res
}

// Duration returns the amount of time the request took.
func (r Result) Duration() time.Duration {
	return r.dur
}
