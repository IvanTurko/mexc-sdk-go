package wsmarket

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

// KlineInterval represents the server-defined candlestick interval
type KlineInterval string

const (
	Kline1Min   KlineInterval = "Min1"
	Kline5Min   KlineInterval = "Min5"
	Kline15Min  KlineInterval = "Min15"
	Kline30Min  KlineInterval = "Min30"
	Kline60Min  KlineInterval = "Min60"
	Kline4Hour  KlineInterval = "Hour4"
	Kline8Hour  KlineInterval = "Hour8"
	Kline1Day   KlineInterval = "Day1"
	Kline1Week  KlineInterval = "Week1"
	Kline1Month KlineInterval = "Month1"
)

func (k KlineInterval) isValid() bool {
	switch k {
	case Kline1Min, Kline5Min, Kline15Min, Kline30Min, Kline60Min,
		Kline4Hour, Kline8Hour, Kline1Day, Kline1Week, Kline1Month:
		return true
	default:
		return false
	}
}

type klineSub struct {
	symbol    string
	interval  KlineInterval
	onInvalid func(error)
	onData    func(Kline)
}

// NewKlineSub creates a subscription for candlestick (kline) updates.
//
// symbol is the trading pair, e.g. "BTC_USDT".
// interval is the kline interval (e.g. Min1, Min5, Hour4).
// onData is called for every received Kline update.
//
// Panics:
//   - symbol is empty
//   - interval is invalid
//   - onData is nil
func NewKlineSub(
	symbol string,
	interval KlineInterval,
	onData func(Kline),
) Subscription {
	if symbol == "" {
		panic("NewKlineSub: invalid symbol name")
	}
	if !interval.isValid() {
		panic("NewKlineSub: invalid interval")
	}
	if onData == nil {
		panic("NewKlineSub: onData function is nil")
	}

	return &klineSub{
		symbol:   symbol,
		interval: interval,
		onData:   onData,
	}
}

// SetOnInvalid registers a callback that is triggered if the subscription
// becomes invalid (invalid payload, server rejection, etc.).
func (k *klineSub) SetOnInvalid(f func(error)) Subscription {
	k.onInvalid = f
	return k
}

func (k *klineSub) matches(msg *message) (bool, error) {
	if msg.Channel == "rs.sub."+k.channel() {
		var r string
		if err := json.Unmarshal(msg.Data, &r); err != nil {
			return false, fmt.Errorf("invalid success payload: %s", string(msg.Data))
		}
		if r == "success" {
			return true, nil
		}
	}

	if msg.Channel == "rs.error" {
		return true, fmt.Errorf("sub failed: %s", string(msg.Data))
	}

	return false, nil
}

func (k *klineSub) acceptEvent(msg *message) bool {
	return msg.Channel == "push."+k.channel() && msg.Symbol == k.symbol
}

func (k *klineSub) handleEvent(msg *message) {
	var kline Kline

	if err := json.Unmarshal(msg.Data, &kline); err != nil {
		err = fmt.Errorf("failed to unmarshal data: %v, raw: %s", err, string(msg.Data))
		if k.onInvalid != nil {
			k.onInvalid(err)
		}
		return
	}

	kline.SendTime = &msg.Ts
	k.onData(kline)
}

func (k *klineSub) id() string {
	return fmt.Sprintf("%s@%s", k.channel(), k.symbol)
}

func (k *klineSub) channel() string {
	return "kline"
}

func (k *klineSub) payload(op subscriptionOp) any {
	return wsRequestPayload{
		Method: fmt.Sprintf("%s.%s", op, k.channel()),
		Param: map[string]any{
			"symbol":   k.symbol,
			"interval": k.interval,
		},
	}
}

// Kline represents a single candlestick snapshot.
type Kline struct {
	Symbol    string          `json:"symbol"`
	Amount    decimal.Decimal `json:"a"`
	Close     decimal.Decimal `json:"c"`
	High      decimal.Decimal `json:"h"`
	Interval  KlineInterval   `json:"interval"`
	Low       decimal.Decimal `json:"l"`
	Open      decimal.Decimal `json:"o"`
	Quantity  decimal.Decimal `json:"q"`
	Timestamp int64           `json:"t"`
	SendTime  *int64
}
