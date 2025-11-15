package wsuser

import (
	"encoding/json"
	"testing"

	"github.com/IvanTurko/mexc-sdk-go/internal/testutil"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewOrderSub(t *testing.T) {
	defer func() {
		r := recover()
		assert.Contains(t, r, "onData function is nil")
	}()

	NewOrderSub(nil)
}

func TestOrderSub_acceptEvent(t *testing.T) {
	sub := NewOrderSub(func(*Order) {}).(*orderSub)

	t.Run("accepts valid push event", func(t *testing.T) {
		msg := &message{
			Channel: "push.personal." + sub.channel(),
		}
		assert.True(t, sub.acceptEvent(msg))
	})

	t.Run("rejects invalid channel", func(t *testing.T) {
		msg := &message{
			Channel: "push.personal.invalid",
		}
		assert.False(t, sub.acceptEvent(msg))
	})
}

func TestOrderSub_handleEvent(t *testing.T) {
	data := map[string]any{
		"orderId":      "testOrderId",
		"symbol":       "testSymbol",
		"positionId":   123,
		"price":        decimal.NewFromFloat(100.0),
		"vol":          decimal.NewFromFloat(1.0),
		"leverage":     decimal.NewFromFloat(10.0),
		"side":         1, // OPEN_LONG
		"category":     1, // LIMIT_ORDER
		"orderType":    1, // LIMIT
		"dealAvgPrice": decimal.NewFromFloat(99.0),
		"dealVol":      decimal.NewFromFloat(0.5),
		"orderMargin":  decimal.NewFromFloat(50.0),
		"usedMargin":   decimal.NewFromFloat(25.0),
		"takerFee":     decimal.NewFromFloat(0.001),
		"makerFee":     decimal.NewFromFloat(0.0005),
		"profit":       decimal.NewFromFloat(10.0),
		"feeCurrency":  "USDT",
		"openType":     1, // ISOLATED
		"state":        2, // UNCOMPLETED
		"errorCode":    0, // NORMAL
		"externalOid":  "external123",
		"createTime":   1678886400000,
		"updateTime":   1678886500000,
	}
	bytes, _ := json.Marshal(data)

	t.Run("handles valid data", func(t *testing.T) {
		var received *Order
		sub := NewOrderSub(func(d *Order) {
			received = d
		}).(*orderSub)

		msg := &message{
			Data: bytes,
			Ts:   1678886600000,
		}

		sub.handleEvent(msg)

		require.NotNil(t, received)
		assert.Equal(t, "testOrderId", received.OrderId)
		assert.Equal(t, OrderSideOpenLong, received.Side)
		assert.Equal(t, OrderCategoryLimitOrder, received.Category)
		assert.Equal(t, OrderTypeLimit, received.OrderType)
		assert.Equal(t, OpenTypeIsolated, received.OpenType)
		assert.Equal(t, OrderStateUncompleted, received.State)
		assert.Equal(t, OrderErrorCodeNormal, received.ErrorCode)
		assert.Equal(t, int64(1678886600000), *received.SendTime)
	})

	t.Run("handles invalid data", func(t *testing.T) {
		var calledErr error
		sub := NewOrderSub(func(*Order) {}).(*orderSub)
		sub.SetOnInvalid(func(err error) {
			calledErr = err
		})

		msg := &message{
			Data: json.RawMessage(`{ invalid json }`),
		}

		sub.handleEvent(msg)
		require.Error(t, calledErr)
		assert.Contains(t, calledErr.Error(), "failed to unmarshal")
	})
}

func TestOrder_UnmarshalJSON(t *testing.T) {
	baseOrderData := func() map[string]any {
		return map[string]any{
			"orderId":      "id1",
			"symbol":       "BTC_USDT",
			"positionId":   int64(1),
			"price":        "100.0",
			"vol":          "1.0",
			"leverage":     "10.0",
			"side":         1,
			"category":     1,
			"orderType":    1,
			"dealAvgPrice": "99.0",
			"dealVol":      "0.5",
			"orderMargin":  "50.0",
			"usedMargin":   "25.0",
			"takerFee":     "0.001",
			"makerFee":     "0.0005",
			"profit":       "10.0",
			"feeCurrency":  "USDT",
			"openType":     1,
			"state":        2,
			"errorCode":    0,
			"externalOid":  "ext1",
			"createTime":   int64(1678886400000),
			"updateTime":   int64(1678886500000),
		}
	}

	t.Run("valid order", func(t *testing.T) {
		data, err := json.Marshal(baseOrderData())
		require.NoError(t, err)

		var order Order
		err = json.Unmarshal(data, &order)

		require.NoError(t, err)
		assert.Equal(t, "id1", order.OrderId)
		assert.Equal(t, "BTC_USDT", order.Symbol)
		assert.Equal(t, int64(1), order.PositionId)
		testutil.AssertDecimalEqual(t, order.Price, "100.0")
		testutil.AssertDecimalEqual(t, order.Vol, "1.0")
		testutil.AssertDecimalEqual(t, order.Leverage, "10.0")
		assert.Equal(t, OrderSideOpenLong, order.Side)
		assert.Equal(t, OrderCategoryLimitOrder, order.Category)
		assert.Equal(t, OrderTypeLimit, order.OrderType)
		testutil.AssertDecimalEqual(t, order.DealAvgPrice, "99.0")
		testutil.AssertDecimalEqual(t, order.DealVol, "0.5")
		testutil.AssertDecimalEqual(t, order.OrderMargin, "50.0")
		testutil.AssertDecimalEqual(t, order.UsedMargin, "25.0")
		testutil.AssertDecimalEqual(t, order.TakerFee, "0.001")
		testutil.AssertDecimalEqual(t, order.MakerFee, "0.0005")
		testutil.AssertDecimalEqual(t, order.Profit, "10.0")
		assert.Equal(t, "USDT", order.FeeCurrency)
		assert.Equal(t, OpenTypeIsolated, order.OpenType)
		assert.Equal(t, OrderStateUncompleted, order.State)
		assert.Equal(t, OrderErrorCodeNormal, order.ErrorCode)
		assert.Equal(t, "ext1", order.ExternalOid)
		assert.Equal(t, int64(1678886400000), order.CreateTime)
		assert.Equal(t, int64(1678886500000), order.UpdateTime)
	})

	t.Run("invalid side", func(t *testing.T) {
		dataMap := baseOrderData()
		dataMap["side"] = 99
		data, err := json.Marshal(dataMap)
		require.NoError(t, err)

		var order Order
		err = json.Unmarshal(data, &order)
		require.Error(t, err)
		assert.ErrorContains(t, err, "unknown order side code: 99")
	})

	t.Run("invalid order type", func(t *testing.T) {
		dataMap := baseOrderData()
		dataMap["orderType"] = 99
		data, err := json.Marshal(dataMap)
		require.NoError(t, err)

		var order Order
		err = json.Unmarshal(data, &order)
		require.Error(t, err)
		assert.ErrorContains(t, err, "unknown order type code: 99")
	})

	t.Run("invalid category", func(t *testing.T) {
		dataMap := baseOrderData()
		dataMap["category"] = 99
		data, err := json.Marshal(dataMap)
		require.NoError(t, err)

		var order Order
		err = json.Unmarshal(data, &order)
		require.Error(t, err)
		assert.ErrorContains(t, err, "unknown order category code: 99")
	})

	t.Run("invalid open type", func(t *testing.T) {
		dataMap := baseOrderData()
		dataMap["openType"] = 99
		data, err := json.Marshal(dataMap)
		require.NoError(t, err)

		var order Order
		err = json.Unmarshal(data, &order)
		require.Error(t, err)
		assert.ErrorContains(t, err, "unknown open type code: 99")
	})

	t.Run("invalid state", func(t *testing.T) {
		dataMap := baseOrderData()
		dataMap["state"] = 99
		data, err := json.Marshal(dataMap)
		require.NoError(t, err)

		var order Order
		err = json.Unmarshal(data, &order)
		require.Error(t, err)
		assert.ErrorContains(t, err, "unknown order state code: 99")
	})

	t.Run("invalid error code", func(t *testing.T) {
		dataMap := baseOrderData()
		dataMap["errorCode"] = 99
		data, err := json.Marshal(dataMap)
		require.NoError(t, err)

		var order Order
		err = json.Unmarshal(data, &order)
		require.Error(t, err)
		assert.ErrorContains(t, err, "unknown order error code: 99")
	})
}
