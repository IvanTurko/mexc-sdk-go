package wsmarket

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

// TradeSide indicates whether the trade was buyer or seller initiated.
type TradeSide string

const (
	TradeSideBuy  TradeSide = "BUY"
	TradeSideSell TradeSide = "SELL"
)

type tradeStreamsSub struct {
	streamName string
	onData     func([]Trade)
	onInvalid  func(error)
}

// NewTradeStreamsSub creates a subscription for aggregated trade streams.
//
// symbol is the trading pair, e.g. "BTCUSDT".
// interval defines the update frequency.
// onData is called for each batch of trades.
//
// Panics:
//   - symbol is empty
//   - interval is invalid
//   - onData is nil
func NewTradeStreamsSub(
	symbol string,
	interval UpdateInterval,
	onData func([]Trade),
) Subscription {
	if symbol == "" {
		panic("NewTradeStreamsSub: invalid symbol name")
	}
	if !interval.isValid() {
		panic("NewTradeStreamsSub: invalid update interval: " + string(interval))
	}
	if onData == nil {
		panic("NewTradeStreamsSub: onData function is nil")
	}
	stream := fmt.Sprintf("spot@public.aggre.deals.v3.api.pb@%s@%s", interval, symbol)

	return &tradeStreamsSub{
		streamName: stream,
		onData:     onData,
	}
}

// SetOnInvalid registers a callback triggered when the subscription becomes
// invalid (malformed message, server error, rejection, etc.).
func (t *tradeStreamsSub) SetOnInvalid(f func(error)) Subscription {
	t.onInvalid = f
	return t
}

func (t *tradeStreamsSub) matches(msg *message) (bool, error) {
	return msg.Msg == t.streamName, nil
}

func (t *tradeStreamsSub) acceptEvent(msg *PushDataV3MarketWrapper) bool {
	return msg.GetChannel() == t.streamName
}

func (t *tradeStreamsSub) handleEvent(msg *PushDataV3MarketWrapper) {
	v := msg.GetPublicAggreDeals()
	if v == nil || msg.Symbol == nil {
		return
	}

	res, err := mapProtoTrades(v, *msg.Symbol, msg.SendTime)
	if err != nil {
		if t.onInvalid != nil {
			t.onInvalid(err)
		}
		return
	}

	if len(res) == 0 {
		if t.onInvalid != nil {
			t.onInvalid(errors.New("trades not found"))
		}
		return
	}

	t.onData(res)
}

func (t *tradeStreamsSub) id() string {
	return t.streamName
}

func (t *tradeStreamsSub) params() any {
	return t.streamName
}

// Trade represents a single executed trade.
type Trade struct {
	Symbol   string
	Price    decimal.Decimal
	Quantity decimal.Decimal
	Side     TradeSide
	Time     int64
	SendTime *int64
}

func mapProtoTrades(msg *PublicAggreDealsV3Api, symbol string, sendTime *int64) ([]Trade, error) {
	deals := msg.Deals
	trades := make([]Trade, 0, len(deals))
	for _, item := range deals {
		if nil == item {
			return nil, errors.New("trade is nil")
		}

		trade, err := mapProtoTrade(item, symbol, sendTime)
		if err != nil {
			return nil, fmt.Errorf("invalid trade: %v", err)
		}

		trades = append(trades, trade)
	}

	return trades, nil
}

func mapProtoTrade(item *PublicAggreDealsV3ApiItem, symbol string, sendTime *int64) (Trade, error) {
	price, err := decimal.NewFromString(item.Price)
	if err != nil {
		return Trade{}, fmt.Errorf("invalid price %q", item.Price)
	}

	quantity, err := decimal.NewFromString(item.Quantity)
	if err != nil {
		return Trade{}, fmt.Errorf("invalid quantity %q", item.Quantity)
	}

	side, err := parseTradeSide(item.TradeType)
	if err != nil {
		return Trade{}, err
	}

	return Trade{
		Symbol:   symbol,
		Price:    price,
		Quantity: quantity,
		Side:     side,
		Time:     item.Time,
		SendTime: sendTime,
	}, nil
}

func parseTradeSide(code int32) (TradeSide, error) {
	switch code {
	case 1:
		return TradeSideBuy, nil
	case 2:
		return TradeSideSell, nil
	default:
		return "", fmt.Errorf("unknown trade side code: %d", code)
	}
}
