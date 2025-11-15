package rest

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderSide_UnmarshalJSON(t *testing.T) {
	var side OrderSide

	t.Run("valid side", func(t *testing.T) {
		jsonStr := fmt.Sprintf(`"%s"`, OrderSideBuy)
		err := json.Unmarshal([]byte(jsonStr), &side)
		assert.NoError(t, err)
		assert.Equal(t, OrderSideBuy, side)
	})

	t.Run("invalid side", func(t *testing.T) {
		err := json.Unmarshal([]byte(`"INVALID"`), &side)
		assert.Error(t, err)
	})
}

func TestOrderType_UnmarshalJSON(t *testing.T) {
	var typ OrderType

	t.Run("valid type", func(t *testing.T) {
		jsonStr := fmt.Sprintf(`"%s"`, OrderTypeLimit)
		err := json.Unmarshal([]byte(jsonStr), &typ)
		assert.NoError(t, err)
		assert.Equal(t, OrderTypeLimit, typ)
	})

	t.Run("invalid type", func(t *testing.T) {
		err := json.Unmarshal([]byte(`"UNKNOWN"`), &typ)
		assert.Error(t, err)
	})
}

func TestOrderStatus_UnmarshalJSON(t *testing.T) {
	var status OrderStatus

	t.Run("valid status", func(t *testing.T) {
		jsonStr := fmt.Sprintf(`"%s"`, OrderStatusFilled)
		err := json.Unmarshal([]byte(jsonStr), &status)
		assert.NoError(t, err)
		assert.Equal(t, OrderStatusFilled, status)
	})

	t.Run("invalid status", func(t *testing.T) {
		err := json.Unmarshal([]byte(`"WRONG"`), &status)
		assert.Error(t, err)
	})
}

func TestKlineInterval_UnmarshalJSON(t *testing.T) {
	var interval KlineInterval

	t.Run("valid interval", func(t *testing.T) {
		jsonStr := fmt.Sprintf(`"%s"`, Interval1M)
		err := json.Unmarshal([]byte(jsonStr), &interval)
		assert.NoError(t, err)
		assert.Equal(t, Interval1M, interval)
	})

	t.Run("invalid interval", func(t *testing.T) {
		err := json.Unmarshal([]byte(`"99z"`), &interval)
		assert.Error(t, err)
	})
}

func TestChangeType_UnmarshalJSON(t *testing.T) {
	var ct ChangeType

	t.Run("valid type", func(t *testing.T) {
		jsonStr := fmt.Sprintf(`"%s"`, ChangeWithdraw)
		err := json.Unmarshal([]byte(jsonStr), &ct)
		assert.NoError(t, err)
		assert.Equal(t, ChangeWithdraw, ct)
	})

	t.Run("invalid type", func(t *testing.T) {
		err := json.Unmarshal([]byte(`"UnknownChange"`), &ct)
		assert.Error(t, err)
	})
}
