package unittest

import (
	"encoding/json"
	"net/http"

	"gopkg.in/check.v1"

	"github.com/radanalyticsio/oshinko-cli/rest/helpers/errors"
)

func (s *OshinkoUnitTestSuite) TestNewSingleErrorReponse(c *check.C) {
	expectedStatus := int32(123)
	expectedTitle := "A good test title"
	expectedDetails := "These are the details of the test"
	observedResponse := errors.NewSingleErrorResponse(
		expectedStatus, expectedTitle, expectedDetails)
	c.Assert(len(observedResponse.Errors), check.Equals, 1)
	c.Assert(*observedResponse.Errors[0].Status, check.Equals, expectedStatus)
	c.Assert(*observedResponse.Errors[0].Title, check.Equals, expectedTitle)
	c.Assert(*observedResponse.Errors[0].Details, check.Equals, expectedDetails)
}

type testResponseWriter struct {
	header      http.Header
	response    *[]byte
	writeCalled bool
}

func (w *testResponseWriter) Header() http.Header {
	return w.header
}

func (w *testResponseWriter) Write(b []byte) (int, error) {
	w.response = &b
	return len(b), nil
}

func (w *testResponseWriter) WriteHeader(s int) {
}

func (s *OshinkoUnitTestSuite) TestAddErrorHandler(c *check.C) {
	expectedResponse := []byte(`{"code": 1234, "message": "a test message"}`)
	expectedWriteLen := len(expectedResponse)
	var observedWriteLen int
	testWriter := testResponseWriter{}
	testRequest := http.Request{}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		observedWriteLen, _ = w.Write(expectedResponse)
		w.WriteHeader(200)
	})
	wrappedHandler := errors.AddErrorHandler(testHandler)
	wrappedHandler.ServeHTTP(&testWriter, &testRequest)
	c.Assert(*testWriter.response, check.DeepEquals, expectedResponse)
	c.Assert(observedWriteLen, check.Equals, expectedWriteLen)

	expectedResponse, _ = json.Marshal(errors.NewSingleErrorResponse(500, "Internal Server Error", "a test message"))
	testResponse := []byte(`{"code": 1234, "message": "a test message"}`)
	expectedWriteLen = len(testResponse)
	testHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		observedWriteLen, _ = w.Write(testResponse)
		w.WriteHeader(500)
	})
	wrappedHandler = errors.AddErrorHandler(testHandler)
	testWriter = testResponseWriter{}
	wrappedHandler.ServeHTTP(&testWriter, &testRequest)
	c.Assert(*testWriter.response, check.DeepEquals, expectedResponse)
	c.Assert(observedWriteLen, check.Equals, expectedWriteLen)

	expectedWriteLen = len(expectedResponse)
	testHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		observedWriteLen, _ = w.Write(expectedResponse)
		w.WriteHeader(500)
	})
	wrappedHandler = errors.AddErrorHandler(testHandler)
	testWriter = testResponseWriter{}
	wrappedHandler.ServeHTTP(&testWriter, &testRequest)
	c.Assert(*testWriter.response, check.DeepEquals, expectedResponse)
	c.Assert(observedWriteLen, check.Equals, expectedWriteLen)

	testHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})
	wrappedHandler = errors.AddErrorHandler(testHandler)
	testWriter = testResponseWriter{}
	wrappedHandler.ServeHTTP(&testWriter, &testRequest)
	c.Assert(testWriter.writeCalled, check.Equals, false)
}
