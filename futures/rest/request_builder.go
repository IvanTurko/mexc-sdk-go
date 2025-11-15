package rest

import (
	"bytes"
	"net/http"
	"net/url"
	"strconv"

	"github.com/IvanTurko/mexc-sdk-go/internal/httpx"
	"github.com/IvanTurko/mexc-sdk-go/internal/signature"
	"github.com/IvanTurko/mexc-sdk-go/internal/timeutil"
	"github.com/IvanTurko/mexc-sdk-go/transport"
)

type requestBuilder struct {
	inner     *httpx.RequestBuilder
	apiKey    string
	secretKey string
	timestamp func() int64
	body      []byte
}

func newRequestBuilder(apiKey, secretKey string) *requestBuilder {
	return &requestBuilder{
		inner:     httpx.NewRequestBuilder(defaultBaseURL),
		apiKey:    apiKey,
		secretKey: secretKey,
		timestamp: timeutil.NowMillis,
	}
}

func (b *requestBuilder) WithMethod(method string) *requestBuilder {
	b.inner = b.inner.WithMethod(method)
	return b
}

func (b *requestBuilder) WithPath(path string) *requestBuilder {
	b.inner = b.inner.WithPath(path)
	return b
}

func (b *requestBuilder) WithQuery(params url.Values) *requestBuilder {
	b.inner = b.inner.WithQuery(params)
	return b
}

func (b *requestBuilder) WithBody(body []byte) *requestBuilder {
	b.body = body
	return b
}

func (b *requestBuilder) Build() *transport.Request {
	if b.apiKey != "" {
		timestampStr := strconv.FormatInt(b.timestamp(), 10)
		var toSign string
		switch b.inner.Method {
		case http.MethodGet, http.MethodDelete:
			toSign = b.inner.Params.Encode()
		default: // POST, PUT
			toSign = string(b.body)
		}

		signPayload := b.apiKey + timestampStr + toSign
		sig := signature.HMACSHA256(signPayload, b.secretKey)

		headers := b.buildHeaders(timestampStr, sig)
		b.inner.WithHeaders(headers)
	}

	if b.body != nil {
		b.inner = b.inner.WithBody(bytes.NewReader(b.body))
	}

	return b.inner.Build()
}

func (b *requestBuilder) buildHeaders(timestamp, sig string) http.Header {
	h := make(http.Header)
	h.Set("ApiKey", b.apiKey)
	h.Set("Request-Time", timestamp)
	h.Set("Signature", sig)
	h.Set("Content-Type", "application/json")
	return h
}
