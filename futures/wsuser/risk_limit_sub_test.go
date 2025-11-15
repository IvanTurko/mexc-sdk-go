package wsuser

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewRiskLimitSub(t *testing.T) {
	defer func() {
		r := recover()
		assert.Contains(t, r, "onData function is nil")
	}()

	NewRiskLimitSub(nil)
}

func TestRiskLimitSub_acceptEvent(t *testing.T) {
	sub := NewRiskLimitSub(func(RiskLimitEvent) {}).(*riskLimitSub)

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

func TestRiskLimitSub_handleEvent(t *testing.T) {
	data := map[string]any{
		"symbol":       "BTC_USDT",
		"positionType": 1,
		"riskSource":   1,
		"level":        1,
		"maxVol":       "100",
		"maxLeverage":  100,
		"mmr":          "0.005",
		"imr":          "0.01",
	}
	bytes, _ := json.Marshal(data)

	t.Run("handles valid data", func(t *testing.T) {
		var received RiskLimitEvent
		sub := NewRiskLimitSub(func(d RiskLimitEvent) {
			received = d
		}).(*riskLimitSub)

		msg := &message{
			Data: bytes,
			Ts:   1610005032231,
		}

		sub.handleEvent(msg)

		assert.Equal(t, "BTC_USDT", received.Symbol)
		assert.Equal(t, PositionTypeLong, received.PositionType)
		assert.Equal(t, RiskSourceLiquidationService, received.RiskSource)
		assert.Equal(t, 1, received.Level)
		assert.Equal(t, decimal.NewFromInt(100), received.MaxVol)
		assert.Equal(t, 100, received.MaxLeverage)
		assert.Equal(t, decimal.NewFromFloat(0.005), received.Mmr)
		assert.Equal(t, decimal.NewFromFloat(0.01), received.Imr)
		assert.Equal(t, int64(1610005032231), *received.SendTime)
	})

	t.Run("handles invalid data", func(t *testing.T) {
		var calledErr error
		sub := NewRiskLimitSub(func(RiskLimitEvent) {}).(*riskLimitSub)
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

func TestRiskLimitEvent_UnmarshalJSON(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		data := []byte(`{
			"symbol": "BTC_USDT",
			"positionType": 1,
			"riskSource": 1,
			"level": 1,
			"maxVol": "100",
			"maxLeverage": 100,
			"mmr": "0.005",
			"imr": "0.01"
		}`)

		var event RiskLimitEvent
		err := json.Unmarshal(data, &event)

		require.NoError(t, err)
		assert.Equal(t, "BTC_USDT", event.Symbol)
		assert.Equal(t, PositionTypeLong, event.PositionType)
		assert.Equal(t, RiskSourceLiquidationService, event.RiskSource)
		assert.Equal(t, 1, event.Level)
		assert.Equal(t, decimal.NewFromInt(100), event.MaxVol)
		assert.Equal(t, 100, event.MaxLeverage)
		assert.Equal(t, decimal.NewFromFloat(0.005), event.Mmr)
		assert.Equal(t, decimal.NewFromFloat(0.01), event.Imr)
	})

	t.Run("invalid position type", func(t *testing.T) {
		data := []byte(`{"positionType": 3}`)

		var event RiskLimitEvent
		err := json.Unmarshal(data, &event)

		require.Error(t, err)
		assert.ErrorContains(t, err, "unknown position type code")
	})

	t.Run("invalid risk source", func(t *testing.T) {
		data := []byte(`{"positionType": 1, "riskSource": 2}`)

		var event RiskLimitEvent
		err := json.Unmarshal(data, &event)

		require.Error(t, err)
		assert.ErrorContains(t, err, "unknown risk source code")
	})
}
