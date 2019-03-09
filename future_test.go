package bulk

import (
	"context"
	"testing"
	"time"
)

// Tests that Done returns false, if the result has not been retrieved yet.
func Test_Done_BeforeResult(t *testing.T) {
	if (&Future{}).Done() {
		t.Error("expectation failed, future is already done")
	}
}

// Tests that Get returns exactly the object that was put into the channel.
func Test_Get(t *testing.T) {
	// given
	resultChan := make(chan Result, 1)
	future := Future{resultChan: resultChan}

	// when
	insertResult := Result{url: "test"}
	resultChan <- insertResult
	result := future.Get()

	// then
	if result != insertResult {
		t.Error("result received, but was somehow mangled")
	}
}

// Tests that GetWithContext with a simple context behaves just like Get.
func Test_Get_WithContext(t *testing.T) {
	// given
	resultChan := make(chan Result, 1)
	future := Future{resultChan: resultChan}

	// when
	insertResult := Result{url: "test"}
	resultChan <- insertResult
	result, err := future.GetWithContext(context.Background())

	// then
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if result != insertResult {
		t.Error("result received, but was somehow mangled")
	}
}

// Tests that GetWithContext returns the contexts error, if the context
// errors (deadline exceeded).
func Test_Get_WithContext_DeadlineExceeded(t *testing.T) {
	// given
	resultChan := make(chan Result, 1)
	future := Future{resultChan: resultChan}

	// this context is immediately exceeded
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
	defer cancel()

	// when
	result, err := future.GetWithContext(ctx)

	// then
	if err != context.DeadlineExceeded {
		t.Errorf("unexpected error: %s", err)
	}

	emptyResult := Result{}
	if result != emptyResult {
		t.Error("non-empty result received")
	}

	if future.Done() {
		t.Error("expectation failed, future is done for error case")
	}
}

// Tests that multiple calls to Get return the same object.
func Test_Get_Multiple_Times(t *testing.T) {
	// given
	resultChan := make(chan Result, 1)
	future := Future{resultChan: resultChan}

	// when
	insertResult := Result{url: "test"}
	resultChan <- insertResult

	result1 := future.Get()
	result2 := future.Get()

	// then
	if result1 != result2 {
		t.Error("subsequent calls to get returned different objects")
	}
}

// Tests that Done returns true, after the result was retrieved via Get.
func Test_Get_Done(t *testing.T) {
	// given
	resultChan := make(chan Result, 1)
	future := Future{resultChan: resultChan}

	// when
	resultChan <- Result{url: "test"}
	future.Get()

	// then
	if !future.Done() {
		t.Error("result received, but future was not set to done")
	}
}

// Tests that Done returns true, after the result was retrieved via GetWithContext.
func Test_Get_WithContext_Done(t *testing.T) {
	// given
	resultChan := make(chan Result, 1)
	future := Future{resultChan: resultChan}

	// when
	resultChan <- Result{url: "test"}
	_, err := future.GetWithContext(context.Background())

	// then
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if !future.Done() {
		t.Error("result received, but future was not set to done")
	}
}
