package wsmarket

import (
	"encoding/json"
	"testing"

	"github.com/IvanTurko/mexc-sdk-go/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func Test_NewIndexPriceSub(t *testing.T) {
	t.Run("normal flow", func(t *testing.T) {
		symbol := "BTC_USDT"

		sub := NewIndexPriceSub(symbol, func(IndexPrice) {}).(*indexPriceSub)
		assert.Equal(t, symbol, sub.symbol)
	})
	t.Run("invalid symbol name", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "invalid symbol name")
		}()

		NewIndexPriceSub("", func(IndexPrice) {})
	})
	t.Run("onData is nil", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "onData function is nil")
		}()

		NewIndexPriceSub("BTC_USDT", nil)
	})
}

func TestIndexPriceSub_matches(t *testing.T) {
	symbol := "BTC_USDT"
	sub := NewIndexPriceSub(symbol, func(IndexPrice) {}).(*indexPriceSub)

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

func TestIndexPriceSub_acceptEvent(t *testing.T) {
	symbol := "BTC_USDT"
	sub := NewIndexPriceSub(symbol, func(IndexPrice) {}).(*indexPriceSub)

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

func TestIndexPriceSub_handleEvent(t *testing.T) {
	symbol := "BTC_USDT"
	data := map[string]any{
		"price":  "12345.67",
		"symbol": symbol,
	}
	bytes, _ := json.Marshal(data)

	t.Run("handles valid data", func(t *testing.T) {
		var received IndexPrice
		sub := NewIndexPriceSub(symbol, func(ip IndexPrice) {
			received = ip
		}).(*indexPriceSub)

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
		sub := NewIndexPriceSub(symbol, func(IndexPrice) {}).(*indexPriceSub)
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
