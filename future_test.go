package bulk

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
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

// Tests that UnmarshalResponse correctly unmarshalls a given response.
func Test_UnmarshalResponse(t *testing.T) {
	// given
	resultChan := make(chan Result, 1)
	future := Future{resultChan: resultChan}

	referenceObj := sampleObject{SomeInt: 4, SomeString: "test"}
	referenceBytes, err := json.Marshal(referenceObj)
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}

	closeRecorder := &closeRecorder{ReadCloser: ioutil.NopCloser(bytes.NewReader(referenceBytes))}
	response := &http.Response{Body: closeRecorder}

	// when
	resultChan <- Result{res: response}

	var responseObj sampleObject
	if err := future.UnmarshalResponse(&responseObj); err != nil {
		t.Errorf("unexpected error %s", err)
	}

	// then
	if responseObj != referenceObj {
		t.Error("result received, but was not like reference object")
	}

	if !closeRecorder.isClosed {
		t.Error("http stream was not closed")
	}
}

// Tests that UnmarshalResponse correctly unmarshalls subsequent calls.
func Test_UnmarshalResponse_SubsequentCall(t *testing.T) {
	// given
	resultChan := make(chan Result, 1)
	future := Future{resultChan: resultChan}

	referenceObj := sampleObject{SomeInt: 4, SomeString: "test"}
	referenceBytes, err := json.Marshal(referenceObj)
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}

	closeRecorder := &closeRecorder{ReadCloser: ioutil.NopCloser(bytes.NewReader(referenceBytes))}
	response := &http.Response{Body: closeRecorder}

	// when
	resultChan <- Result{res: response}

	var responseObj1, responseObj2 sampleObject
	if err := future.UnmarshalResponse(&responseObj1); err != nil {
		t.Errorf("unexpected error %s", err)
	}
	if err := future.UnmarshalResponse(&responseObj2); err != nil {
		t.Errorf("unexpected error %s", err)
	}

	// then
	if responseObj1 != referenceObj {
		t.Error("first result received, but was not like reference object")
	}

	if responseObj2 != referenceObj {
		t.Error("second result received, but was not like reference object")
	}

	if !closeRecorder.isClosed {
		t.Error("http stream was not closed")
	}
}

// Tests that UnmarshalResponse returns the exact error if an error occurred while
// reading the stream, and subsequent calls return the same error.
func Test_UnmarshalResponse_ReadError(t *testing.T) {
	// given
	resultChan := make(chan Result, 1)
	future := Future{resultChan: resultChan}

	closeRecorder := &errorCloser{closeRecorder: closeRecorder{ReadCloser: ioutil.NopCloser(bytes.NewReader(nil))}}
	response := &http.Response{Body: closeRecorder}

	// when
	resultChan <- Result{res: response}

	firstErr := future.UnmarshalResponse(&sampleObject{})
	if firstErr == nil {
		t.Error("expected error, but none occurred")
	}

	secondErr := future.UnmarshalResponse(&sampleObject{})

	// then
	if !closeRecorder.isClosed {
		t.Error("http stream was not closed")
	}

	if firstErr != secondErr {
		t.Error("subsequent calls did not return the same error")
	}
}

// Tests that UnmarshalResponse returns the exact error of the result, if existing.
func Test_UnmarshalResponse_ResultError(t *testing.T) {
	// given
	resultChan := make(chan Result, 1)
	future := Future{resultChan: resultChan}

	referenceErr := errors.New("expected error")

	// when
	resultChan <- Result{err: referenceErr}
	err := future.UnmarshalResponse(&sampleObject{})

	// then
	if err != referenceErr {
		t.Errorf("expected result error, received unexpected error %s", err)
	}
}
