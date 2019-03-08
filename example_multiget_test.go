package bulk_test

import (
	"github.com/kernle32dll/httpbulk-go"

	"context"
	"log"
	"testing"
)

// Demonstrates, how to quickly fetch results in parallel.
func TestMultiGet(t *testing.T) {
	executor := bulk.NewExecutor()

	urls := []string{
		"https://www.google.com",
		"https://www.bing.com",
		"https://www.yahoo.com",
		"https://www.tarent.de",
	}

	results := executor.AddRequests(context.Background(), urls...)

	for _, resultChan := range results {
		result := <-resultChan

		log.Printf("%s responded with %s", result.URL(), result.Res().Status)
	}
}
