package wsmarket

import (
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
	streamName string
	onData     func(Kline)
	onInvalid  func(error)
}

// NewKlineSub creates a subscription for candlestick (kline) updates.
//
// symbol is the trading pair, e.g. "BTCUSDT".
// interval is the kline interval (e.g. 1m, 5m, 1h).
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
		panic("NewKlineSub: invalid kline interval: " + string(interval))
	}
	if onData == nil {
		panic("NewKlineSub: onData function is nil")
	}
	stream := fmt.Sprintf("spot@public.kline.v3.api.pb@%s@%s", symbol, interval)

	return &klineSub{
		streamName: stream,
		onData:     onData,
	}
}

// SetOnInvalid registers a callback that is triggered if the subscription
// becomes invalid (invalid payload, server rejection, etc.).
func (k *klineSub) SetOnInvalid(f func(error)) Subscription {
	k.onInvalid = f
	return k
}

func (k *klineSub) matches(msg *message) (bool, error) {
	return msg.Msg == k.streamName, nil
}

func (k *klineSub) acceptEvent(msg *PushDataV3MarketWrapper) bool {
	return msg.GetChannel() == k.streamName
}

func (k *klineSub) handleEvent(msg *PushDataV3MarketWrapper) {
	v := msg.GetPublicSpotKline()
	if v == nil || msg.Symbol == nil {
		return
	}

	res, err := mapProtoKline(v, *msg.Symbol, msg.SendTime)
	if err != nil {
		if k.onInvalid != nil {
			k.onInvalid(err)
		}
		return
	}

	k.onData(res)
}

func (k *klineSub) id() string {
	return k.streamName
}

func (k *klineSub) params() any {
	return k.streamName
}

// Kline represents a single candlestick snapshot.
type Kline struct {
	Symbol       string
	Interval     KlineInterval
	WindowStart  int64
	OpeningPrice decimal.Decimal
	ClosingPrice decimal.Decimal
	HighestPrice decimal.Decimal
	LowestPrice  decimal.Decimal
	Volume       decimal.Decimal
	Amount       decimal.Decimal
	WindowEnd    int64
	SendTime     *int64
}

func mapProtoKline(msg *PublicSpotKlineV3Api, symbol string, sendTime *int64) (Kline, error) {
	interval := KlineInterval(msg.Interval)
	if !interval.isValid() {
		return Kline{}, fmt.Errorf("invalid interval %q", msg.Interval)
	}

	open, err := decimal.NewFromString(msg.OpeningPrice)
	if err != nil {
		return Kline{}, fmt.Errorf("invalid openingPrice %q", msg.OpeningPrice)
	}

	closePrice, err := decimal.NewFromString(msg.ClosingPrice)
	if err != nil {
		return Kline{}, fmt.Errorf("invalid closingPrice %q", msg.ClosingPrice)
	}

	high, err := decimal.NewFromString(msg.HighestPrice)
	if err != nil {
		return Kline{}, fmt.Errorf("invalid highestPrice %q", msg.HighestPrice)
	}

	low, err := decimal.NewFromString(msg.LowestPrice)
	if err != nil {
		return Kline{}, fmt.Errorf("invalid lowestPrice %q", msg.LowestPrice)
	}

	vol, err := decimal.NewFromString(msg.Volume)
	if err != nil {
		return Kline{}, fmt.Errorf("invalid volume %q", msg.Volume)
	}

	amt, err := decimal.NewFromString(msg.Amount)
	if err != nil {
		return Kline{}, fmt.Errorf("invalid amount %q", msg.Amount)
	}

	return Kline{
		Symbol:       symbol,
		Interval:     interval,
		WindowStart:  msg.WindowStart,
		OpeningPrice: open,
		ClosingPrice: closePrice,
		HighestPrice: high,
		LowestPrice:  low,
		Volume:       vol,
		Amount:       amt,
		WindowEnd:    msg.WindowEnd,
		SendTime:     sendTime,
	}, nil
}
