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

func TestCloseListenKeyService_validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		svc := NewCloseListenKeyService("", "").
			ListenKey("fake").
			RecvWindow(50)

		err := svc.validate()
		assert.NoError(t, err)
	})

	t.Run("missing listenKey", func(t *testing.T) {
		svc := NewCloseListenKeyService("", "")

		err := svc.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "listenKey is required")
	})
}

func TestCloseListenKeyService_buildQuery(t *testing.T) {
	mockTime := func() int64 { return 1620000000000 }
	secretKey := "my-secret-key"

	recvWindow := int64(5000)

	svc := NewCloseListenKeyService("api-key", secretKey).
		ListenKey("fake").
		RecvWindow(recvWindow)

	svc.timestamp = mockTime

	q := svc.buildQuery()

	assert.Equal(t, "fake", q.Get("listenKey"))
	assert.Equal(t, strconv.FormatInt(mockTime(), 10), q.Get("timestamp"))
	assert.NotEmpty(t, q.Get("signature"))
}

func TestNewCloseListenKeyService_Validate_ErrorsWrapped(t *testing.T) {
	svc := NewCloseListenKeyService("", "")

	err := svc.Validate()
	assert.Error(t, err)

	sdkErr, ok := err.(*sdkerr.SDKError)
	assert.True(t, ok)
	assert.Equal(t, sdkerr.ErrValidation, sdkErr.Kind())
	assert.Equal(t, "CloseListenKeyService.Validate", sdkErr.Op())
	assert.Contains(t, sdkErr.Message(), "listenKey is required")
}

func TestCloseListenKeyService_Do_Success(t *testing.T) {
	expectedRecvWindow := int64(5000)
	expectedListenKey := "pqia91ma19a5s61cv6a81va65sdf19v8a65a1a5s61cv6a81va65sdf19v8a65a1"

	fakeClient := &testutil.FakeHTTPClient{
		DoFunc: func(ctx context.Context, req *transport.Request) (*transport.Response, error) {
			assert.Equal(t, http.MethodDelete, req.Method)
			assert.Contains(t, req.FullURL, "/api/v3/userDataStream")

			assert.NotEmpty(t, req.Headers.Get("X-MEXC-APIKEY"))
			assert.Equal(t, "application/json", req.Headers.Get("Content-Type"))

			query := testutil.ExtractQuery(t, req.FullURL)
			assert.Equal(t, expectedListenKey, query.Get("listenKey"))
			assert.Equal(t, strconv.FormatInt(expectedRecvWindow, 10), query.Get("recvWindow"))
			assert.NotEmpty(t, query.Get("timestamp"))

			return &transport.Response{
				StatusCode: 200,
				Body: []byte(`{
    				"listenKey": "pqia91ma19a5s61cv6a81va65sdf19v8a65a1a5s61cv6a81va65sdf19v8a65a1"
				}`),
			}, nil
		},
	}

	svc := NewCloseListenKeyService("API_KEY", "SECRET_KEY").
		WithClient(fakeClient).
		ListenKey(expectedListenKey).
		RecvWindow(expectedRecvWindow)

	ctx := context.Background()
	actualKey, err := svc.Do(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expectedListenKey, actualKey)
}

func TestCloseListenKeyService_Do_Errors(t *testing.T) {
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

			svc := NewCloseListenKeyService("", "").WithClient(client)
			ctx := context.Background()
			_, err := svc.Do(ctx)

			assert.Error(t, err)

			var sdkErr *sdkerr.SDKError
			assert.ErrorAs(t, err, &sdkErr)
			assert.Equal(t, tc.wantKind, sdkErr.Kind())
		})
	}
}
