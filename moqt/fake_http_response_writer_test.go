package moqt

import (
	"net/http"
)

var _ http.ResponseWriter = (*FakeHTTPResponseWriter)(nil)

// FakeHTTPResponseWriter is a fake implementation of http.ResponseWriter.
type FakeHTTPResponseWriter struct {
	HeaderFunc      func() http.Header
	WriteFunc       func(data []byte) (int, error)
	WriteHeaderFunc func(statusCode int)

	header http.Header
}

func (m *FakeHTTPResponseWriter) Header() http.Header {
	if m.HeaderFunc != nil {
		return m.HeaderFunc()
	}
	if m.header == nil {
		m.header = make(http.Header)
	}
	return m.header
}

func (m *FakeHTTPResponseWriter) Write(data []byte) (int, error) {
	if m.WriteFunc != nil {
		return m.WriteFunc(data)
	}
	return len(data), nil
}

func (m *FakeHTTPResponseWriter) WriteHeader(statusCode int) {
	if m.WriteHeaderFunc != nil {
		m.WriteHeaderFunc(statusCode)
	}
}
