package wsmarket

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type bookTickerSub struct {
	streamName string
	onData     func(L1Ticker)
	onInvalid  func(error)
}

// NewBookTickerSub returns a subscription for aggregated L1 book tickers.
//
// symbol is the trading pair (e.g. "BTCUSDT").
// interval defines the update frequency (e.g. 100ms/1000ms).
// onData is invoked for each received L1Ticker.
//
// Panics:
//   - symbol is empty
//   - interval is invalid
//   - onData is nil
func NewBookTickerSub(
	symbol string,
	interval UpdateInterval,
	onData func(L1Ticker),
) Subscription {
	if symbol == "" {
		panic("NewBookTickerSub: invalid symbol name")
	}
	if !interval.isValid() {
		panic("NewBookTickerSub: invalid update interval: " + string(interval))
	}
	if onData == nil {
		panic("NewBookTickerSub: onData function is nil")
	}
	stream := fmt.Sprintf("spot@public.aggre.bookTicker.v3.api.pb@%s@%s", interval, symbol)

	return &bookTickerSub{
		streamName: stream,
		onData:     onData,
	}
}

// SetOnInvalid registers a callback that fires when the subscription becomes
// invalid (malformed message, server rejection, etc.).
func (b *bookTickerSub) SetOnInvalid(f func(error)) Subscription {
	b.onInvalid = f
	return b
}

func (b *bookTickerSub) matches(msg *message) (bool, error) {
	return msg.Msg == b.streamName, nil
}

func (b *bookTickerSub) acceptEvent(msg *PushDataV3MarketWrapper) bool {
	return msg.GetChannel() == b.streamName
}

func (b *bookTickerSub) handleEvent(msg *PushDataV3MarketWrapper) {
	v := msg.GetPublicAggreBookTicker()
	if v == nil || msg.Symbol == nil {
		return
	}

	res, err := mapProtoL1Ticker(v, *msg.Symbol, msg.SendTime)
	if err != nil {
		if b.onInvalid != nil {
			b.onInvalid(err)
		}
		return
	}

	b.onData(res)
}

func (b *bookTickerSub) id() string {
	return b.streamName
}

func (b *bookTickerSub) params() any {
	return b.streamName
}

// L1Ticker represents a level-1 book ticker with optional server send-time.
type L1Ticker struct {
	Symbol string
	BookTicker
	SendTime *int64
}

// BookTicker represents the best bid/ask with quantities.
type BookTicker struct {
	BidPrice    decimal.Decimal
	BidQuantity decimal.Decimal
	AskPrice    decimal.Decimal
	AskQuantity decimal.Decimal
}

func mapProtoL1Ticker(msg *PublicAggreBookTickerV3Api, symbol string, sendTime *int64) (L1Ticker, error) {
	bidPrice, err := decimal.NewFromString(msg.BidPrice)
	if err != nil {
		return L1Ticker{}, fmt.Errorf("invalid bidPrice %q", msg.BidPrice)
	}

	bidQty, err := decimal.NewFromString(msg.BidQuantity)
	if err != nil {
		return L1Ticker{}, fmt.Errorf("invalid bidQuantity %q", msg.BidQuantity)
	}

	askPrice, err := decimal.NewFromString(msg.AskPrice)
	if err != nil {
		return L1Ticker{}, fmt.Errorf("invalid askPrice %q", msg.AskPrice)
	}

	askQty, err := decimal.NewFromString(msg.AskQuantity)
	if err != nil {
		return L1Ticker{}, fmt.Errorf("invalid askQuantity %q", msg.AskQuantity)
	}

	return L1Ticker{
		Symbol: symbol,
		BookTicker: BookTicker{
			BidPrice:    bidPrice,
			BidQuantity: bidQty,
			AskPrice:    askPrice,
			AskQuantity: askQty,
		},
		SendTime: sendTime,
	}, nil
}
