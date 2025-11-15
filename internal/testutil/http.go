package testutil

import (
	"context"
	"net/url"
	"testing"

	"github.com/IvanTurko/mexc-sdk-go/transport"
	"github.com/stretchr/testify/assert"
)

func ExtractQuery(t *testing.T, fullURL string) url.Values {
	t.Helper()
	parsed, err := url.Parse(fullURL)
	assert.NoError(t, err)
	return parsed.Query()
}

type FakeHTTPClient struct {
	DoFunc func(ctx context.Context, req *transport.Request) (*transport.Response, error)
}

func (f *FakeHTTPClient) Do(ctx context.Context, req *transport.Request) (*transport.Response, error) {
	return f.DoFunc(ctx, req)
}
