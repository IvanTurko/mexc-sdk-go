package wsmarket

import (
	"encoding/json"
	"testing"

	"github.com/IvanTurko/mexc-sdk-go/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func Test_NewFairPriceSub(t *testing.T) {
	t.Run("normal flow", func(t *testing.T) {
		symbol := "BTC_USDT"

		sub := NewFairPriceSub(symbol, func(FairPrice) {}).(*fairPriceSub)
		assert.Equal(t, symbol, sub.symbol)
	})
	t.Run("invalid symbol name", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "invalid symbol name")
		}()

		NewFairPriceSub("", func(FairPrice) {})
	})
	t.Run("onData is nil", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "onData function is nil")
		}()

		NewFairPriceSub("BTC_USDT", nil)
	})
}

func TestFairPriceSub_matches(t *testing.T) {
	symbol := "BTC_USDT"
	sub := NewFairPriceSub(symbol, func(FairPrice) {}).(*fairPriceSub)

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

func TestFairPriceSub_acceptEvent(t *testing.T) {
	symbol := "BTC_USDT"
	sub := NewFairPriceSub(symbol, func(FairPrice) {}).(*fairPriceSub)

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

func TestFairPriceSub_handleEvent(t *testing.T) {
	symbol := "BTC_USDT"
	data := map[string]any{
		"price":  "12345.67",
		"symbol": symbol,
	}
	bytes, _ := json.Marshal(data)

	t.Run("handles valid data", func(t *testing.T) {
		var received FairPrice
		sub := NewFairPriceSub(symbol, func(fp FairPrice) {
			received = fp
		}).(*fairPriceSub)

		msg := &message{
			Data:   bytes,
			Symbol: symbol,
			Ts:     1234567890,
		}

		sub.handleEvent(msg)

		assert.Equal(t, symbol, received.Symbol)
		testutil.AssertDecimalEqual(t, received.Price, "12345.67")
		assert.Equal(t, int64(1234567890), *received.SendTime)
	})

	t.Run("handles invalid data", func(t *testing.T) {
		var calledErr error
		sub := NewFairPriceSub(symbol, func(FairPrice) {}).(*fairPriceSub)
		sub.SetOnInvalid(func(err error) {
			calledErr = err
		})

		msg := &message{
			Data: json.RawMessage(`{ invalid json }`),
		}

		sub.handleEvent(msg)
		assert.Error(t, calledErr)
		assert.Contains(t, calledErr.Error(), "failed to unmarshal")
	})
}
