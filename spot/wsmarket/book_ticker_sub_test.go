package wsmarket

import (
	"fmt"
	"testing"

	"github.com/IvanTurko/mexc-sdk-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func Test_NewBookTickerSub(t *testing.T) {
	t.Run("normal flow", func(t *testing.T) {
		symbol := "BTCUSDT"
		interval := Update100ms
		expCh := fmt.Sprintf("spot@public.aggre.bookTicker.v3.api.pb@%s@%s", interval, symbol)

		sub := NewBookTickerSub(symbol, interval, func(L1Ticker) {}).(*bookTickerSub)
		assert.Equal(t, expCh, sub.streamName)
	})
	t.Run("invalid symbol name", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "invalid symbol name")
		}()

		NewBookTickerSub("", Update100ms, func(L1Ticker) {})
	})
	t.Run("invalid update interval", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "invalid update interval")
		}()

		NewBookTickerSub("BTCUSDT", UpdateInterval("hello"), func(L1Ticker) {})
	})
	t.Run("onData is nil", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "onData function is nil")
		}()

		NewBookTickerSub("BTCUSDT", Update100ms, nil)
	})
}

func TestBookTickerSub_matches(t *testing.T) {
	sub := NewBookTickerSub("BTCUSDT", Update10ms, func(L1Ticker) {}).(*bookTickerSub)

	msg := &message{
		Msg: sub.streamName,
	}

	ok, _ := sub.matches(msg)
	assert.True(t, ok)
}

func TestBookTickerSub_acceptEvent(t *testing.T) {
	sub := NewBookTickerSub("BTCUSDT", Update10ms, func(L1Ticker) {}).(*bookTickerSub)

	msg := &PushDataV3MarketWrapper{
		Channel: sub.streamName,
	}
	assert.True(t, sub.acceptEvent(msg))
}

func TestBookTickerSub_handleEvent(t *testing.T) {
	t.Run("calls onData for valid input", func(t *testing.T) {
		var called bool

		sub := NewBookTickerSub("BTCUSDT", Update10ms, func(L1Ticker) {
			called = true
		}).(*bookTickerSub)

		msg := &PushDataV3MarketWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3MarketWrapper_PublicAggreBookTicker{
				PublicAggreBookTicker: &PublicAggreBookTickerV3Api{
					BidPrice:    "192.921",
					BidQuantity: "8329.9",
					AskPrice:    "3838",
					AskQuantity: "92939",
				},
			},
		}

		sub.handleEvent(msg)
		assert.True(t, called, "onData should be called")
	})

	t.Run("PublicAggreBookTicker is nil", func(t *testing.T) {
		sub := NewBookTickerSub("BTCUSDT", Update10ms, func(L1Ticker) {
			t.Fatal("unexpected call to onData")
		}).(*bookTickerSub)

		sub.SetOnInvalid(func(error) {
			t.Fatal("unexpected call to onInvalid")
		})

		msg := &PushDataV3MarketWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3MarketWrapper_PublicAggreBookTicker{
				PublicAggreBookTicker: nil,
			},
		}

		sub.handleEvent(msg)
	})

	t.Run("symbol is nil", func(t *testing.T) {
		sub := NewBookTickerSub("BTCUSDT", Update10ms, func(L1Ticker) {
			t.Fatal("unexpected call to onData")
		}).(*bookTickerSub)

		sub.SetOnInvalid(func(error) {
			t.Fatal("unexpected call to onInvalid")
		})

		msg := &PushDataV3MarketWrapper{
			Symbol: nil,
			Body: &PushDataV3MarketWrapper_PublicAggreBookTicker{
				PublicAggreBookTicker: &PublicAggreBookTickerV3Api{},
			},
		}

		sub.handleEvent(msg)
	})

	t.Run("calls onInvalid for invalid input", func(t *testing.T) {
		var invalidCalled bool

		sub := NewBookTickerSub("BTCUSDT", Update10ms, func(L1Ticker) {}).(*bookTickerSub)
		sub.SetOnInvalid(func(error) {
			invalidCalled = true
		})

		msg := &PushDataV3MarketWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3MarketWrapper_PublicAggreBookTicker{
				PublicAggreBookTicker: &PublicAggreBookTickerV3Api{
					BidPrice:    "192.921",
					BidQuantity: "8329.9",
					AskPrice:    "INVALID",
					AskQuantity: "92939",
				},
			},
		}

		sub.handleEvent(msg)
		assert.True(t, invalidCalled, "onInvalid should be called")
	})
}

func Test_mapProtoL1Ticker(t *testing.T) {
	t.Run("valid l1 book ticker", func(t *testing.T) {
		sendTime := int64(1234567890)

		msg := &PublicAggreBookTickerV3Api{
			BidPrice:    "100.1",
			BidQuantity: "5.5",
			AskPrice:    "100.2",
			AskQuantity: "6.6",
		}

		got, err := mapProtoL1Ticker(msg, "BTCUSDT", &sendTime)
		require.NoError(t, err)

		assert.Equal(t, "BTCUSDT", got.Symbol)
		assert.Equal(t, &sendTime, got.SendTime)

		testutil.AssertDecimalEqual(t, got.BidPrice, "100.1")
		testutil.AssertDecimalEqual(t, got.BidQuantity, "5.5")
		testutil.AssertDecimalEqual(t, got.AskPrice, "100.2")
		testutil.AssertDecimalEqual(t, got.AskQuantity, "6.6")
	})

	t.Run("invalid bidPrice", func(t *testing.T) {
		msg := &PublicAggreBookTickerV3Api{
			BidPrice:    "INVALID",
			BidQuantity: "5.5",
			AskPrice:    "100.2",
			AskQuantity: "6.6",
		}

		_, err := mapProtoL1Ticker(msg, "BTCUSDT", nil)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid bidPrice")
	})

	t.Run("invalid bidQuantity", func(t *testing.T) {
		msg := &PublicAggreBookTickerV3Api{
			BidPrice:    "100.1",
			BidQuantity: "INVALID",
			AskPrice:    "100.2",
			AskQuantity: "6.6",
		}

		_, err := mapProtoL1Ticker(msg, "BTCUSDT", nil)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid bidQuantity")
	})

	t.Run("invalid askPrice", func(t *testing.T) {
		msg := &PublicAggreBookTickerV3Api{
			BidPrice:    "100.1",
			BidQuantity: "5.5",
			AskPrice:    "INVALID",
			AskQuantity: "6.6",
		}

		_, err := mapProtoL1Ticker(msg, "BTCUSDT", nil)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid askPrice")
	})

	t.Run("invalid askQuantity", func(t *testing.T) {
		msg := &PublicAggreBookTickerV3Api{
			BidPrice:    "100.1",
			BidQuantity: "5.5",
			AskPrice:    "100.2",
			AskQuantity: "INVALID",
		}

		_, err := mapProtoL1Ticker(msg, "BTCUSDT", nil)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid askQuantity")
	})
}
