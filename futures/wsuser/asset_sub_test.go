package wsuser

import (
	"encoding/json"
	"testing"

	"github.com/IvanTurko/mexc-sdk-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewAssetSub(t *testing.T) {
	defer func() {
		r := recover()
		assert.Contains(t, r, "onData function is nil")
	}()

	NewAssetSub(nil)
}

func TestAssetSub_acceptEvent(t *testing.T) {
	sub := NewAssetSub(func(Asset) {}).(*assetSub)

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

func TestAssetSub_handleEvent(t *testing.T) {
	data := map[string]any{
		"availableBalance": 0.7514236,
		"currency":         "USDT",
		"frozenBalance":    0,
		"positionMargin":   0,
		"cashBalance":      0.7514236,
	}
	bytes, _ := json.Marshal(data)

	t.Run("handles valid data", func(t *testing.T) {
		var received Asset
		sub := NewAssetSub(func(d Asset) {
			received = d
		}).(*assetSub)

		msg := &message{
			Data: bytes,
			Ts:   1610005070083,
		}

		sub.handleEvent(msg)

		assert.Equal(t, "USDT", received.Currency)
		testutil.AssertDecimalEqual(t, received.PositionMargin, "0")
		testutil.AssertDecimalEqual(t, received.FrozenBalance, "0")
		testutil.AssertDecimalEqual(t, received.AvailableBalance, "0.7514236")
		testutil.AssertDecimalEqual(t, received.CashBalance, "0.7514236")
		assert.Equal(t, int64(1610005070083), *received.SendTime)
	})

	t.Run("handles invalid data", func(t *testing.T) {
		var calledErr error
		sub := NewAssetSub(func(Asset) {}).(*assetSub)
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
