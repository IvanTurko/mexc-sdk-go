package wsmarket

import (
	"encoding/json"
	"testing"

	"github.com/IvanTurko/mexc-sdk-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewBookDepthSub(t *testing.T) {
	t.Run("normal flow", func(t *testing.T) {
		symbol := "BTC_USDT"

		sub := NewBookDepthSub(symbol, func(*DepthSnapshot) {}).(*bookDepthSub)
		assert.Equal(t, symbol, sub.symbol)
	})
	t.Run("invalid symbol name", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "invalid symbol name")
		}()

		NewBookDepthSub("", func(*DepthSnapshot) {})
	})
	t.Run("onData is nil", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "onData function is nil")
		}()

		NewBookDepthSub("BTC_USDT", nil)
	})
}

func TestBookDepthSub_matches(t *testing.T) {
	symbol := "BTC_USDT"
	sub := NewBookDepthSub(symbol, func(*DepthSnapshot) {}).(*bookDepthSub)

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

func TestBookDepthSub_acceptEvent(t *testing.T) {
	symbol := "BTC_USDT"
	sub := NewBookDepthSub(symbol, func(*DepthSnapshot) {}).(*bookDepthSub)

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

func TestBookDepthSub_handleEvent(t *testing.T) {
	symbol := "BTC_USDT"

	data := depthSnapshotJSON{
		Asks: [][]any{
			{411.8, 10, 1},
			{412.5, 5, 2},
		},
		Bids: [][]any{
			{410.0, 8, 1},
		},
		Version: 12345678,
	}
	bytes, _ := json.Marshal(data)

	t.Run("handles valid data", func(t *testing.T) {
		var received *DepthSnapshot
		sub := NewBookDepthSub(symbol, func(d *DepthSnapshot) {
			received = d
		}).(*bookDepthSub)

		msg := &message{
			Data: bytes,
			Ts:   1610005070157,
		}

		sub.handleEvent(msg)

		assert.NotNil(t, received)
		assert.NotNil(t, received.SendTime)
		assert.Equal(t, int64(1610005070157), *received.SendTime)

		assert.Len(t, received.Asks, 2)
		assert.Len(t, received.Bids, 1)
		assert.Equal(t, int64(12345678), received.Version)
	})

	t.Run("handles invalid data", func(t *testing.T) {
		var calledErr error
		sub := NewBookDepthSub(symbol, func(*DepthSnapshot) {}).(*bookDepthSub)
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

func TestDepthSnapshot_UnmarshalJSON_Valid(t *testing.T) {
	jsonData := []byte(`{
		"asks": [
			[411.8, 10, 1],
			[412.5, 5, 2]
		],
		"bids": [
			[410.0, 8, 1]
		],
		"version": 12345678
	}`)

	var snapshot DepthSnapshot
	err := json.Unmarshal(jsonData, &snapshot)

	require.NoError(t, err)
	require.Len(t, snapshot.Asks, 2)
	require.Len(t, snapshot.Bids, 1)
	require.Equal(t, int64(12345678), snapshot.Version)

	testutil.AssertDecimalEqual(t, snapshot.Asks[0].Price, "411.8")
	testutil.AssertDecimalEqual(t, snapshot.Asks[0].OrderCount, "10")
	testutil.AssertDecimalEqual(t, snapshot.Asks[0].Quantity, "1")

	testutil.AssertDecimalEqual(t, snapshot.Asks[1].Price, "412.5")
	testutil.AssertDecimalEqual(t, snapshot.Asks[1].OrderCount, "5")
	testutil.AssertDecimalEqual(t, snapshot.Asks[1].Quantity, "2")

	testutil.AssertDecimalEqual(t, snapshot.Bids[0].Price, "410.0")
	testutil.AssertDecimalEqual(t, snapshot.Bids[0].OrderCount, "8")
	testutil.AssertDecimalEqual(t, snapshot.Bids[0].Quantity, "1")
}

func TestDepthSnapshot_UnmarshalJSON_Invalid(t *testing.T) {
	t.Run("invalid asks", func(t *testing.T) {
		data := []byte(`{
			"asks": [
				["INVALID", 10, 1]
			],
			"bids": [
				[410.0, 8, 1]
			],
			"version": 12345678
		}`)

		var snapshot DepthSnapshot
		err := json.Unmarshal(data, &snapshot)

		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid asks")
	})

	t.Run("invalid bids", func(t *testing.T) {
		data := []byte(`{
			"asks": [
				[411.8, 10, 1]
			],
			"bids": [
				[410.0, 8]
			],
			"version": 12345678
		}`)

		var snapshot DepthSnapshot
		err := json.Unmarshal(data, &snapshot)

		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid bids")
	})
}

func Test_parseDepthLevels(t *testing.T) {
	t.Run("valid values", func(t *testing.T) {
		raw := [][]any{
			{411.8, float64(10), float64(1)},
		}

		levels, err := parseDepthLevels(raw)

		require.NoError(t, err)
		require.Len(t, levels, 1)

		testutil.AssertDecimalEqual(t, levels[0].Price, "411.8")
		testutil.AssertDecimalEqual(t, levels[0].OrderCount, "10")
		testutil.AssertDecimalEqual(t, levels[0].Quantity, "1")
	})

	t.Run("invalid depth level format", func(t *testing.T) {
		raw := [][]any{
			{411.8, float64(10)},
		}

		_, err := parseDepthLevels(raw)

		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid depth level format")
	})

	t.Run("invalid price", func(t *testing.T) {
		raw := [][]any{
			{"INVALID", float64(10), float64(1)},
		}

		_, err := parseDepthLevels(raw)

		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid price")
	})

	t.Run("invalid orderCount", func(t *testing.T) {
		raw := [][]any{
			{411.8, "INVALID", float64(1)},
		}

		_, err := parseDepthLevels(raw)

		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid order count")
	})

	t.Run("invalid quantity", func(t *testing.T) {
		raw := [][]any{
			{411.8, float64(10), "INVALID"},
		}

		_, err := parseDepthLevels(raw)

		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid quantity")
	})
}
