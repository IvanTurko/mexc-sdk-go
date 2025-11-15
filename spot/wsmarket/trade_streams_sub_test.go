package wsmarket

import (
	"fmt"
	"testing"
	"time"

	"github.com/IvanTurko/mexc-sdk-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func Test_NewTradeStreamsSub(t *testing.T) {
	t.Run("normal flow", func(t *testing.T) {
		symbol := "BTCUSDT"
		interval := Update10ms
		expCh := fmt.Sprintf("spot@public.aggre.deals.v3.api.pb@%s@%s", interval, symbol)

		sub := NewTradeStreamsSub(symbol, interval, func([]Trade) {}).(*tradeStreamsSub)
		assert.Equal(t, expCh, sub.streamName)
	})
	t.Run("invalid symbol name", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "invalid symbol name")
		}()

		NewTradeStreamsSub("", Update10ms, func([]Trade) {})
	})
	t.Run("invalid update interval", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "invalid update interval")
		}()

		NewTradeStreamsSub("BTCUSDT", UpdateInterval("hello"), func([]Trade) {})
	})
	t.Run("onData is nil", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "onData function is nil")
		}()

		NewTradeStreamsSub("BTCUSDT", Update10ms, nil)
	})
}

func TestTradeStreamsSub_matches(t *testing.T) {
	sub := NewTradeStreamsSub("BTCUSDT", Update10ms, func([]Trade) {}).(*tradeStreamsSub)

	msg := &message{
		Msg: sub.streamName,
	}
	ok, _ := sub.matches(msg)
	assert.True(t, ok)
}

func TestTradeStreamsSub_acceptEvent(t *testing.T) {
	sub := NewTradeStreamsSub("BTCUSDT", Update10ms, func([]Trade) {}).(*tradeStreamsSub)

	msg := &PushDataV3MarketWrapper{
		Channel: sub.streamName,
	}
	assert.True(t, sub.acceptEvent(msg))
}

func TestTradeStreamsSub_handleEvent(t *testing.T) {
	t.Run("calls onData for valid input", func(t *testing.T) {
		var called bool

		sub := NewTradeStreamsSub("BTCUSDT", Update10ms, func([]Trade) {
			called = true
		}).(*tradeStreamsSub)

		msg := &PushDataV3MarketWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3MarketWrapper_PublicAggreDeals{
				PublicAggreDeals: &PublicAggreDealsV3Api{
					Deals: []*PublicAggreDealsV3ApiItem{
						{
							Price:     "100",
							Quantity:  "1.5",
							TradeType: 1,
							Time:      12345678,
						},
					},
				},
			},
		}

		sub.handleEvent(msg)
		assert.True(t, called, "onData should be called")
	})

	t.Run("PublicAggreDeals is nil", func(t *testing.T) {
		sub := NewTradeStreamsSub("BTCUSDT", Update10ms, func([]Trade) {
			t.Fatal("unexpected call to onData")
		}).(*tradeStreamsSub)

		sub.SetOnInvalid(func(error) {
			t.Fatal("unexpected call to onInvalid")
		})

		msg := &PushDataV3MarketWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3MarketWrapper_PublicAggreDeals{
				PublicAggreDeals: nil,
			},
		}

		sub.handleEvent(msg)
	})

	t.Run("symbol is nil", func(t *testing.T) {
		sub := NewTradeStreamsSub("BTCUSDT", Update10ms, func([]Trade) {
			t.Fatal("unexpected call to onData")
		}).(*tradeStreamsSub)

		sub.SetOnInvalid(func(error) {
			t.Fatal("unexpected call to onInvalid")
		})

		msg := &PushDataV3MarketWrapper{
			Symbol: nil,
			Body: &PushDataV3MarketWrapper_PublicAggreDeals{
				PublicAggreDeals: &PublicAggreDealsV3Api{},
			},
		}

		sub.handleEvent(msg)
	})

	t.Run("calls onInvalid for invalid input", func(t *testing.T) {
		var invalidCalled bool

		sub := NewTradeStreamsSub("BTCUSDT", Update10ms, func([]Trade) {}).(*tradeStreamsSub)
		sub.SetOnInvalid(func(error) {
			invalidCalled = true
		})

		msg := &PushDataV3MarketWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3MarketWrapper_PublicAggreDeals{
				PublicAggreDeals: &PublicAggreDealsV3Api{
					Deals: []*PublicAggreDealsV3ApiItem{
						{
							Price:     "INVALID",
							Quantity:  "1.0",
							TradeType: 1,
							Time:      123,
						},
					},
				},
			},
		}

		sub.handleEvent(msg)
		assert.True(t, invalidCalled, "onInvalid should be called")
	})

	t.Run("trades not found", func(t *testing.T) {
		sub := NewTradeStreamsSub("BTCUSDT", Update10ms, func([]Trade) {}).(*tradeStreamsSub)
		sub.SetOnInvalid(func(err error) {
			assert.ErrorContains(t, err, "trades not found")
		})

		msg := &PushDataV3MarketWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3MarketWrapper_PublicAggreDeals{
				PublicAggreDeals: &PublicAggreDealsV3Api{},
			},
		}

		sub.handleEvent(msg)
	})
}

func Test_mapProtoTrades(t *testing.T) {
	sendTime := time.Now().Unix()
	symbol := "BTCUSDT"

	t.Run("valid trades", func(t *testing.T) {
		msg := &PublicAggreDealsV3Api{
			Deals: []*PublicAggreDealsV3ApiItem{
				{
					Price:     "30000.5",
					Quantity:  "0.01",
					TradeType: 1,
					Time:      1111,
				},
			},
		}

		trades, err := mapProtoTrades(msg, symbol, &sendTime)
		require.NoError(t, err)
		require.Len(t, trades, 1)

		t1 := trades[0]
		assert.Equal(t, symbol, t1.Symbol)
		assert.Equal(t, int64(1111), t1.Time)
		assert.Equal(t, TradeSideBuy, t1.Side)
		assert.Equal(t, &sendTime, t1.SendTime)
		testutil.AssertDecimalEqual(t, t1.Price, "30000.5")
		testutil.AssertDecimalEqual(t, t1.Quantity, "0.01")
	})

	t.Run("trade is nil", func(t *testing.T) {
		msg := &PublicAggreDealsV3Api{
			Deals: []*PublicAggreDealsV3ApiItem{nil},
		}

		_, err := mapProtoTrades(msg, symbol, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "trade is nil")
	})

	t.Run("invalid trades", func(t *testing.T) {
		msg := &PublicAggreDealsV3Api{
			Deals: []*PublicAggreDealsV3ApiItem{
				{
					Price:     "INVALID",
					Quantity:  "0.01",
					TradeType: 1,
					Time:      1111,
				},
			},
		}

		_, err := mapProtoTrades(msg, symbol, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid trade")
	})
}

func Test_mapProtoTrade(t *testing.T) {
	sendTime := time.Now().Unix()
	symbol := "ETHUSDT"

	t.Run("valid trade", func(t *testing.T) {
		item := &PublicAggreDealsV3ApiItem{
			Price:     "2500.55",
			Quantity:  "0.1",
			TradeType: 1,
			Time:      1234567890,
		}

		trade, err := mapProtoTrade(item, symbol, &sendTime)
		require.NoError(t, err)

		assert.Equal(t, symbol, trade.Symbol)
		assert.Equal(t, int64(1234567890), trade.Time)
		assert.Equal(t, TradeSideBuy, trade.Side)
		assert.Equal(t, &sendTime, trade.SendTime)

		testutil.AssertDecimalEqual(t, trade.Price, "2500.55")
		testutil.AssertDecimalEqual(t, trade.Quantity, "0.1")
	})

	t.Run("invalid price", func(t *testing.T) {
		item := &PublicAggreDealsV3ApiItem{
			Price:     "INVALID",
			Quantity:  "0.1",
			TradeType: 1,
			Time:      987654321,
		}

		_, err := mapProtoTrade(item, symbol, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid price")
	})

	t.Run("invalid quantity", func(t *testing.T) {
		item := &PublicAggreDealsV3ApiItem{
			Price:     "25",
			Quantity:  "INVALID",
			TradeType: 1,
			Time:      987654321,
		}

		_, err := mapProtoTrade(item, symbol, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid quantity")
	})

	t.Run("unknown trade side", func(t *testing.T) {
		item := &PublicAggreDealsV3ApiItem{
			Price:     "25",
			Quantity:  "0.1",
			TradeType: 99, // invalid
			Time:      987654321,
		}

		_, err := mapProtoTrade(item, symbol, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "unknown trade side")
	})
}
