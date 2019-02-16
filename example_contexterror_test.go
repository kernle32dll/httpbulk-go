package bulk_test

import (
	"github.com/kernle32dll/httpbulk-go"

	"context"
	"log"
	"testing"
)

// Demonstrates, how errors for a canceled context propagate to the result.
func TestContextError(t *testing.T) {
	executor := bulk.NewExecutor()

	urls := []string{
		"https://www.google.com",
		"https://www.bing.com",
		"https://www.yahoo.com",
		"https://www.tarent.de",
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	cancelFunc() // Immediately cancel the context

	results := executor.AddRequests(ctx, urls...)
	for _, resultChan := range results {
		result := <-resultChan

		log.Printf("%s responded with %s", result.Url(), result.Err())
	}
}
