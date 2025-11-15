package wsuser

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewAdlLevelSub(t *testing.T) {
	defer func() {
		r := recover()
		assert.Contains(t, r, "onData function is nil")
	}()

	NewAdlLevelSub(nil)
}

func TestAdlLevelSub_acceptEvent(t *testing.T) {
	sub := NewAdlLevelSub(func(AdlLevelEvent) {}).(*AdlLevelSub)

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

func TestAdlLevelSub_handleEvent(t *testing.T) {
	data := map[string]any{
		"adlLevel":   1,
		"positionId": 1397818,
	}
	bytes, _ := json.Marshal(data)

	t.Run("handles valid data", func(t *testing.T) {
		var received AdlLevelEvent
		sub := NewAdlLevelSub(func(d AdlLevelEvent) {
			received = d
		}).(*AdlLevelSub)

		msg := &message{
			Data: bytes,
			Ts:   1610005032231,
		}

		sub.handleEvent(msg)

		assert.Equal(t, AdlLevel1, received.AdlLevel)
		assert.Equal(t, int64(1397818), received.PositionId)
		assert.Equal(t, int64(1610005032231), *received.SendTime)
	})

	t.Run("handles invalid data", func(t *testing.T) {
		var calledErr error
		sub := NewAdlLevelSub(func(AdlLevelEvent) {}).(*AdlLevelSub)
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

func TestAdlLevel_UnmarshalJSON(t *testing.T) {
	t.Run("valid level", func(t *testing.T) {
		for i := 0; i <= 5; i++ {
			var l AdlLevel
			err := json.Unmarshal(fmt.Appendf([]byte{}, "%d", i), &l)
			require.NoError(t, err)
			assert.Equal(t, AdlLevel(i), l)
		}
	})

	t.Run("invalid level", func(t *testing.T) {
		var l AdlLevel
		err := json.Unmarshal([]byte("6"), &l)
		assert.Error(t, err)
	})
}
