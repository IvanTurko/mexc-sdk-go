package httpx

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/IvanTurko/mexc-sdk-go/transport"
	"github.com/stretchr/testify/assert"
)

func Test_DefaultHTTPClient_Do(t *testing.T) {
	expectedBody := []byte(`{"ok": true}`)
	expectedStatus := 200
	expectedHeader := http.Header{"Content-Type": []string{"application/json"}}
	expectedMethod := http.MethodPost
	expectedPath := "/test"
	expectedURL := "https://fake.com" + expectedPath
	expectedCustomHeader := "X-Test"
	expectedCustomHeaderValue := "123"

	client := &fakeHttpDoer{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, expectedMethod, req.Method)
			assert.Equal(t, expectedURL, req.URL.String())
			assert.Equal(t, expectedCustomHeaderValue, req.Header.Get(expectedCustomHeader))

			return &http.Response{
				StatusCode: expectedStatus,
				Header:     expectedHeader,
				Body:       io.NopCloser(bytes.NewReader(expectedBody)),
			}, nil
		},
	}

	executor := DefaultHTTPClient{client: client}

	req := &transport.Request{
		Method:  expectedMethod,
		FullURL: expectedURL,
		Headers: http.Header{expectedCustomHeader: []string{expectedCustomHeaderValue}},
	}

	resp, err := executor.Do(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, expectedBody, resp.Body)
	assert.Equal(t, expectedStatus, resp.StatusCode)
	assert.Equal(t, expectedHeader, resp.Headers)
}

type fakeHttpDoer struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (f *fakeHttpDoer) Do(req *http.Request) (*http.Response, error) {
	return f.DoFunc(req)
}

var _ httpDoer = (*fakeHttpDoer)(nil)
