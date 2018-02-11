package utils

import (
	"bytes"
	"net/http"
	"net/url"
)

type lambdaResponseWriter struct {
	header         http.Header
	StatusCode     int
	ResponseBuffer bytes.Buffer
}

func NewLambdaResponseWriter() (*lambdaResponseWriter, error) {
	return &lambdaResponseWriter{
		header:     make(map[string][]string),
		StatusCode: 0,
	}, nil
}

func (w *lambdaResponseWriter) Header() http.Header {
	return w.header
}

func (w *lambdaResponseWriter) Write(bytes []byte) (int, error) {
	if w.StatusCode == 0 {
		w.StatusCode = 200
	}
	bytesWritten, err := w.ResponseBuffer.Write(bytes)
	return bytesWritten, err
}

func (w *lambdaResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
}

func QueryStringParams2Values(qsp map[string]string) url.Values {
	if qsp == nil {
		return nil
	}
	values := url.Values{}
	for k, v := range qsp {
		values.Add(k, v)
	}
	return values
}

func Headers2Header(headers map[string]string) http.Header {
	if headers == nil {
		return nil
	}
	values := http.Header{}
	for k, v := range headers {
		values.Add(k, v)
	}
	return values
}

func Header2Headers(header http.Header) map[string]string {
	if header == nil {
		return nil
	}
	values := make(map[string]string)
	for k, v := range header {
		values[k] = v[0]
	}
	return values
}
