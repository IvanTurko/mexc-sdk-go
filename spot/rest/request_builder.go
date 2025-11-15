package rest

import (
	"bytes"
	"net/http"
	"net/url"

	"github.com/IvanTurko/mexc-sdk-go/internal/httpx"
	"github.com/IvanTurko/mexc-sdk-go/transport"
)

type requestBuilder struct {
	inner  *httpx.RequestBuilder
	apiKey string
}

func newRequestBuilder(apiKey string) *requestBuilder {
	return &requestBuilder{
		inner:  httpx.NewRequestBuilder(defaultBaseURL),
		apiKey: apiKey,
	}
}

func (s *requestBuilder) WithMethod(method string) *requestBuilder {
	s.inner = s.inner.WithMethod(method)
	return s
}

func (s *requestBuilder) WithPath(path string) *requestBuilder {
	s.inner = s.inner.WithPath(path)
	return s
}

func (s *requestBuilder) WithQuery(query url.Values) *requestBuilder {
	s.inner = s.inner.WithQuery(query)
	return s
}

func (s *requestBuilder) WithBody(body []byte) *requestBuilder {
	s.inner = s.inner.WithBody(bytes.NewReader(body))
	return s
}

func (s *requestBuilder) Build() *transport.Request {
	if s.inner.Method != http.MethodGet {
		headers := buildHeaders(s.apiKey)
		s.inner.WithHeaders(headers)
	}
	return s.inner.Build()
}

func buildHeaders(apiKey string) http.Header {
	h := make(http.Header)
	h.Set("X-MEXC-APIKEY", apiKey)
	h.Set("Content-Type", "application/json")
	return h
}
