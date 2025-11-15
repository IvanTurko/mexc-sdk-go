package httpx

import (
	"io"
	"net/http"
	"net/url"

	"github.com/IvanTurko/mexc-sdk-go/transport"
)

type RequestBuilder struct {
	BaseURL string
	Path    string
	Method  string
	Params  url.Values
	Headers http.Header
	Body    io.Reader
}

func NewRequestBuilder(baseURL string) *RequestBuilder {
	return &RequestBuilder{
		BaseURL: baseURL,
		Params:  make(url.Values),
		Headers: make(http.Header),
	}
}

func (b *RequestBuilder) WithPath(path string) *RequestBuilder {
	b.Path = path
	return b
}

func (b *RequestBuilder) WithMethod(method string) *RequestBuilder {
	b.Method = method
	return b
}

func (b *RequestBuilder) WithQuery(params url.Values) *RequestBuilder {
	b.Params = params
	return b
}

func (b *RequestBuilder) WithHeaders(headers http.Header) *RequestBuilder {
	b.Headers = headers
	return b
}

func (b *RequestBuilder) WithBody(body io.Reader) *RequestBuilder {
	b.Body = body
	return b
}

func (b *RequestBuilder) Build() *transport.Request {
	fullURL := b.BaseURL + b.Path
	if len(b.Params) > 0 {
		fullURL += "?" + b.Params.Encode()
	}
	return &transport.Request{
		Method:  b.Method,
		FullURL: fullURL,
		Headers: b.Headers,
		Body:    b.Body,
	}
}
