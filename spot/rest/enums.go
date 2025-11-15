package rest

import (
	"encoding/json"
	"fmt"
)

// OrderSide represents the direction of an order.
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

// OrderType represents the type of an order.
type OrderType string

const (
	OrderTypeLimit             OrderType = "LIMIT"
	OrderTypeMarket            OrderType = "MARKET"
	OrderTypeLimitMaker        OrderType = "LIMIT_MAKER"
	OrderTypeImmediateOrCancel OrderType = "IMMEDIATE_OR_CANCEL"
	OrderTypeFillOrKill        OrderType = "FILL_OR_KILL"
)

// OrderStatus represents the current status of an order.
type OrderStatus string

const (
	OrderStatusNew               OrderStatus = "NEW"
	OrderStatusFilled            OrderStatus = "FILLED"
	OrderStatusPartiallyFilled   OrderStatus = "PARTIALLY_FILLED"
	OrderStatusCanceled          OrderStatus = "CANCELED"
	OrderStatusPartiallyCanceled OrderStatus = "PARTIALLY_CANCELED"
)

// KlineInterval represents candlestick intervals.
type KlineInterval string

const (
	Interval1m  KlineInterval = "1m"
	Interval5m  KlineInterval = "5m"
	Interval15m KlineInterval = "15m"
	Interval30m KlineInterval = "30m"
	Interval1h  KlineInterval = "60m"
	Interval4h  KlineInterval = "4h"
	Interval1d  KlineInterval = "1d"
	Interval1w  KlineInterval = "1W"
	Interval1M  KlineInterval = "1M"
)

// ChangeType represents various account activity types.
type ChangeType string

const (
	ChangeWithdraw        ChangeType = "WITHDRAW"
	ChangeWithdrawFee     ChangeType = "WITHDRAW_FEE"
	ChangeDeposit         ChangeType = "DEPOSIT"
	ChangeDepositFee      ChangeType = "DEPOSIT_FEE"
	ChangeEntrust         ChangeType = "ENTRUST"
	ChangeEntrustPlace    ChangeType = "ENTRUST_PLACE"
	ChangeEntrustCancel   ChangeType = "ENTRUST_CANCEL"
	ChangeTradeFee        ChangeType = "TRADE_FEE"
	ChangeEntrustUnfrozen ChangeType = "ENTRUST_UNFROZEN"
	ChangeSugar           ChangeType = "SUGAR"
	ChangeETFIndex        ChangeType = "ETF_INDEX"
)

func (s OrderSide) isValid() bool {
	switch s {
	case OrderSideBuy, OrderSideSell:
		return true
	default:
		return false
	}
}

func (t OrderType) isValid() bool {
	switch t {
	case OrderTypeLimit,
		OrderTypeMarket,
		OrderTypeLimitMaker,
		OrderTypeImmediateOrCancel,
		OrderTypeFillOrKill:
		return true
	default:
		return false
	}
}

func (s OrderStatus) isValid() bool {
	switch s {
	case OrderStatusNew,
		OrderStatusFilled,
		OrderStatusPartiallyFilled,
		OrderStatusCanceled,
		OrderStatusPartiallyCanceled:
		return true
	default:
		return false
	}
}

func (k KlineInterval) isValid() bool {
	switch k {
	case Interval1m, Interval5m, Interval15m, Interval30m,
		Interval1h, Interval4h, Interval1d, Interval1w, Interval1M:
		return true
	default:
		return false
	}
}

func (c ChangeType) isValid() bool {
	switch c {
	case ChangeWithdraw,
		ChangeWithdrawFee,
		ChangeDeposit,
		ChangeDepositFee,
		ChangeEntrust,
		ChangeEntrustPlace,
		ChangeEntrustCancel,
		ChangeTradeFee,
		ChangeEntrustUnfrozen,
		ChangeSugar,
		ChangeETFIndex:
		return true
	default:
		return false
	}
}

func (o *OrderSide) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsed := OrderSide(s)
	if !parsed.isValid() {
		return fmt.Errorf("invalid order side: %s", s)
	}

	*o = parsed
	return nil
}

func (o *OrderType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsed := OrderType(s)
	if !parsed.isValid() {
		return fmt.Errorf("invalid order type: %s", s)
	}

	*o = parsed
	return nil
}

func (o *OrderStatus) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsed := OrderStatus(s)
	if !parsed.isValid() {
		return fmt.Errorf("invalid order status: %s", s)
	}

	*o = parsed
	return nil
}

func (k *KlineInterval) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsed := KlineInterval(s)
	if !parsed.isValid() {
		return fmt.Errorf("invalid kline interval: %s", s)
	}

	*k = parsed
	return nil
}

func (c *ChangeType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsed := ChangeType(s)
	if !parsed.isValid() {
		return fmt.Errorf("invalid change type: %s", s)
	}

	*c = parsed
	return nil
}
