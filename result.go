package bulk

import (
	"encoding/json"
	"io/ioutil"
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

// URL returns the originally requested url. If you want to know the final URL, look at the HTTP response.
func (r Result) URL() string {
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

// UnmarshalResponse unmarshals the http response directly into the provided interface
// type (remember to provide a reference, not a value!), and closes the stream afterwards.
//
// If the Result did have an error, or something goes wrong while unmarshalling,
// the error is returned (result error of course taking precedence).
func (r Result) UnmarshalResponse(target interface{}) error {
	if r.Err() != nil {
		return r.Err()
	}

	defer r.Res().Body.Close()
	body, err := ioutil.ReadAll(r.Res().Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, target)
}
