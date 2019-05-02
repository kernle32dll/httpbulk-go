package bulk_test

import (
	bulk "github.com/kernle32dll/httpbulk-go"

	"net/http"
	"testing"
)

// Tests that the ConcurrencyLimit option correctly applies.
func Test_Option_ConcurrencyLimit(t *testing.T) {
	// given
	option := bulk.ConcurrencyLimit(9001)
	options := &bulk.Options{ConcurrencyLimit: 42}

	// when
	option(options)

	// then
	if options.ConcurrencyLimit != 9001 {
		t.Errorf("concurrencly limit not correctly applied, got %d", options.ConcurrencyLimit)
	}
}

// Tests that the Client option correctly applies.
func Test_Option_Client(t *testing.T) {
	// given
	option := bulk.Client(http.DefaultClient)
	options := &bulk.Options{Client: nil}

	// when
	option(options)

	// then
	if options.Client != http.DefaultClient {
		t.Error("client not correctly applied")
	}
}
