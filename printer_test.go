package httppayload_test

import (
	"bytes"
	"net/http"
	"testing"

	payload "github.com/canpacis/http-payload"
	"github.com/stretchr/testify/assert"
)

type ResponseWriter struct {
	header http.Header
	buffer *bytes.Buffer
	status int
}

func (w *ResponseWriter) Header() http.Header {
	return w.header
}

func (w *ResponseWriter) Write(p []byte) (int, error) {
	return w.buffer.Write(p)
}

func (w *ResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}

func NewResponseWriter() *ResponseWriter {
	return &ResponseWriter{
		header: http.Header{},
		buffer: new(bytes.Buffer),
		status: http.StatusOK,
	}
}

func TestPipePrinter(t *testing.T) {
	w := NewResponseWriter()

	printer := payload.NewPipePrinter(
		payload.NewJSONPrinter(w),
		payload.NewHeaderPrinter(w),
		payload.NewCookiePrinter(w),
	)

	err := printer.Print(&Params{
		Email:    "test@example.com",
		Language: "en",
		Token:    "access-token",
	})

	assert := assert.New(t)
	assert.NoError(err)
	assert.Equal("en", w.header.Get("Accept-Language"))
	assert.Equal("token=access-token; Path=/", w.header.Get("Set-Cookie"))
	assert.Equal("{\"email\":\"test@example.com\"}\n", w.buffer.String())
}
