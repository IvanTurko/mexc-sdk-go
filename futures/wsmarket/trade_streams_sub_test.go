package wsmarket

import (
	"encoding/json"
	"testing"

	"github.com/IvanTurko/mexc-sdk-go/internal/testutil"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewTradeStreamsSub(t *testing.T) {
	t.Run("normal flow", func(t *testing.T) {
		symbol := "BTC_USDT"

		sub := NewTradeStreamsSub(symbol, func(Trade) {}).(*tradeStreamsSub)
		assert.Equal(t, symbol, sub.symbol)
	})
	t.Run("invalid symbol name", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "invalid symbol name")
		}()

		NewTradeStreamsSub("", func(Trade) {})
	})
	t.Run("onData is nil", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "onData function is nil")
		}()

		NewTradeStreamsSub("BTC_USDT", nil)
	})
}

func TestTradeStreamsSub_matches(t *testing.T) {
	symbol := "BTC_USDT"
	sub := NewTradeStreamsSub(symbol, func(Trade) {}).(*tradeStreamsSub)

	t.Run("success match", func(t *testing.T) {
		msg := &message{
			Channel: "rs.sub." + sub.channel(),
			Data:    json.RawMessage(`"success"`),
		}
		ok, err := sub.matches(msg)
		assert.True(t, ok, "expected match on success")
		assert.NoError(t, err)
	})

	t.Run("invalid success payload", func(t *testing.T) {
		msg := &message{
			Channel: "rs.sub." + sub.channel(),
			Data:    json.RawMessage(`"failure"`),
		}
		ok, err := sub.matches(msg)
		assert.False(t, ok, "expected no match on non-success")
		assert.NoError(t, err)
	})

	t.Run("error channel match", func(t *testing.T) {
		msg := &message{
			Channel: "rs.error",
			Data:    json.RawMessage(`"some error occurred"`),
		}
		ok, err := sub.matches(msg)
		assert.True(t, ok, "expected match on error channel")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sub failed")
	})

	t.Run("unrelated channel", func(t *testing.T) {
		msg := &message{
			Channel: "rs.unknown",
			Data:    json.RawMessage(`"whatever"`),
		}
		ok, err := sub.matches(msg)
		assert.False(t, ok, "expected no match on unrelated channel")
		assert.NoError(t, err)
	})
}

func TestTradeStreamsSub_acceptEvent(t *testing.T) {
	symbol := "BTC_USDT"
	sub := NewTradeStreamsSub(symbol, func(Trade) {}).(*tradeStreamsSub)

	t.Run("accepts valid push event", func(t *testing.T) {
		msg := &message{
			Channel: "push." + sub.channel(),
			Symbol:  symbol,
		}
		assert.True(t, sub.acceptEvent(msg))
	})

	t.Run("rejects invalid channel", func(t *testing.T) {
		msg := &message{
			Channel: "push.invalid",
			Symbol:  symbol,
		}
		assert.False(t, sub.acceptEvent(msg))
	})

	t.Run("rejects invalid symbol", func(t *testing.T) {
		msg := &message{
			Channel: "push." + sub.channel(),
			Symbol:  "OTHER_SYMBOL",
		}
		assert.False(t, sub.acceptEvent(msg))
	})
}

func TestTradeStreamsSub_handleEvent(t *testing.T) {
	symbol := "BTC_USDT"
	data := tradeJSON{
		Price:        decimal.NewFromInt(10),
		Volume:       decimal.NewFromInt(1),
		Side:         1,
		OpenType:     1,
		AutoTransact: 2,
		Timestamp:    123456789,
	}
	bytes, _ := json.Marshal(data)

	t.Run("handles valid data", func(t *testing.T) {
		var received Trade
		sub := NewTradeStreamsSub(symbol, func(d Trade) {
			received = d
		}).(*tradeStreamsSub)

		msg := &message{
			Data: bytes,
			Ts:   1610005070157,
		}

		sub.handleEvent(msg)

		assert.NotNil(t, received.SendTime)
		assert.Equal(t, int64(1610005070157), *received.SendTime)
		testutil.AssertDecimalEqual(t, received.Price, "10")
		testutil.AssertDecimalEqual(t, received.Volume, "1")
		assert.Equal(t, TradeSideBuy, received.Side)
		assert.Equal(t, OpenTypeOpen, received.OpenType)
		assert.Equal(t, AutoTransactNo, received.IsAutoTransact)
		assert.Equal(t, int64(123456789), received.Timestamp)
		assert.Equal(t, int64(1610005070157), *received.SendTime)
	})

	t.Run("handles invalid data", func(t *testing.T) {
		var calledErr error
		sub := NewTradeStreamsSub(symbol, func(Trade) {}).(*tradeStreamsSub)
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

func TestTrade_UnmarshalJSON(t *testing.T) {
	t.Run("valid value", func(t *testing.T) {
		data := []byte(`{
			"p": "30500.5",
			"v": "0.002",
			"T": 1,
			"O": 2,
			"M": 1,
			"t": 1688192023000
		}`)

		var trade Trade
		err := json.Unmarshal(data, &trade)

		require.NoError(t, err)
		testutil.AssertDecimalEqual(t, trade.Price, "30500.5")
		testutil.AssertDecimalEqual(t, trade.Volume, "0.002")

		assert.Equal(t, TradeSideBuy, trade.Side)
		assert.Equal(t, OpenTypeClose, trade.OpenType)
		assert.Equal(t, AutoTransactYes, trade.IsAutoTransact)
		assert.Equal(t, int64(1688192023000), trade.Timestamp)
	})

	t.Run("invalid trade side", func(t *testing.T) {
		data := []byte(`{
			"p": "30500.5",
			"v": "0.002",
			"T": 999,
			"O": 2,
			"M": 1,
			"t": 1688192023000
		}`)

		var trade Trade
		err := json.Unmarshal(data, &trade)
		require.Error(t, err)
		assert.ErrorContains(t, err, "unknown trade side code")
	})

	t.Run("invalid open type", func(t *testing.T) {
		data := []byte(`{
			"p": "30500.5",
			"v": "0.002",
			"T": 1,
			"O": 999,
			"M": 1,
			"t": 1688192023000
		}`)

		var trade Trade
		err := json.Unmarshal(data, &trade)
		require.Error(t, err)
		assert.ErrorContains(t, err, "unknown open type code")
	})

	t.Run("invalid auto transact", func(t *testing.T) {
		data := []byte(`{
			"p": "30500.5",
			"v": "0.002",
			"T": 1,
			"O": 2,
			"M": 999,
			"t": 1688192023000
		}`)

		var trade Trade
		err := json.Unmarshal(data, &trade)
		require.Error(t, err)
		assert.ErrorContains(t, err, "unknown auto transact code")
	})
}
