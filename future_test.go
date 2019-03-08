package bulk

import (
	"testing"
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
