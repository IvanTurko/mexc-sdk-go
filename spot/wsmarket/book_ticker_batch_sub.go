package wsmarket

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

type bookTickerBatchSub struct {
	streamName string
	onData     func(*BookTickerBatch)
	onInvalid  func(error)
}

// NewBookTickerBatchSub returns a subscription for batched L1 book tickers.
//
// symbol is the trading pair (e.g. "BTCUSDT").
// onData is called for every received BookTickerBatch.
//
// Panics:
//   - symbol is empty
//   - onData is nil
func NewBookTickerBatchSub(
	symbol string,
	onData func(*BookTickerBatch),
) Subscription {
	if symbol == "" {
		panic("NewBookTickerBatchSub: invalid symbol name")
	}
	if onData == nil {
		panic("NewBookTickerBatchSub: onData function is nil")
	}
	stream := fmt.Sprintf("spot@public.bookTicker.batch.v3.api.pb@%s", symbol)

	return &bookTickerBatchSub{
		streamName: stream,
		onData:     onData,
	}
}

// SetOnInvalid sets a callback invoked when the subscription becomes invalid
// (e.g. malformed message, server rejection).
func (b *bookTickerBatchSub) SetOnInvalid(f func(error)) Subscription {
	b.onInvalid = f
	return b
}

func (b *bookTickerBatchSub) matches(msg *message) (bool, error) {
	return msg.Msg == b.streamName, nil
}

func (b *bookTickerBatchSub) acceptEvent(msg *PushDataV3MarketWrapper) bool {
	return msg.GetChannel() == b.streamName
}

func (b *bookTickerBatchSub) handleEvent(msg *PushDataV3MarketWrapper) {
	v := msg.GetPublicBookTickerBatch()
	if v == nil || msg.Symbol == nil {
		return
	}

	res, err := mapProtoBookTickerBatch(v, *msg.Symbol, msg.SendTime)
	if err != nil {
		if b.onInvalid != nil {
			b.onInvalid(err)
		}
		return
	}

	b.onData(res)
}

func (b *bookTickerBatchSub) id() string {
	return b.streamName
}

func (b *bookTickerBatchSub) params() any {
	return b.streamName
}

// BookTickerBatch represents a batch of book tickers.
type BookTickerBatch struct {
	Symbol   string
	Tickers  []BookTicker
	SendTime *int64
}

func mapProtoBookTickerBatch(msg *PublicBookTickerBatchV3Api, symbol string, sendTime *int64) (*BookTickerBatch, error) {
	var items []BookTicker

	for _, item := range msg.Items {
		if nil == item {
			return nil, errors.New("book ticker is nil")
		}

		bidPrice, err := decimal.NewFromString(item.BidPrice)
		if err != nil {
			return nil, fmt.Errorf("invalid bidPrice %q", item.BidPrice)
		}

		bidQty, err := decimal.NewFromString(item.BidQuantity)
		if err != nil {
			return nil, fmt.Errorf("invalid bidQuantity %q", item.BidQuantity)
		}

		askPrice, err := decimal.NewFromString(item.AskPrice)
		if err != nil {
			return nil, fmt.Errorf("invalid askPrice %q", item.AskPrice)
		}

		askQty, err := decimal.NewFromString(item.AskQuantity)
		if err != nil {
			return nil, fmt.Errorf("invalid askQuantity %q", item.AskQuantity)
		}

		items = append(items, BookTicker{
			BidPrice:    bidPrice,
			BidQuantity: bidQty,
			AskPrice:    askPrice,
			AskQuantity: askQty,
		})
	}

	return &BookTickerBatch{
		Symbol:   symbol,
		Tickers:  items,
		SendTime: sendTime,
	}, nil
}
