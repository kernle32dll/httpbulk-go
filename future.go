package bulk

import "sync"

// Future is a holder for a bulk.Response channel, which allows to
// safely "read" the result multiple times
type Future struct {
	resultChan chan Result
	done       bool
	result     Result

	mutex sync.Mutex
}

// Done allows to introspect if the channel has already been read.
func (future *Future) Done() bool {
	return future.done
}

// Get retrieves the underlying result, and caches it.
// Subsequent calls to get will not query the channel, but
// return the cached result.
func (future *Future) Get() Result {
	future.mutex.Lock()
	defer future.mutex.Unlock()

	if !future.done {
		future.result = <-future.resultChan
		future.done = true
	}

	return future.result
}
