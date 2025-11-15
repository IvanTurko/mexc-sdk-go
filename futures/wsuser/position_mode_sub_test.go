package wsuser

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewPositionModeSub(t *testing.T) {
	defer func() {
		r := recover()
		assert.Contains(t, r, "onData function is nil")
	}()

	NewPositionModeSub(nil)
}

func TestPositionModeSub_acceptEvent(t *testing.T) {
	sub := NewPositionModeSub(func(PositionModeEvent) {}).(*positionModeSub)

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

func TestPositionModeSub_handleEvent(t *testing.T) {
	data := map[string]any{
		"positionMode": 1,
	}
	bytes, _ := json.Marshal(data)

	t.Run("handles valid data", func(t *testing.T) {
		var received PositionModeEvent
		sub := NewPositionModeSub(func(d PositionModeEvent) {
			received = d
		}).(*positionModeSub)

		msg := &message{
			Data: bytes,
			Ts:   1587442022003,
		}

		sub.handleEvent(msg)

		assert.Equal(t, PositionModeHedge, received.PositionMode)
		assert.Equal(t, int64(1587442022003), *received.SendTime)
	})

	t.Run("handles invalid data", func(t *testing.T) {
		var calledErr error
		sub := NewPositionModeSub(func(PositionModeEvent) {}).(*positionModeSub)
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
			"positionMode": 1
		}`)

		var mode PositionModeEvent
		err := json.Unmarshal(data, &mode)

		require.NoError(t, err)
		assert.Equal(t, PositionModeHedge, mode.PositionMode)
	})

	t.Run("invalid position mode", func(t *testing.T) {
		data := []byte(`{
			"positionMode": 9
		}`)

		var mode PositionModeEvent
		err := json.Unmarshal(data, &mode)

		require.Error(t, err)
		assert.ErrorContains(t, err, "unknown position mode code")
	})
}
