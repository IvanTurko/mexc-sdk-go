package wsuser

import (
	"testing"
	"time"

	"github.com/IvanTurko/mexc-sdk-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewSpotAccountOrdersSub(t *testing.T) {
	t.Run("normal flow", func(t *testing.T) {
		expCh := "spot@private.orders.v3.api.pb"

		sub := NewSpotAccountOrdersSub(func(*PrivateOrder) {}).(*spotAccountOrdersSub)
		assert.Equal(t, expCh, sub.streamName)
	})
	t.Run("onData is nil", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "onData function is nil")
		}()

		NewSpotAccountOrdersSub(nil)
	})
}

func TestSpotAccountOrdersSub_matches(t *testing.T) {
	sub := NewSpotAccountOrdersSub(func(*PrivateOrder) {}).(*spotAccountOrdersSub)

	msg := &message{
		Msg: sub.streamName,
	}

	ok, _ := sub.matches(msg)
	assert.True(t, ok)
}

func TestSpotAccountOrdersSub_acceptEvent(t *testing.T) {
	sub := NewSpotAccountOrdersSub(func(*PrivateOrder) {}).(*spotAccountOrdersSub)

	msg := &PushDataV3UserWrapper{
		Channel: sub.streamName,
	}
	assert.True(t, sub.acceptEvent(msg))
}

func TestSpotAccountOrdersSub_handleEvent(t *testing.T) {
	t.Run("calls onData for valid input", func(t *testing.T) {
		var called bool

		sub := NewSpotAccountOrdersSub(func(*PrivateOrder) {
			called = true
		}).(*spotAccountOrdersSub)

		msg := &PushDataV3UserWrapper{
			Body: &PushDataV3UserWrapper_PrivateOrders{
				PrivateOrders: &PrivateOrdersV3Api{
					Id:                 "123",
					ClientId:           "client-abc",
					Price:              "100.5",
					Quantity:           "0.01",
					Amount:             "1.005",
					AvgPrice:           "100.5",
					OrderType:          1,
					TradeType:          1,
					IsMaker:            true,
					RemainAmount:       "0.5",
					RemainQuantity:     "0.005",
					LastDealQuantity:   ptr("0.001"),
					CumulativeQuantity: "0.005",
					CumulativeAmount:   "0.5",
					Status:             1,
					CreateTime:         1234567890,
					Market:             ptr("BTCUSDT"),
					TriggerType:        ptr(int32(0)),
					TriggerPrice:       ptr("101"),
					State:              ptr(int32(0)),
					OcoId:              ptr("oco-123"),
					RouteFactor:        ptr("0.1"),
					SymbolId:           ptr("BTCUSDT"),
					MarketId:           ptr("BTC"),
					MarketCurrencyId:   ptr("BTC"),
					CurrencyId:         ptr("USDT"),
				},
			},
		}

		sub.handleEvent(msg)
		assert.True(t, called, "onData should be called")
	})

	t.Run("PrivateOrders is nil", func(t *testing.T) {
		sub := NewSpotAccountOrdersSub(func(*PrivateOrder) {
			t.Fatal("unexpected call to onData")
		}).(*spotAccountOrdersSub)

		sub.SetOnInvalid(func(error) {
			t.Fatal("unexpected call to onInvalid")
		})

		msg := &PushDataV3UserWrapper{
			Body: &PushDataV3UserWrapper_PrivateOrders{
				PrivateOrders: nil,
			},
		}

		sub.handleEvent(msg)
	})

	t.Run("calls onInvalid for invalid input", func(t *testing.T) {
		var invalidCalled bool

		sub := NewSpotAccountOrdersSub(func(*PrivateOrder) {}).(*spotAccountOrdersSub)
		sub.SetOnInvalid(func(err error) {
			invalidCalled = true
		})

		msg := &PushDataV3UserWrapper{
			Body: &PushDataV3UserWrapper_PrivateOrders{
				PrivateOrders: &PrivateOrdersV3Api{
					Id:                 "123",
					ClientId:           "client-abc",
					Price:              "INVALID",
					Quantity:           "0.01",
					Amount:             "1.005",
					AvgPrice:           "100.5",
					OrderType:          1,
					TradeType:          1,
					IsMaker:            true,
					RemainAmount:       "0.5",
					RemainQuantity:     "0.005",
					LastDealQuantity:   ptr("0.001"),
					CumulativeQuantity: "0.005",
					CumulativeAmount:   "0.5",
					Status:             1,
					CreateTime:         1234567890,
					Market:             ptr("BTCUSDT"),
					TriggerType:        ptr(int32(0)),
					TriggerPrice:       ptr("101"),
					State:              ptr(int32(0)),
					OcoId:              ptr("oco-123"),
					RouteFactor:        ptr("0.1"),
					SymbolId:           ptr("BTCUSDT"),
					MarketId:           ptr("BTC"),
					MarketCurrencyId:   ptr("BTC"),
					CurrencyId:         ptr("USDT"),
				},
			},
		}

		sub.handleEvent(msg)
		assert.True(t, invalidCalled, "onInvalid should be called")
	})
}

func Test_mapProtoPrivateOrder_Valid(t *testing.T) {
	sendTime := time.Now().Unix()

	lastDealQty := "5.5"
	msg := &PrivateOrdersV3Api{
		Id:                 "123",
		ClientId:           "456",
		Price:              "100.5",
		Quantity:           "2.0",
		Amount:             "201.0",
		AvgPrice:           "100.5",
		OrderType:          1,
		TradeType:          1,
		IsMaker:            true,
		RemainAmount:       "50.25",
		RemainQuantity:     "0.5",
		LastDealQuantity:   &lastDealQty,
		CumulativeQuantity: "1.5",
		CumulativeAmount:   "150.75",
		Status:             4,
		CreateTime:         123456789,
		Market:             ptr("BTCUSDT"),
		TriggerType:        ptr(int32(1)),
		TriggerPrice:       ptr("95.0"),
		State:              ptr(int32(2)),
		OcoId:              ptr("789"),
		RouteFactor:        ptr("factor"),
		SymbolId:           ptr("BTC"),
		MarketId:           ptr("1"),
		MarketCurrencyId:   ptr("2"),
		CurrencyId:         ptr("3"),
	}

	order, err := mapProtoPrivateOrder(msg, &sendTime)
	assert.NoError(t, err)

	assert.Equal(t, "123", order.Id)
	assert.Equal(t, "456", order.ClientId)
	assert.Equal(t, true, order.IsMaker)
	assert.Equal(t, int64(123456789), order.CreateTime)
	assert.Equal(t, ptr("BTCUSDT"), order.Market)
	assert.Equal(t, ptr(int32(1)), order.TriggerType)
	assert.Equal(t, ptr("95.0"), order.TriggerPrice)
	assert.Equal(t, ptr(int32(2)), order.State)
	assert.Equal(t, ptr("789"), order.OcoId)
	assert.Equal(t, ptr("factor"), order.RouteFactor)
	assert.Equal(t, ptr("BTC"), order.SymbolId)
	assert.Equal(t, ptr("1"), order.MarketId)
	assert.Equal(t, ptr("2"), order.MarketCurrencyId)
	assert.Equal(t, ptr("3"), order.CurrencyId)
	assert.Equal(t, &sendTime, order.SendTime)

	assert.Equal(t, OrderTypeLimitOrder, order.OrderType)
	assert.Equal(t, TradeSideBuy, order.TradeSide)
	assert.Equal(t, OrderStatusCanceled, order.Status)

	testutil.AssertDecimalEqual(t, order.Price, "100.5")
	testutil.AssertDecimalEqual(t, order.Quantity, "2.0")
	testutil.AssertDecimalEqual(t, order.Amount, "201.0")
	testutil.AssertDecimalEqual(t, order.AvgPrice, "100.5")
	testutil.AssertDecimalEqual(t, order.RemainAmount, "50.25")
	testutil.AssertDecimalEqual(t, order.RemainQuantity, "0.5")
	testutil.AssertDecimalEqual(t, *order.LastDealQuantity, "5.5")
	testutil.AssertDecimalEqual(t, order.CumulativeQuantity, "1.5")
	testutil.AssertDecimalEqual(t, order.CumulativeAmount, "150.75")
}

func Test_mapProtoPrivateOrder_InvalidFields(t *testing.T) {
	sendTime := int64(123456789)

	validMsg := func() *PrivateOrdersV3Api {
		lastDealQty := "5.5"
		return &PrivateOrdersV3Api{
			Id:                 "123",
			ClientId:           "456",
			Price:              "100.5",
			Quantity:           "2.0",
			Amount:             "201.0",
			AvgPrice:           "100.5",
			OrderType:          1,
			TradeType:          1,
			IsMaker:            true,
			RemainAmount:       "50.25",
			RemainQuantity:     "0.5",
			LastDealQuantity:   &lastDealQty,
			CumulativeQuantity: "1.5",
			CumulativeAmount:   "150.75",
			Status:             4,
			CreateTime:         123456789,
			Market:             ptr("BTCUSDT"),
			TriggerType:        ptr(int32(1)),
			TriggerPrice:       ptr("95.0"),
			State:              ptr(int32(2)),
			OcoId:              ptr("789"),
			RouteFactor:        ptr("factor"),
			SymbolId:           ptr("BTC"),
			MarketId:           ptr("1"),
			MarketCurrencyId:   ptr("2"),
			CurrencyId:         ptr("3"),
		}
	}

	t.Run("invalid price", func(t *testing.T) {
		msg := validMsg()
		msg.Price = "INVALID"
		_, err := mapProtoPrivateOrder(msg, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid price")
	})

	t.Run("invalid quantity", func(t *testing.T) {
		msg := validMsg()
		msg.Quantity = "INVALID"
		_, err := mapProtoPrivateOrder(msg, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid quantity")
	})

	t.Run("invalid amount", func(t *testing.T) {
		msg := validMsg()
		msg.Amount = "INVALID"
		_, err := mapProtoPrivateOrder(msg, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid amount")
	})

	t.Run("invalid avgPrice", func(t *testing.T) {
		msg := validMsg()
		msg.AvgPrice = "INVALID"
		_, err := mapProtoPrivateOrder(msg, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid avgPrice")
	})

	t.Run("invalid remainAmount", func(t *testing.T) {
		msg := validMsg()
		msg.RemainAmount = "INVALID"
		_, err := mapProtoPrivateOrder(msg, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid remainAmount")
	})

	t.Run("invalid remainQuantity", func(t *testing.T) {
		msg := validMsg()
		msg.RemainQuantity = "INVALID"
		_, err := mapProtoPrivateOrder(msg, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid remainQuantity")
	})

	t.Run("invalid lastDealQuantity", func(t *testing.T) {
		msg := validMsg()
		v := "INVALID"
		msg.LastDealQuantity = &v
		_, err := mapProtoPrivateOrder(msg, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid lastDealQuantity")
	})

	t.Run("invalid cumulativeQuantity", func(t *testing.T) {
		msg := validMsg()
		msg.CumulativeQuantity = "INVALID"
		_, err := mapProtoPrivateOrder(msg, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid cumulativeQuantity")
	})

	t.Run("invalid cumulativeAmount", func(t *testing.T) {
		msg := validMsg()
		msg.CumulativeAmount = "INVALID"
		_, err := mapProtoPrivateOrder(msg, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid cumulativeAmount")
	})

	t.Run("unknown order type", func(t *testing.T) {
		msg := validMsg()
		msg.OrderType = 999
		_, err := mapProtoPrivateOrder(msg, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "unknown order type")
	})

	t.Run("unknown trade side", func(t *testing.T) {
		msg := validMsg()
		msg.TradeType = 999
		_, err := mapProtoPrivateOrder(msg, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "unknown trade side")
	})

	t.Run("unknown order status", func(t *testing.T) {
		msg := validMsg()
		msg.Status = 999
		_, err := mapProtoPrivateOrder(msg, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "unknown order status")
	})
}

func ptr[T any](v T) *T { return &v }
