package wsmarket

import (
	"encoding/json"
	"testing"

	"github.com/IvanTurko/mexc-sdk-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewKlineSub(t *testing.T) {
	t.Run("normal flow", func(t *testing.T) {
		symbol := "BTC_USDT"
		interval := Kline5Min

		sub := NewKlineSub(symbol, interval, func(Kline) {}).(*klineSub)
		assert.Equal(t, symbol, sub.symbol)
		assert.Equal(t, interval, sub.interval)
	})
	t.Run("invalid symbol name", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "invalid symbol name")
		}()

		NewKlineSub("", Kline5Min, func(Kline) {})
	})
	t.Run("invalid interval", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "invalid interval")
		}()

		NewKlineSub("BTC_USDT", "invalid", func(Kline) {})
	})
	t.Run("onData is nil", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "onData function is nil")
		}()

		NewKlineSub("BTC_USDT", Kline5Min, nil)
	})
}

func TestKlineSub_matches(t *testing.T) {
	symbol := "BTC_USDT"
	interval := Kline5Min
	sub := NewKlineSub(symbol, interval, func(Kline) {}).(*klineSub)

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

func TestKlineSub_acceptEvent(t *testing.T) {
	symbol := "BTC_USDT"
	interval := Kline5Min
	sub := NewKlineSub(symbol, interval, func(Kline) {}).(*klineSub)

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

func TestKlineSub_handleEvent(t *testing.T) {
	symbol := "BTC_USDT"
	interval := Kline5Min
	data := map[string]any{
		"a":        "233.74",
		"c":        "6885",
		"h":        "6910.5",
		"interval": "Min5",
		"l":        "6885",
		"o":        "6894.5",
		"q":        "1611754",
		"symbol":   symbol,
		"t":        1587448800,
	}
	bytes, _ := json.Marshal(data)

	t.Run("handles valid data", func(t *testing.T) {
		var received Kline
		sub := NewKlineSub(symbol, interval, func(k Kline) {
			received = k
		}).(*klineSub)

		msg := &message{
			Data:   bytes,
			Symbol: symbol,
			Ts:     1234567890,
		}

		sub.handleEvent(msg)

		assert.Equal(t, symbol, received.Symbol)
		testutil.AssertDecimalEqual(t, received.Amount, "233.74")
		testutil.AssertDecimalEqual(t, received.Close, "6885")
		testutil.AssertDecimalEqual(t, received.High, "6910.5")
		assert.Equal(t, interval, received.Interval)
		testutil.AssertDecimalEqual(t, received.Low, "6885")
		testutil.AssertDecimalEqual(t, received.Open, "6894.5")
		testutil.AssertDecimalEqual(t, received.Quantity, "1611754")
		assert.Equal(t, int64(1587448800), received.Timestamp)
		assert.Equal(t, int64(1234567890), *received.SendTime)
	})

	t.Run("handles invalid data", func(t *testing.T) {
		var calledErr error
		sub := NewKlineSub(symbol, interval, func(Kline) {}).(*klineSub)
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
