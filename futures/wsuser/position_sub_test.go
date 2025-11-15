package wsuser

import (
	"encoding/json"
	"testing"

	"github.com/IvanTurko/mexc-sdk-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewPositionSub(t *testing.T) {
	defer func() {
		r := recover()
		assert.Contains(t, r, "onData function is nil")
	}()

	NewPositionSub(nil)
}

func TestPositionSub_acceptEvent(t *testing.T) {
	sub := NewPositionSub(func(*PositionEvent) {}).(*positionSub)

	t.Run("accepts valid push event", func(t *testing.T) {
		msg := &message{
			Channel: "push.personal." + sub.channel(),
		}
		assert.True(t, sub.acceptEvent(msg))
	})

	t.Run("rejects invalid channel", func(t *testing.T) {
		msg := &message{
			Channel: "push.personal.invalid",
		}
		assert.False(t, sub.acceptEvent(msg))
	})
}

func TestPositionSub_handleEvent(t *testing.T) {
	symbol := "BTC_USDT"
	data := map[string]any{
		"positionId":     1397818,
		"symbol":         symbol,
		"holdVol":        0,
		"positionType":   1,
		"openType":       1,
		"state":          3,
		"frozenVol":      0,
		"closeVol":       1,
		"holdAvgPrice":   0.736,
		"closeAvgPrice":  0.731,
		"openAvgPrice":   0.736,
		"liquidatePrice": 0,
		"oim":            0,
		"adlLevel":       0,
		"im":             0,
		"holdFee":        0,
		"realised":       -0.0005,
		"autoAddIm":      false,
		"leverage":       15,
	}
	bytes, _ := json.Marshal(data)

	t.Run("handles valid data", func(t *testing.T) {
		var received *PositionEvent
		sub := NewPositionSub(func(d *PositionEvent) {
			received = d
		}).(*positionSub)

		msg := &message{
			Data: bytes,
			Ts:   1610005070157,
		}

		sub.handleEvent(msg)
		require.NotNil(t, received)

		assert.Equal(t, int64(1397818), received.PositionId)
		assert.Equal(t, symbol, received.Symbol)
		testutil.AssertDecimalEqual(t, received.HoldVol, "0")
		assert.Equal(t, PositionTypeLong, received.PositionType)
		assert.Equal(t, OpenTypeIsolated, received.OpenType)
		assert.Equal(t, PositionStateClosed, received.State)
		testutil.AssertDecimalEqual(t, received.FrozenVol, "0")
		testutil.AssertDecimalEqual(t, received.CloseVol, "1")
		testutil.AssertDecimalEqual(t, received.HoldAvgPrice, "0.736")
		testutil.AssertDecimalEqual(t, received.CloseAvgPrice, "0.731")
		testutil.AssertDecimalEqual(t, received.OpenAvgPrice, "0.736")
		testutil.AssertDecimalEqual(t, received.LiquidatePrice, "0")
		testutil.AssertDecimalEqual(t, received.Oim, "0")
		assert.Equal(t, 0, received.AdlLevel)
		testutil.AssertDecimalEqual(t, received.Im, "0")
		testutil.AssertDecimalEqual(t, received.HoldFee, "0")
		testutil.AssertDecimalEqual(t, received.Realised, "-0.0005")
		assert.False(t, received.AutoAddIm)
		assert.Equal(t, 15, received.Leverage)
		assert.Equal(t, int64(1610005070157), *received.SendTime)
	})

	t.Run("handles invalid data", func(t *testing.T) {
		var calledErr error
		sub := NewPositionSub(func(*PositionEvent) {}).(*positionSub)
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

func TestPositionEvent_UnmarshalJSON(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		data := []byte(`{
			"positionId": 1397818,
			"symbol": "BTC_USDT",
			"holdVol": 0,
			"positionType": 1,
			"openType": 1,
			"state": 3,
			"frozenVol": 0,
			"closeVol": 1,
			"holdAvgPrice": 0.736,
			"closeAvgPrice": 0.731,
			"openAvgPrice": 0.736,
			"liquidatePrice": 0,
			"oim": 0,
			"adlLevel": 0,
			"im": 0,
			"holdFee": 0,
			"realised": -0.0005,
			"autoAddIm": false,
			"leverage": 15
		}`)

		var event PositionEvent
		err := json.Unmarshal(data, &event)

		require.NoError(t, err)
		assert.Equal(t, int64(1397818), event.PositionId)
		assert.Equal(t, "BTC_USDT", event.Symbol)
		testutil.AssertDecimalEqual(t, event.HoldVol, "0")
		assert.Equal(t, PositionTypeLong, event.PositionType)
		assert.Equal(t, OpenTypeIsolated, event.OpenType)
		assert.Equal(t, PositionStateClosed, event.State)
		testutil.AssertDecimalEqual(t, event.FrozenVol, "0")
		testutil.AssertDecimalEqual(t, event.CloseVol, "1")
		testutil.AssertDecimalEqual(t, event.HoldAvgPrice, "0.736")
		testutil.AssertDecimalEqual(t, event.CloseAvgPrice, "0.731")
		testutil.AssertDecimalEqual(t, event.OpenAvgPrice, "0.736")
		testutil.AssertDecimalEqual(t, event.LiquidatePrice, "0")
		testutil.AssertDecimalEqual(t, event.Oim, "0")
		assert.Equal(t, 0, event.AdlLevel)
		testutil.AssertDecimalEqual(t, event.Im, "0")
		testutil.AssertDecimalEqual(t, event.HoldFee, "0")
		testutil.AssertDecimalEqual(t, event.Realised, "-0.0005")
		assert.False(t, event.AutoAddIm)
		assert.Equal(t, 15, event.Leverage)
	})

	t.Run("invalid position type", func(t *testing.T) {
		data := []byte(`{"positionType": 3}`)

		var event PositionEvent
		err := json.Unmarshal(data, &event)

		require.Error(t, err)
		assert.ErrorContains(t, err, "unknown position type code")
	})

	t.Run("invalid open type", func(t *testing.T) {
		data := []byte(`{"positionType": 1, "openType": 3}`)

		var event PositionEvent
		err := json.Unmarshal(data, &event)

		require.Error(t, err)
		assert.ErrorContains(t, err, "unknown open type code")
	})

	t.Run("invalid state", func(t *testing.T) {
		data := []byte(`{"positionType": 1, "openType": 1, "state": 4}`)

		var event PositionEvent
		err := json.Unmarshal(data, &event)

		require.Error(t, err)
		assert.ErrorContains(t, err, "unknown position state code")
	})
}
