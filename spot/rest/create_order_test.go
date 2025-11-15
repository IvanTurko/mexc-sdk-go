package rest

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"testing"

	"github.com/IvanTurko/mexc-sdk-go/internal/testutil"
	"github.com/IvanTurko/mexc-sdk-go/sdkerr"
	"github.com/IvanTurko/mexc-sdk-go/transport"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCreateOrderService_validate(t *testing.T) {
	validQty := decimal.NewFromFloat(1.23)
	validPrice := decimal.NewFromFloat(100.5)
	validQuoteQty := decimal.NewFromFloat(50.0)
	validRecvWindow := int64(5000)

	t.Run("valid config", func(t *testing.T) {
		svc := NewCreateOrderService("", "").
			Symbol("BTCUSDT").
			Side(OrderSideBuy).
			Type(OrderTypeLimit).
			Quantity(validQty).
			QuoteOrderQty(validQuoteQty).
			Price(validPrice).
			RecvWindow(validRecvWindow)

		err := svc.validate()
		assert.NoError(t, err)
	})

	t.Run("missing symbol", func(t *testing.T) {
		svc := NewCreateOrderService("", "").
			Side(OrderSideBuy).
			Type(OrderTypeLimit).
			Quantity(validQty)

		err := svc.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "symbol is required")
	})

	t.Run("invalid side", func(t *testing.T) {
		svc := NewCreateOrderService("", "").
			Symbol("BTCUSDT").
			Side(OrderSide("INVALID")).
			Type(OrderTypeLimit).
			Quantity(validQty)

		err := svc.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "side is invalid")
	})

	t.Run("invalid type", func(t *testing.T) {
		svc := NewCreateOrderService("", "").
			Symbol("BTCUSDT").
			Side(OrderSideBuy).
			Type(OrderType("INVALID")).
			Quantity(validQty)

		err := svc.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "type is invalid")
	})

	t.Run("quantity <= 0", func(t *testing.T) {
		svc := NewCreateOrderService("", "").
			Symbol("BTCUSDT").
			Side(OrderSideBuy).
			Type(OrderTypeLimit).
			Quantity(decimal.NewFromInt(0))

		err := svc.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "quantity must be greater than zero")
	})

	t.Run("quoteOrderQty <= 0", func(t *testing.T) {
		svc := NewCreateOrderService("", "").
			Symbol("BTCUSDT").
			Side(OrderSideBuy).
			Type(OrderTypeMarket).
			QuoteOrderQty(decimal.NewFromInt(0))

		err := svc.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "quoteOrderQty must be greater than zero")
	})

	t.Run("price <= 0", func(t *testing.T) {
		svc := NewCreateOrderService("", "").
			Symbol("BTCUSDT").
			Side(OrderSideBuy).
			Type(OrderTypeLimit).
			Quantity(validQty).
			Price(decimal.NewFromInt(0))

		err := svc.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "price must be greater than zero")
	})

	t.Run("recvWindow too small", func(t *testing.T) {
		svc := NewCreateOrderService("", "").
			Symbol("BTCUSDT").
			Side(OrderSideBuy).
			Type(OrderTypeLimit).
			Quantity(validQty).
			RecvWindow(0)

		err := svc.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "recvWindow must be between 1 and 60000")
	})

	t.Run("recvWindow too large", func(t *testing.T) {
		svc := NewCreateOrderService("", "").
			Symbol("BTCUSDT").
			Side(OrderSideBuy).
			Type(OrderTypeLimit).
			Quantity(validQty).
			RecvWindow(60001)

		err := svc.validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "recvWindow must be between 1 and 60000")
	})

	t.Run("multiple errors", func(t *testing.T) {
		svc := NewCreateOrderService("", "").
			Side(OrderSide("INVALID")).
			Type(OrderType("INVALID")).
			Quantity(decimal.NewFromInt(0)).
			QuoteOrderQty(decimal.NewFromInt(-5)).
			Price(decimal.NewFromInt(0)).
			RecvWindow(999999)

		err := svc.validate()
		assert.Error(t, err)

		assert.Contains(t, err.Error(), "symbol is required")
		assert.Contains(t, err.Error(), "side is invalid")
		assert.Contains(t, err.Error(), "type is invalid")
		assert.Contains(t, err.Error(), "quantity must be greater than zero")
		assert.Contains(t, err.Error(), "quoteOrderQty must be greater than zero")
		assert.Contains(t, err.Error(), "price must be greater than zero")
		assert.Contains(t, err.Error(), "recvWindow must be between 1 and 60000")
	})
}

func TestCreateOrderService_buildQuery(t *testing.T) {
	mockTime := func() int64 { return 1620000000000 }
	secretKey := "my-secret-key"

	t.Run("all fields set", func(t *testing.T) {
		quantity := decimal.NewFromFloat(0.5)
		quoteQty := decimal.NewFromFloat(100.0)
		price := decimal.NewFromFloat(2000.0)
		clientOrderID := "custom-order-id"
		recvWindow := int64(5000)

		svc := NewCreateOrderService("api-key", secretKey).
			Symbol("ETHUSDT").
			Side(OrderSideBuy).
			Type(OrderTypeLimit).
			Quantity(quantity).
			QuoteOrderQty(quoteQty).
			Price(price).
			NewClientOrderId(clientOrderID).
			RecvWindow(recvWindow)

		svc.timestamp = mockTime

		q := svc.buildQuery()

		assert.Equal(t, "ETHUSDT", q.Get("symbol"))
		assert.Equal(t, "BUY", q.Get("side"))
		assert.Equal(t, "LIMIT", q.Get("type"))
		assert.Equal(t, quantity.String(), q.Get("quantity"))
		assert.Equal(t, quoteQty.String(), q.Get("quoteOrderQty"))
		assert.Equal(t, price.String(), q.Get("price"))
		assert.Equal(t, clientOrderID, q.Get("newClientOrderId"))
		assert.Equal(t, "5000", q.Get("recvWindow"))
		assert.Equal(t, strconv.FormatInt(mockTime(), 10), q.Get("timestamp"))
		assert.NotEmpty(t, q.Get("signature"))
	})

	t.Run("minimal fields", func(t *testing.T) {
		svc := NewCreateOrderService("api-key", secretKey).
			Symbol("BTCUSDT").
			Side(OrderSideSell).
			Type(OrderTypeMarket)

		svc.timestamp = mockTime

		q := svc.buildQuery()

		assert.Equal(t, "BTCUSDT", q.Get("symbol"))
		assert.Equal(t, "SELL", q.Get("side"))
		assert.Equal(t, "MARKET", q.Get("type"))
		assert.Empty(t, q.Get("quantity"))
		assert.Empty(t, q.Get("quoteOrderQty"))
		assert.Empty(t, q.Get("price"))
		assert.Empty(t, q.Get("newClientOrderId"))
		assert.Empty(t, q.Get("recvWindow"))
		assert.Equal(t, strconv.FormatInt(mockTime(), 10), q.Get("timestamp"))
		assert.NotEmpty(t, q.Get("signature"))
	})
}

func TestCreateOrderService_Validate_ErrorsWrapped(t *testing.T) {
	svc := NewCreateOrderService("", "").Symbol("")

	err := svc.Validate()
	assert.Error(t, err)

	sdkErr, ok := err.(*sdkerr.SDKError)
	assert.True(t, ok)
	assert.Equal(t, sdkerr.ErrValidation, sdkErr.Kind())
	assert.Equal(t, "CreateOrderService.Validate", sdkErr.Op())
	assert.Contains(t, sdkErr.Message(), "symbol is required")
}

func TestCreateOrderService_Do_Success(t *testing.T) {
	expectedSymbol := "BTCUSDT"
	expectedSide := OrderSideBuy
	expectedType := OrderTypeLimit
	expectedQty := decimal.NewFromFloat(0.1)
	expectedPrice := decimal.NewFromFloat(50000.0)
	expectedClientOrderId := "test-order-id"
	expectedRecvWindow := int64(5000)

	fakeClient := &testutil.FakeHTTPClient{
		DoFunc: func(ctx context.Context, req *transport.Request) (*transport.Response, error) {
			assert.Equal(t, http.MethodPost, req.Method)
			assert.Contains(t, req.FullURL, "/api/v3/order")

			assert.NotEmpty(t, req.Headers.Get("X-MEXC-APIKEY"))
			assert.Equal(t, "application/json", req.Headers.Get("Content-Type"))

			query := testutil.ExtractQuery(t, req.FullURL)
			assert.Equal(t, expectedSymbol, query.Get("symbol"))
			assert.Equal(t, string(expectedSide), query.Get("side"))
			assert.Equal(t, string(expectedType), query.Get("type"))
			assert.Equal(t, expectedQty.String(), query.Get("quantity"))
			assert.Equal(t, expectedPrice.String(), query.Get("price"))
			assert.Equal(t, expectedClientOrderId, query.Get("newClientOrderId"))
			assert.Equal(t, strconv.FormatInt(expectedRecvWindow, 10), query.Get("recvWindow"))
			assert.NotEmpty(t, query.Get("timestamp"))

			return &transport.Response{
				StatusCode: 200,
				Body: []byte(`{
					"symbol": "BTCUSDT",
					"orderId": "123456",
					"orderListId": 0,
					"price": "50000.00",
					"origQty": "0.10000000",
					"type": "LIMIT",
					"side": "BUY",
					"transactTime": 1615555555555
				}`),
			}, nil
		},
	}

	svc := NewCreateOrderService("API_KEY", "SECRET_KEY").
		WithClient(fakeClient).
		Symbol(expectedSymbol).
		Side(expectedSide).
		Type(expectedType).
		Quantity(expectedQty).
		Price(expectedPrice).
		NewClientOrderId(expectedClientOrderId).
		RecvWindow(expectedRecvWindow)

	ctx := context.Background()
	result, err := svc.Do(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	assert.Equal(t, expectedSymbol, result.Symbol)
	assert.Equal(t, "123456", result.OrderID)
	assert.Equal(t, int64(0), result.OrderListID)
	assert.Equal(t, expectedType, result.Type)
	assert.Equal(t, expectedSide, result.Side)
	assert.Equal(t, int64(1615555555555), result.TransactTime)

	testutil.AssertDecimalEqual(t, result.Price, "50000.00", "order price mismatch")
	testutil.AssertDecimalEqual(t, result.OrigQty, "0.10000000", "order quantity mismatch")
}

func TestCreateOrderService_Do_Errors(t *testing.T) {
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

			svc := NewCreateOrderService("", "").
				WithClient(client).
				Symbol("BTCUSDT").
				Side(OrderSideBuy).
				Type(OrderTypeMarket)

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
