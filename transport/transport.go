package transport

import (
	"context"
	"io"
	"net/http"
)

// HTTPClient defines the minimal interface required by the SDK to execute
// HTTP requests. Users may inject any custom implementation (e.g., mocks or wrappers).
type HTTPClient interface {
	// Do executes an HTTP request. The implementation must respect the context.
	Do(ctx context.Context, req *Request) (*Response, error)
}

// Request is the SDK's lightweight representation of an HTTP request.
type Request struct {
	Method  string
	FullURL string
	Headers http.Header
	Body    io.Reader
}

// Response is the fully-buffered result of an HTTP request.
type Response struct {
	Body       []byte
	StatusCode int
	Headers    http.Header
}

type internalHTTPClientAdapter struct {
	client *http.Client
}

// NewHTTPClient wraps a standard *http.Client into an HTTPClient.
// If nil is provided, a default http.Client is used.
func NewHTTPClient(stdClient *http.Client) HTTPClient {
	if stdClient == nil {
		stdClient = &http.Client{}
	}
	return &internalHTTPClientAdapter{client: stdClient}
}

// Do executes the request using the underlying standard http.Client.
func (a *internalHTTPClientAdapter) Do(ctx context.Context, req *Request) (*Response, error) {
	stdReq, err := http.NewRequestWithContext(ctx, req.Method, req.FullURL, req.Body)
	if err != nil {
		return nil, err
	}
	stdReq.Header = req.Headers

	stdResp, err := a.client.Do(stdReq)
	if err != nil {
		return nil, err
	}
	defer stdResp.Body.Close()

	bodyBytes, err := io.ReadAll(stdResp.Body)
	if err != nil {
		return nil, err
	}

	return &Response{
		Body:       bodyBytes,
		StatusCode: stdResp.StatusCode,
		Headers:    stdResp.Header,
	}, nil
}
