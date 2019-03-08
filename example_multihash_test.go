package bulk_test

import (
	"github.com/kernle32dll/httpbulk-go"

	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"
)

// Demonstrates, how to quickly calculate the hash of multiple resources.
func TestMultiHash(t *testing.T) {
	executor := bulk.NewExecutor()

	urls := []string{
		"https://gobyexample.com",
		"https://reactjs.org",
		"https://www.tarent.de",
	}

	results := executor.AddRequestsWithInterceptor(context.Background(), func(r *http.Request) error {
		// Change the request method to head - we don't need the body
		r.Method = "HEAD"
		return nil
	}, urls...)

	h := sha256.New()
	for _, resultChan := range results {
		result := <-resultChan

		hash := result.Res().Header.Get("etag")
		t.Logf("%s hash %s", result.URL(), hash)

		h.Write([]byte(hash))
	}
	hash := h.Sum(nil)

	t.Logf("Final hash: %s", hex.EncodeToString(hash))
}
