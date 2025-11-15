package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/IvanTurko/mexc-sdk-go/internal/testutil"
	"github.com/IvanTurko/mexc-sdk-go/sdkerr"
	"github.com/IvanTurko/mexc-sdk-go/transport"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderBookService_validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		svc := NewOrderBookService().Symbol("BTCUSDT").Limit(100)
		err := svc.validate()
		assert.NoError(t, err)
	})

	t.Run("missing symbol", func(t *testing.T) {
		svc := NewOrderBookService().Limit(100)
		err := svc.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "symbol is required")
	})

	t.Run("limit too small", func(t *testing.T) {
		svc := NewOrderBookService().Symbol("BTCUSDT").Limit(0)
		err := svc.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "limit must be between 1 and 5000")
	})

	t.Run("limit too large", func(t *testing.T) {
		svc := NewOrderBookService().Symbol("BTCUSDT").Limit(5001)
		err := svc.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "limit must be between 1 and 5000")
	})

	t.Run("multiple errors", func(t *testing.T) {
		svc := NewOrderBookService().Limit(5001) // no symbol, bad limit
		err := svc.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "symbol is required")
		assert.Contains(t, err.Error(), "limit must be between 1 and 5000")
	})
}

func TestOrderBookService_buildQuery(t *testing.T) {
	t.Run("symbol and limit", func(t *testing.T) {
		svc := NewOrderBookService().Symbol("ETHUSDT").Limit(50)
		q := svc.buildQuery()
		assert.Equal(t, "ETHUSDT", q.Get("symbol"))
		assert.Equal(t, "50", q.Get("limit"))
	})

	t.Run("symbol only", func(t *testing.T) {
		svc := NewOrderBookService().Symbol("BNBUSDT")
		q := svc.buildQuery()
		assert.Equal(t, "BNBUSDT", q.Get("symbol"))
		assert.Empty(t, q.Get("limit"))
	})
}

func TestOrderBookService_Validate_ErrorsWrapped(t *testing.T) {
	svc := NewOrderBookService().Limit(0)

	err := svc.Validate()
	assert.Error(t, err)

	sdkErr, ok := err.(*sdkerr.SDKError)
	assert.True(t, ok)
	assert.Equal(t, sdkerr.ErrValidation, sdkErr.Kind())
	assert.Equal(t, "OrderBookService.Validate", sdkErr.Op())
	assert.Contains(t, sdkErr.Message(), "symbol is required")
	assert.Contains(t, sdkErr.Message(), "limit must be between 1 and 5000")
}

func TestOrderBookService_Do_Success(t *testing.T) {
	expectedSymbol := "BTCUSDT"
	expectedLimit := "5"

	fakeClient := &testutil.FakeHTTPClient{
		DoFunc: func(ctx context.Context, req *transport.Request) (*transport.Response, error) {
			assert.Equal(t, http.MethodGet, req.Method)
			assert.Contains(t, req.FullURL, "/api/v3/depth")

			query := testutil.ExtractQuery(t, req.FullURL)
			assert.Equal(t, expectedSymbol, query.Get("symbol"))
			assert.Equal(t, expectedLimit, query.Get("limit"))

			return &transport.Response{
				StatusCode: 200,
				Body: []byte(`{
					"lastUpdateId": 12345,
					"bids": [["50000.00", "0.1"]],
					"asks": [["51000.00", "0.2"]]
				}`),
			}, nil
		},
	}

	svc := NewOrderBookService().WithClient(fakeClient).Symbol(expectedSymbol).Limit(5)

	ctx := context.Background()
	result, err := svc.Do(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	assert.Equal(t, int64(12345), result.LastUpdateId)
	require.Len(t, result.Bids, 1)
	require.Len(t, result.Asks, 1)

	testutil.AssertDecimalEqual(t, result.Bids[0].Price, "50000.00", "bid price mismatch")
	testutil.AssertDecimalEqual(t, result.Bids[0].Quantity, "0.1", "bid qty mismatch")
	testutil.AssertDecimalEqual(t, result.Asks[0].Price, "51000.00", "ask price mismatch")
	testutil.AssertDecimalEqual(t, result.Asks[0].Quantity, "0.2", "ask qty mismatch")
}

func TestOrderBookService_Do_Errors(t *testing.T) {
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

			svc := NewOrderBookService().WithClient(client).Symbol("BTCUSDT").Limit(5)

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

func TestBookLevel_UnmarshalJSON(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		data := []byte(`["123.45", "67.89"]`)

		var b BookLevel
		err := json.Unmarshal(data, &b)
		assert.NoError(t, err)

		expectedPrice, _ := decimal.NewFromString("123.45")
		expectedQty, _ := decimal.NewFromString("67.89")

		assert.True(t, b.Price.Equal(expectedPrice), "price should match")
		assert.True(t, b.Quantity.Equal(expectedQty), "quantity should match")
	})
	t.Run("invalid length", func(t *testing.T) {
		data := []byte(`["123.45"]`)

		var b BookLevel
		err := json.Unmarshal(data, &b)
		assert.Error(t, err, "should fail if array has less than 2 elements")
	})
}
