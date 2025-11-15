package keyservice

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"testing"

	"github.com/IvanTurko/mexc-sdk-go/internal/testutil"
	"github.com/IvanTurko/mexc-sdk-go/sdkerr"
	"github.com/IvanTurko/mexc-sdk-go/transport"
	"github.com/stretchr/testify/assert"
)

func TestGetListenKeysService_buildQuery(t *testing.T) {
	mockTime := func() int64 { return 1620000000000 }
	secretKey := "my-secret-key"

	recvWindow := int64(5000)

	svc := NewGetListenKeysService("api-key", secretKey).
		RecvWindow(recvWindow)

	svc.timestamp = mockTime

	q := svc.buildQuery()

	assert.Equal(t, strconv.FormatInt(mockTime(), 10), q.Get("timestamp"))
	assert.NotEmpty(t, q.Get("signature"))
}

func TestGetListenKeysService_Do_Success(t *testing.T) {
	expectedRecvWindow := int64(5000)
	expectedListenKey := []string{
		"c285bc363cfeac6646576b801a2ed1f9523310fcda9e927e509aaaaaaaaaaaaaa",
		"87cb8da0fb150e36c232c2c060bc3848693312008caf3acae73bbbbbbbbbbbb",
	}

	fakeClient := &testutil.FakeHTTPClient{
		DoFunc: func(ctx context.Context, req *transport.Request) (*transport.Response, error) {
			assert.Equal(t, http.MethodGet, req.Method)
			assert.Contains(t, req.FullURL, "/api/v3/userDataStream")

			assert.NotEmpty(t, req.Headers.Get("X-MEXC-APIKEY"))
			assert.Equal(t, "application/json", req.Headers.Get("Content-Type"))

			query := testutil.ExtractQuery(t, req.FullURL)
			assert.Equal(t, strconv.FormatInt(expectedRecvWindow, 10), query.Get("recvWindow"))
			assert.NotEmpty(t, query.Get("timestamp"))

			return &transport.Response{
				StatusCode: 200,
				Body: []byte(`{
					"listenKey": [
					    "c285bc363cfeac6646576b801a2ed1f9523310fcda9e927e509aaaaaaaaaaaaaa",
		   				"87cb8da0fb150e36c232c2c060bc3848693312008caf3acae73bbbbbbbbbbbb"
		    		]
				}`),
			}, nil
		},
	}

	svc := NewGetListenKeysService("API_KEY", "SECRET_KEY").
		WithClient(fakeClient).
		RecvWindow(expectedRecvWindow)

	ctx := context.Background()
	result, err := svc.Do(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	assert.Len(t, result.ListenKey, len(expectedListenKey))
	for i, key := range expectedListenKey {
		assert.Equal(t, key, result.ListenKey[i])
	}
}

func TestGetListenKeysService_Do_Errors(t *testing.T) {
	type testCase struct {
		name     string
		setup    func() transport.HTTPClient
		wantKind error
	}

	cases := []testCase{
		{
			name: "client fails",
			setup: func() transport.HTTPClient {
				return &testutil.FakeHTTPClient{
					DoFunc: func(ctx context.Context, req *transport.Request) (*transport.Response, error) {
						return nil, errors.New("network is down")
					},
				}
			},
			wantKind: sdkerr.ErrRequestFailed,
		},
		{
			name: "bad status",
			setup: func() transport.HTTPClient {
				return &testutil.FakeHTTPClient{
					DoFunc: func(ctx context.Context, req *transport.Request) (*transport.Response, error) {
						return &transport.Response{
							StatusCode: 500,
							Body:       []byte(`internal error`),
						}, nil
					},
				}
			},
			wantKind: sdkerr.ErrAPIError,
		},
		{
			name: "decode fails",
			setup: func() transport.HTTPClient {
				return &testutil.FakeHTTPClient{
					DoFunc: func(ctx context.Context, req *transport.Request) (*transport.Response, error) {
						return &transport.Response{
							StatusCode: 200,
							Body:       []byte(`{invalid json}`),
						}, nil
					},
				}
			},
			wantKind: sdkerr.ErrDecodeError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := tc.setup()

			svc := NewGetListenKeysService("", "").WithClient(client)
			ctx := context.Background()
			result, err := svc.Do(ctx)

			assert.Nil(t, result)
			assert.Error(t, err)

			var sdkErr *sdkerr.SDKError
			assert.ErrorAs(t, err, &sdkErr)
			assert.Equal(t, tc.wantKind, sdkErr.Kind())
		})
	}
}
