package errors

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/radanalyticsio/oshinko-cli/rest/models"
)

// errorResponseWriter is a wrapper struct to help with inspecting the
// status code and reponse produced by a request.
type errorResponseWriter struct {
	writer   http.ResponseWriter
	response *[]byte
	status   int
}

func (w *errorResponseWriter) Header() http.Header {
	return w.writer.Header()
}

func (w *errorResponseWriter) Write(b []byte) (int, error) {
	// NOTE(elmiko) this function needs to save the original response bytes
	// so that they can be inspected by the error handler middleware. we
	// save the original and return the length of the bytes so that the next
	// caller will continue to operate normally.
	w.response = &b
	return len(b), nil
}

func (w *errorResponseWriter) WriteHeader(s int) {
	w.status = s
	w.writer.WriteHeader(s)
}

func marshalErrorResponse(m *models.ErrorResponse) (resp *[]byte) {
	if r, err := json.Marshal(m); err != nil {
		r = []byte("Unmarshal error")
		resp = &r
	} else {
		resp = &r
	}
	return
}

func statusCode(s int) (text string) {
	switch s {
	case 404:
		text = "Not Found"
	case 422:
		text = "Unprocessable Entity"
	case 500:
		text = "Internal Server Error"
	default:
		text = fmt.Sprintf("Unrecognized Error (%d)", s)
	}
	return
}

type errorSchema struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func isWrongSchemaMessage(b *[]byte) (isWrong bool, msg string) {
	var es errorSchema
	if err := json.Unmarshal(*b, &es); err == nil {
		if es.Code != 0 && es.Message != "" {
			msg = es.Message
			isWrong = true
		}
	}
	return
}

// AddErrorHandler will decorate the passed handler with a wrapper which will
// transform error output from the go-swagger schema into the error format
// schema defined by the oshinko api definition.
func AddErrorHandler(next http.Handler) http.Handler {
	// NOTE(elmiko) This function is doing a bunch of stuff to determine if
	// the response from an error is formatted according to the schema
	// defined by the oshinko-rest api.
	// For all normal messages, we want to process them and return the result.
	// But for errors, we want to determine if the response is using the error
	// schema defined by the go-swagger tooling. If the response is using the
	// go-swagger schema, we want to pull out the message and use that in
	// the reformatted error response.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		er := errorResponseWriter{w, nil, 0}
		next.ServeHTTP(&er, r)
		resp := er.response
		// Because the underlying layers produce errors for some things
		// that are not checked by our handlers, there are a few specific
		// codes we are interested in reformatting.
		if er.status > 399 && er.status < 600 {
			wrong, mesg := isWrongSchemaMessage(resp)
			if wrong {
				resp = marshalErrorResponse(
					NewSingleErrorResponse(int32(er.status), statusCode(er.status), mesg))
			}
		}
		if resp != nil {
			w.Write(*resp)
		}
	})
}
