package bulk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

type sampleObject struct {
	SomeInt    int    `json:"someInt"`
	SomeString string `json:"someString"`
}

type closeRecorder struct {
	io.ReadCloser
	isClosed bool
}

func (closeRecorder *closeRecorder) Close() error {
	closeRecorder.isClosed = true
	return closeRecorder.ReadCloser.Close()
}

type errorCloser struct {
	closeRecorder
}

func (errorRecorder *errorCloser) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("expected failure at %d", time.Now().Unix())
}

// Tests that the getters return the correct objects.
func Test_Result_Getters(t *testing.T) {
	// given
	testError := errors.New("expected error")
	testResponse := &http.Response{
		StatusCode: http.StatusTeapot,
	}

	result := Result{
		url: "test-url",
		res: testResponse,
		dur: time.Hour,
		err: testError,
	}

	if result.URL() != "test-url" {
		t.Error("getter for URL broken")
	}

	if result.Res().StatusCode != testResponse.StatusCode {
		t.Error("getter for http response broken")
	}

	if result.Duration() != time.Hour {
		t.Error("getter for duration broken")
	}

	if !errors.Is(result.Err(), testError) {
		t.Error("getter for error broken")
	}
}

// Tests that UnmarshalResponse correctly unmarshalls a given response.
func Test_Result_UnmarshalResponse(t *testing.T) {
	// given
	referenceObj := sampleObject{SomeInt: 4, SomeString: "test"}
	referenceBytes, err := json.Marshal(referenceObj)
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}

	closeRecorder := &closeRecorder{ReadCloser: ioutil.NopCloser(bytes.NewReader(referenceBytes))}
	response := &http.Response{Body: closeRecorder}
	result := Result{res: response}

	// when
	var responseObj sampleObject
	if err := result.UnmarshalResponse(&responseObj); err != nil {
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

// Tests that UnmarshalResponse returns the exact error if an error occurred while
// reading the stream.
func Test_Result_UnmarshalResponse_ReadError(t *testing.T) {
	// given
	closeRecorder := &errorCloser{closeRecorder: closeRecorder{ReadCloser: ioutil.NopCloser(bytes.NewReader(nil))}}
	response := &http.Response{Body: closeRecorder}
	result := Result{res: response}

	// when
	err := result.UnmarshalResponse(&sampleObject{})

	// then
	if err == nil {
		t.Error("expected error, but none occurred")
	}

	if !closeRecorder.isClosed {
		t.Error("http stream was not closed")
	}
}

// Tests that UnmarshalResponse returns the exact error of the result, if existing.
func Test_Result_UnmarshalResponse_ResultError(t *testing.T) {
	// given
	referenceErr := errors.New("expected error")
	result := Result{err: referenceErr}

	// when
	err := result.UnmarshalResponse(&sampleObject{})

	// then
	if !errors.Is(err, referenceErr) {
		t.Errorf("expected result error, received unexpected error %s", err)
	}
}
