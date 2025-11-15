package httpx

import (
	"context"
	"io"
	"net/http"

	"github.com/IvanTurko/mexc-sdk-go/transport"
)

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type DefaultHTTPClient struct {
	client httpDoer
}

func NewDefaultHTTPClient() *DefaultHTTPClient {
	return &DefaultHTTPClient{
		client: http.DefaultClient,
	}
}

func (d *DefaultHTTPClient) Do(ctx context.Context, r *transport.Request) (*transport.Response, error) {
	req, err := http.NewRequestWithContext(ctx, r.Method, r.FullURL, r.Body)
	if err != nil {
		return nil, err
	}

	if r.Headers != nil {
		for k, vs := range r.Headers {
			for _, v := range vs {
				req.Header.Add(k, v)
			}
		}
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &transport.Response{
		Body:       body,
		Headers:    resp.Header,
		StatusCode: resp.StatusCode,
	}, nil
}
