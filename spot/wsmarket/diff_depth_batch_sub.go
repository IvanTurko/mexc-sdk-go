package wsmarket

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/shopspring/decimal"
)

type diffDepthBatchSub struct {
	streamName string
	onData     func(*DepthBatch)
	onInvalid  func(error)
}

// NewDiffDepthBatchSub creates a subscription for incremental (diff) depth updates.
//
// symbol is the trading pair (e.g. "BTCUSDT").
// onData is invoked for each received DepthBatch.
//
// Panics:
//   - symbol is empty
//   - onData is nil
func NewDiffDepthBatchSub(
	symbol string,
	onData func(*DepthBatch),
) Subscription {
	if symbol == "" {
		panic("NewDiffDepthBatchSub: invalid symbol name")
	}
	if onData == nil {
		panic("NewDiffDepthBatchSub: onData function is nil")
	}
	stream := fmt.Sprintf("spot@public.increase.depth.batch.v3.api.pb@%s", symbol)

	return &diffDepthBatchSub{
		streamName: stream,
		onData:     onData,
	}
}

// SetOnInvalid registers a callback that is triggered when the subscription
// becomes invalid (malformed data, server rejection, etc.).
func (d *diffDepthBatchSub) SetOnInvalid(f func(error)) Subscription {
	d.onInvalid = f
	return d
}

func (d *diffDepthBatchSub) matches(msg *message) (bool, error) {
	return msg.Msg == d.streamName, nil
}

func (d *diffDepthBatchSub) acceptEvent(msg *PushDataV3MarketWrapper) bool {
	return msg.GetChannel() == d.streamName
}

func (d *diffDepthBatchSub) handleEvent(msg *PushDataV3MarketWrapper) {
	v := msg.GetPublicIncreaseDepthsBatch()
	if v == nil || msg.Symbol == nil {
		return
	}

	res, err := mapProtoDepthBatch(v, *msg.Symbol, msg.SendTime)
	if err != nil {
		if d.onInvalid != nil {
			d.onInvalid(err)
		}
		return
	}

	d.onData(res)
}

func (d *diffDepthBatchSub) id() string {
	return d.streamName
}

func (d *diffDepthBatchSub) params() any {
	return d.streamName
}

// DepthBatch represents an incremental depth update batch.
type DepthBatch struct {
	Symbol   string
	Items    []DepthChange
	SendTime *int64
}

// DepthChange is a single incremental update.
type DepthChange struct {
	Version int64
	Bids    []DepthLevel
	Asks    []DepthLevel
}

func mapProtoDepthBatch(msg *PublicIncreaseDepthsBatchV3Api, symbol string, sendTime *int64) (*DepthBatch, error) {
	var items []DepthChange

	for _, item := range msg.Items {
		if nil == item {
			return nil, errors.New("depth change is nil")
		}

		versionStr := item.Version
		version, err := strconv.ParseInt(versionStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid version %q", versionStr)
		}

		asks, err := mapProtoDepthLevelFromIncreaseDepth(item.Asks)
		if err != nil {
			return nil, fmt.Errorf("invalid asks: %v", err)
		}

		bids, err := mapProtoDepthLevelFromIncreaseDepth(item.Bids)
		if err != nil {
			return nil, fmt.Errorf("invalid bids: %v", err)
		}

		items = append(items, DepthChange{
			Version: version,
			Bids:    bids,
			Asks:    asks,
		})
	}

	return &DepthBatch{
		Symbol:   symbol,
		Items:    items,
		SendTime: sendTime,
	}, nil
}

func mapProtoDepthLevelFromIncreaseDepth(items []*PublicIncreaseDepthV3ApiItem) ([]DepthLevel, error) {
	var result []DepthLevel

	for _, item := range items {
		if nil == item {
			return nil, errors.New("depth level is nil")
		}

		price, err := decimal.NewFromString(item.Price)
		if err != nil {
			return nil, fmt.Errorf("invalid price %q", item.Price)
		}

		qty, err := decimal.NewFromString(item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("invalid quantity %q", item.Quantity)
		}

		result = append(result, DepthLevel{
			Price:    price,
			Quantity: qty,
		})
	}

	return result, nil
}
