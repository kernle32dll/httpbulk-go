package bulk

import (
	"encoding/json"
	"io/ioutil"
	"sync"
)

// Future is a holder for a bulk.Response channel, which allows to
// safely "read" the result multiple times
type Future struct {
	resultChan chan Result
	done       bool
	result     Result

	mutex sync.Mutex

	readMutex sync.Mutex
	readDone  bool
	readBytes []byte
	readErr   error
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

// UnmarshalResponse is the concurrency and multi-read safe version
// of the Result UnmarshalResponse function.
//
// First, the result is retrieved and cached as described in the Get() method.
// Afterwards, (if not done already) the body of the response stream is read (and afterwards
// closed!), and also cached. Finally, this cached body is then used for all unmarshalling attempts.
//
// This effectively means, that the same response can be marshalled into different
// target types via this function.
//
// If the Result did have an error, or something goes wrong while unmarshalling,
// the error is returned (result error of course taking precedence). If the error originated
// from reading the body, all subsequent calls return the same error, and no attempt to
// read the stream will be done again.
func (future *Future) UnmarshalResponse(target interface{}) error {
	future.readMutex.Lock()
	defer future.readMutex.Unlock()

	result := future.Get()

	// error of the response
	if result.Err() != nil {
		return result.Err()
	}

	// (previous) error while reading the response
	if future.readErr != nil {
		return future.readErr
	}

	if !future.readDone {
		defer result.Res().Body.Close()
		body, err := ioutil.ReadAll(result.Res().Body)
		if err != nil {
			future.readErr = err
			return err
		}

		future.readBytes = body
		future.readDone = true
	}

	return json.Unmarshal(future.readBytes, target)
}
