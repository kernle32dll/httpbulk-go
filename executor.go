package bulk

import (
	"context"
	"golang.org/x/net/context/ctxhttp"
	"net/http"
	"time"
)

/**
This code has been adapted from the following code gist:
https://gist.github.com/montanaflynn/ea4b92ed640f790c4b9cee36046a5383

All kudos to Montana Flynn (montanaflynn)
*/

// Executor is the central bulk request maintainer.
type Executor struct {
	client        *http.Client
	semaphoreChan chan struct{}
}

// Close closes the internal channels, and makes the Executor unavailable for further usage.
func (e Executor) Close() {
	if e.semaphoreChan != nil {
		close(e.semaphoreChan)
	}
}

// NewExecutor instantiates a new Executor.
func NewExecutor(setters ...Option) Executor {
	// Default Options
	args := &Options{
		ConcurrencyLimit: 10,
		Client:           http.DefaultClient,
	}

	for _, setter := range setters {
		setter(args)
	}

	// this buffered channel will block at the concurrency limit
	var semaphoreChan chan struct{}
	if args.ConcurrencyLimit > 0 {
		semaphoreChan = make(chan struct{}, args.ConcurrencyLimit)
	}

	return Executor{
		client:        args.Client,
		semaphoreChan: semaphoreChan,
	}
}

// Issues one or more urls to be called.
// For each call, optional hooks for modifying the request are executed (if not nil).
func (e Executor) AddRequestsWithInterceptor(
	ctx context.Context,
	modifyRequest func(r *http.Request) error,
	urls ...string,
) []chan Result {
	results := make([]chan Result, len(urls))
	for i, url := range urls {
		results[i] = e.addRequestInternal(ctx, modifyRequest, url)
	}

	return results
}

// Issues one or more urls to be called.
func (e Executor) AddRequests(
	ctx context.Context,
	urls ...string,
) []chan Result {
	return e.AddRequestsWithInterceptor(ctx, nil, urls...)
}

// Issues one or more urls to be called and wrapped in a bulk.Future.
// For each call, optional hooks for modifying the request are executed (if not nil).
func (e Executor) AddFutureRequestsWithInterceptor(
	ctx context.Context,
	modifyRequest func(r *http.Request) error,
	urls ...string,
) []*Future {
	results := make([]*Future, len(urls))
	for i, url := range urls {
		results[i] = &Future{resultChan: e.addRequestInternal(ctx, modifyRequest, url)}
	}

	return results
}

// Issues one or more urls to be called and wrapped in a bulk.Future.
func (e Executor) AddFutureRequests(
	ctx context.Context,
	urls ...string,
) []*Future {
	return e.AddFutureRequestsWithInterceptor(ctx, nil, urls...)
}

func (e Executor) addRequestInternal(
	ctx context.Context,
	modifyRequest func(r *http.Request) error,
	url string,
) chan Result {
	resultChannel := make(chan Result, 1)

	// start a go routine with the index and url in a closure
	go func(url string, ctx context.Context) {
		// If concurrency limit enabled...
		if e.semaphoreChan != nil {
			// this sends an empty struct into the semaphoreChan which
			// is basically saying add one to the limit, but when the
			// limit has been reached block until there is room
			e.semaphoreChan <- struct{}{}
		}

		var result Result
		if req, err := http.NewRequest("GET", url, nil); err == nil {
			doSend := true
			if modifyRequest != nil {
				if err := modifyRequest(req); err != nil {
					result = Result{url, nil, 0, err}
					doSend = false
				}
			}

			if doSend {
				start := time.Now()

				// send the request and put the response in a result struct
				// along with the index so we can sort them later along with
				// any error that might have occurred
				res, err := ctxhttp.Do(ctx, e.client, req)
				result = Result{url, res, time.Since(start), err}
			}
		} else {
			result = Result{url, nil, 0, err}
		}

		// now we can send the result struct through the results channel
		resultChannel <- result

		// If concurrency limit enabled...
		if e.semaphoreChan != nil {
			// once we're done it's we read from the semaphoreChan which
			// has the effect of removing one from the limit and allowing
			// another goroutine to start
			<-e.semaphoreChan
		}
	}(url, ctx)

	return resultChannel
}
