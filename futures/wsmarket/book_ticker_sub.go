package wsmarket

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

type bookTickerSub struct {
	symbol    string
	onInvalid func(error)
	onData    func(*BookTicker)
}

// NewBookTickerSub creates a new subscription for book tickers.
//
// symbol is the trading pair, e.g., "BTC_USDT".
// onData is the callback function to handle the book ticker.
//
// Panics:
//   - symbol is empty
//   - onData is nil
func NewBookTickerSub(
	symbol string,
	onData func(*BookTicker),
) Subscription {
	if symbol == "" {
		panic("NewBookTickerSub: invalid symbol name")
	}
	if onData == nil {
		panic("NewBookTickerSub: onData function is nil")
	}

	return &bookTickerSub{
		symbol: symbol,
		onData: onData,
	}
}

// SetOnInvalid sets a callback invoked when the subscription becomes invalid
// (e.g. malformed message, server rejection).
func (b *bookTickerSub) SetOnInvalid(f func(error)) Subscription {
	b.onInvalid = f
	return b
}

func (b *bookTickerSub) matches(msg *message) (bool, error) {
	if msg.Channel == "rs.sub."+b.channel() {
		var s string
		if err := json.Unmarshal(msg.Data, &s); err != nil {
			return false, fmt.Errorf("invalid success payload: %s", string(msg.Data))
		}
		if s == "success" {
			return true, nil
		}
	}

	if msg.Channel == "rs.error" {
		return true, fmt.Errorf("sub failed: %s", string(msg.Data))
	}

	return false, nil
}

func (b *bookTickerSub) acceptEvent(msg *message) bool {
	return msg.Channel == "push."+b.channel() && msg.Symbol == b.symbol
}

func (b *bookTickerSub) handleEvent(msg *message) {
	var ticker *BookTicker

	if err := json.Unmarshal(msg.Data, &ticker); err != nil {
		err = fmt.Errorf("failed to unmarshal data: %v, raw: %s", err, string(msg.Data))
		if b.onInvalid != nil {
			b.onInvalid(err)
		}
		return
	}

	ticker.SendTime = &msg.Ts
	b.onData(ticker)
}

func (b *bookTickerSub) id() string {
	return fmt.Sprintf("%s@%s", b.channel(), b.symbol)
}

func (b *bookTickerSub) channel() string {
	return "ticker"
}

func (b *bookTickerSub) payload(op subscriptionOp) any {
	return wsRequestPayload{
		Method: fmt.Sprintf("%s.%s", op, b.channel()),
		Param: map[string]any{
			"symbol": b.symbol,
		},
	}
}

// BookTicker represents a book ticker.
type BookTicker struct {
	Symbol        string          `json:"symbol"`
	LastPrice     decimal.Decimal `json:"lastPrice"`
	Bid1          decimal.Decimal `json:"bid1"`
	Ask1          decimal.Decimal `json:"ask1"`
	Volume24      decimal.Decimal `json:"volume24"`
	HoldVol       decimal.Decimal `json:"holdVol"`
	Lower24Price  decimal.Decimal `json:"lower24Price"`
	High24Price   decimal.Decimal `json:"high24Price"`
	RiseFallRate  decimal.Decimal `json:"riseFallRate"`
	RiseFallValue decimal.Decimal `json:"riseFallValue"`
	IndexPrice    decimal.Decimal `json:"indexPrice"`
	FairPrice     decimal.Decimal `json:"fairPrice"`
	FundingRate   decimal.Decimal `json:"fundingRate"`
	Timestamp     int64           `json:"timestamp"`
	SendTime      *int64
}
