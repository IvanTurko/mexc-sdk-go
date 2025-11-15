package wsmarket

import (
	"encoding/json"
	"testing"

	"github.com/IvanTurko/mexc-sdk-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewBookTickerSub(t *testing.T) {
	t.Run("normal flow", func(t *testing.T) {
		symbol := "BTC_USDT"

		sub := NewBookTickerSub(symbol, func(*BookTicker) {}).(*bookTickerSub)
		assert.Equal(t, symbol, sub.symbol)
	})
	t.Run("invalid symbol name", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "invalid symbol name")
		}()

		NewBookTickerSub("", func(*BookTicker) {})
	})
	t.Run("onData is nil", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "onData function is nil")
		}()

		NewBookTickerSub("BTC_USDT", nil)
	})
}

func TestBookTickerSub_matches(t *testing.T) {
	symbol := "BTC_USDT"
	sub := NewBookTickerSub(symbol, func(*BookTicker) {}).(*bookTickerSub)

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

func TestBookTickerSub_acceptEvent(t *testing.T) {
	symbol := "BTC_USDT"
	sub := NewBookTickerSub(symbol, func(*BookTicker) {}).(*bookTickerSub)

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

func TestBookTickerSub_handleEvent(t *testing.T) {
	symbol := "BTC_USDT"
	data := map[string]any{
		"symbol":        symbol,
		"lastPrice":     "100.0",
		"bid1":          "99.9",
		"ask1":          "100.1",
		"volume24":      "1000.0",
		"holdVol":       "500.0",
		"lower24Price":  "90.0",
		"high24Price":   "110.0",
		"riseFallRate":  "0.01",
		"riseFallValue": "1.0",
		"indexPrice":    "99.5",
		"fairPrice":     "99.6",
		"fundingRate":   "0.0001",
		"timestamp":     123456789,
	}
	bytes, _ := json.Marshal(data)

	t.Run("handles valid data", func(t *testing.T) {
		var received *BookTicker
		sub := NewBookTickerSub(symbol, func(d *BookTicker) {
			received = d
		}).(*bookTickerSub)

		msg := &message{
			Data:   bytes,
			Symbol: symbol,
			Ts:     1610005070157,
		}

		sub.handleEvent(msg)

		assert.NotNil(t, received.SendTime)
		assert.Equal(t, symbol, received.Symbol)
		testutil.AssertDecimalEqual(t, received.LastPrice, "100")
		testutil.AssertDecimalEqual(t, received.Bid1, "99.9")
		testutil.AssertDecimalEqual(t, received.Ask1, "100.1")
		testutil.AssertDecimalEqual(t, received.Volume24, "1000")
		testutil.AssertDecimalEqual(t, received.HoldVol, "500")
		testutil.AssertDecimalEqual(t, received.Lower24Price, "90")
		testutil.AssertDecimalEqual(t, received.High24Price, "110")
		testutil.AssertDecimalEqual(t, received.RiseFallRate, "0.01")
		testutil.AssertDecimalEqual(t, received.RiseFallValue, "1")
		testutil.AssertDecimalEqual(t, received.IndexPrice, "99.5")
		testutil.AssertDecimalEqual(t, received.FairPrice, "99.6")
		testutil.AssertDecimalEqual(t, received.FundingRate, "0.0001")
		assert.Equal(t, int64(123456789), received.Timestamp)
		assert.Equal(t, int64(1610005070157), *received.SendTime)
	})

	t.Run("handles invalid data", func(t *testing.T) {
		var calledErr error
		sub := NewBookTickerSub(symbol, func(*BookTicker) {}).(*bookTickerSub)
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
