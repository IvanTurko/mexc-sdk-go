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
		svc := NewOrderBookService().Symbol("BTC_USDT").Limit(100)
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
		svc := NewOrderBookService().Symbol("BTC_USDT").Limit(0)
		err := svc.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "limit must be above 0")
	})

	t.Run("multiple errors", func(t *testing.T) {
		svc := NewOrderBookService().Limit(0) // no symbol, bad limit
		err := svc.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "symbol is required")
		assert.Contains(t, err.Error(), "limit must be above 0")
	})
}

func TestOrderBookService_buildQuery(t *testing.T) {
	t.Run("with limit", func(t *testing.T) {
		svc := NewOrderBookService().Symbol("BTC_USDT").Limit(50)
		q := svc.buildQuery()
		assert.Equal(t, "50", q.Get("limit"))
	})

	t.Run("without limit", func(t *testing.T) {
		svc := NewOrderBookService().Symbol("BTC_USDT")
		q := svc.buildQuery()
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
	assert.Contains(t, sdkErr.Message(), "limit must be above 0")
}

func TestOrderBookService_Do_Success(t *testing.T) {
	expectedSymbol := "BTC_USDT"
	expectedLimit := "5"

	fakeClient := &testutil.FakeHTTPClient{
		DoFunc: func(ctx context.Context, req *transport.Request) (*transport.Response, error) {
			assert.Equal(t, http.MethodGet, req.Method)
			assert.Contains(t, req.FullURL, "/api/v1/contract/depth/"+expectedSymbol)

			query := testutil.ExtractQuery(t, req.FullURL)
			assert.Equal(t, expectedLimit, query.Get("limit"))

			return &transport.Response{
				StatusCode: 200,
				Body: []byte(`{
					"data": {
						"asks": [["30971.7", "12238", "1"]],
						"bids": [["30971.6", "2594", "1"]],
						"version": 2622334331,
						"timestamp": 1658996910762
					}
				}`),
			}, nil
		},
	}

	svc := NewOrderBookService().WithClient(fakeClient).Symbol(expectedSymbol).Limit(5)

	ctx := context.Background()
	result, err := svc.Do(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	assert.Equal(t, int64(2622334331), result.Version)
	assert.Equal(t, int64(1658996910762), result.Timestamp)
	require.Len(t, result.Bids, 1)
	require.Len(t, result.Asks, 1)

	testutil.AssertDecimalEqual(t, result.Bids[0].Price, "30971.6", "bid price mismatch")
	testutil.AssertDecimalEqual(t, result.Bids[0].Quantity, "2594", "bid qty mismatch")
	testutil.AssertDecimalEqual(t, result.Bids[0].Orders, "1", "bid orders mismatch")
	testutil.AssertDecimalEqual(t, result.Asks[0].Price, "30971.7", "ask price mismatch")
	testutil.AssertDecimalEqual(t, result.Asks[0].Quantity, "12238", "ask qty mismatch")
	testutil.AssertDecimalEqual(t, result.Asks[0].Orders, "1", "ask orders mismatch")
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

			svc := NewOrderBookService().WithClient(client).Symbol("BTC_USDT").Limit(5)

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
		data := []byte(`[123.45, 67.89, 3]`)

		var b BookLevel
		err := json.Unmarshal(data, &b)
		assert.NoError(t, err)

		expectedPrice, _ := decimal.NewFromString("123.45")
		expectedQty, _ := decimal.NewFromString("67.89")
		expectedOrders, _ := decimal.NewFromString("3")

		assert.True(t, b.Price.Equal(expectedPrice), "price should match")
		assert.True(t, b.Quantity.Equal(expectedQty), "quantity should match")
		assert.True(t, b.Orders.Equal(expectedOrders), "orders should match")
	})
	t.Run("invalid length", func(t *testing.T) {
		data := []byte(`[123.45, 67.89]`)

		var b BookLevel
		err := json.Unmarshal(data, &b)
		assert.Error(t, err, "should fail if array has less than 3 elements")
	})
}
