package wsmarket

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

type bookTickerBatchSub struct {
	onInvalid func(error)
	onData    func(*BookTickerBatch)
}

// NewBookTickerBatchSub creates a new subscription for book ticker batches.
//
// onData is the callback function to handle the book ticker batch.
//
// Panics:
//   - onData is nil
func NewBookTickerBatchSub(
	onData func(*BookTickerBatch),
) Subscription {
	if onData == nil {
		panic("NewBookTickerBatchSub: onData function is nil")
	}

	return &bookTickerBatchSub{
		onData: onData,
	}
}

// SetOnInvalid sets a callback invoked when the subscription becomes invalid
// (e.g. malformed message, server rejection).
func (b *bookTickerBatchSub) SetOnInvalid(f func(error)) Subscription {
	b.onInvalid = f
	return b
}

func (b *bookTickerBatchSub) matches(msg *message) (bool, error) {
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

func (b *bookTickerBatchSub) acceptEvent(msg *message) bool {
	return msg.Channel == "push."+b.channel()
}

func (b *bookTickerBatchSub) handleEvent(msg *message) {
	var tickers []BatchTicker

	if err := json.Unmarshal(msg.Data, &tickers); err != nil {
		err = fmt.Errorf("failed to unmarshal data: %v, raw: %s", err, string(msg.Data))
		if b.onInvalid != nil {
			b.onInvalid(err)
		}
		return
	}

	batch := &BookTickerBatch{
		Tickers:  tickers,
		SendTime: &msg.Ts,
	}
	b.onData(batch)
}

func (b *bookTickerBatchSub) id() string {
	return fmt.Sprintf("%s", b.channel())
}

func (b *bookTickerBatchSub) channel() string {
	return "tickers"
}

func (b *bookTickerBatchSub) payload(op subscriptionOp) any {
	return wsRequestPayload{
		Method: fmt.Sprintf("%s.%s", op, b.channel()),
		Gzip:   true,
	}
}

// BookTickerBatch represents a batch of book tickers.
type BookTickerBatch struct {
	Tickers  []BatchTicker
	SendTime *int64
}

// BatchTicker represents a single ticker in a batch.
type BatchTicker struct {
	Symbol       string          `json:"symbol"`
	LastPrice    decimal.Decimal `json:"lastPrice"`
	Volume24     decimal.Decimal `json:"volume24"`
	RiseFallRate decimal.Decimal `json:"riseFallRate"`
	FairPrice    decimal.Decimal `json:"fairPrice"`
}
