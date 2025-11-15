package wsmarket

import (
	"encoding/json"
	"testing"

	"github.com/IvanTurko/mexc-sdk-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewBookTickerBatchSub(t *testing.T) {
	defer func() {
		r := recover()
		assert.Contains(t, r, "onData function is nil")
	}()

	NewBookTickerBatchSub(nil)
}

func TestBookTickerBatchSub_matches(t *testing.T) {
	sub := NewBookTickerBatchSub(func(*BookTickerBatch) {}).(*bookTickerBatchSub)

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

func TestBookTickerBatchSub_acceptEvent(t *testing.T) {
	sub := NewBookTickerBatchSub(func(*BookTickerBatch) {}).(*bookTickerBatchSub)

	t.Run("accepts valid push event", func(t *testing.T) {
		msg := &message{
			Channel: "push." + sub.channel(),
		}
		assert.True(t, sub.acceptEvent(msg))
	})

	t.Run("rejects invalid channel", func(t *testing.T) {
		msg := &message{
			Channel: "push.invalid",
		}
		assert.False(t, sub.acceptEvent(msg))
	})
}

func TestBookTickerBatchSub_handleEvent(t *testing.T) {
	symbol := "BTC_USDT"
	data := []map[string]any{
		{
			"fairPrice":    "183.01",
			"lastPrice":    "183",
			"riseFallRate": "-0.0708",
			"symbol":       symbol,
			"volume24":     "200",
		},
	}
	bytes, _ := json.Marshal(data)

	t.Run("handles valid data", func(t *testing.T) {
		var received *BookTickerBatch
		sub := NewBookTickerBatchSub(func(d *BookTickerBatch) {
			received = d
		}).(*bookTickerBatchSub)

		msg := &message{
			Data: bytes,
			Ts:   1587442022003,
		}

		sub.handleEvent(msg)

		require.NotNil(t, received)
		require.Len(t, received.Tickers, 1)

		assert.Equal(t, symbol, received.Tickers[0].Symbol)
		testutil.AssertDecimalEqual(t, received.Tickers[0].FairPrice, "183.01")
		testutil.AssertDecimalEqual(t, received.Tickers[0].LastPrice, "183")
		testutil.AssertDecimalEqual(t, received.Tickers[0].RiseFallRate, "-0.0708")
		testutil.AssertDecimalEqual(t, received.Tickers[0].Volume24, "200")
		assert.Equal(t, int64(1587442022003), *received.SendTime)
	})

	t.Run("handles invalid data", func(t *testing.T) {
		var calledErr error
		sub := NewBookTickerBatchSub(func(*BookTickerBatch) {}).(*bookTickerBatchSub)
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
