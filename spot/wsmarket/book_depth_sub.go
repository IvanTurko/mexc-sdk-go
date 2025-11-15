package wsmarket

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/shopspring/decimal"
)

type bookDepthSub struct {
	streamName string
	onData     func(*DepthSnapshot)
	onInvalid  func(error)
}

// NewBookDepthSub returns a subscription for level-based order book snapshots.
//
// symbol is the trading pair (e.g. "BTCUSDT").
// level is the depth size; must be valid.
// onData is called for every received DepthSnapshot.
//
// Panics:
//   - symbol is empty
//   - level is invalid
//   - onData is nil
func NewBookDepthSub(
	symbol string,
	level DepthSize,
	onData func(*DepthSnapshot),
) Subscription {
	if symbol == "" {
		panic("NewBookDepthSub: invalid symbol name")
	}
	if !level.isValid() {
		panic("NewBookDepthSub: invalid depth level: " + strconv.FormatUint(uint64(level), 10))
	}
	if onData == nil {
		panic("NewBookDepthSub: onData function is nil")
	}
	stream := fmt.Sprintf("spot@public.limit.depth.v3.api.pb@%s@%d", symbol, level)

	return &bookDepthSub{
		streamName: stream,
		onData:     onData,
	}
}

// SetOnInvalid sets a callback invoked when the subscription becomes invalid
// (e.g. malformed message, server rejection).
func (b *bookDepthSub) SetOnInvalid(f func(error)) Subscription {
	b.onInvalid = f
	return b
}

func (b *bookDepthSub) matches(msg *message) (bool, error) {
	return msg.Msg == b.streamName, nil
}

func (b *bookDepthSub) acceptEvent(msg *PushDataV3MarketWrapper) bool {
	return msg.GetChannel() == b.streamName
}

func (b *bookDepthSub) handleEvent(msg *PushDataV3MarketWrapper) {
	v := msg.GetPublicLimitDepths()
	if v == nil || msg.Symbol == nil {
		return
	}

	res, err := mapProtoLimitDepthSnapshot(v, *msg.Symbol, msg.SendTime)
	if err != nil {
		if b.onInvalid != nil {
			b.onInvalid(err)
		}
		return
	}

	b.onData(res)
}

func (b *bookDepthSub) id() string {
	return b.streamName
}

func (b *bookDepthSub) params() any {
	return b.streamName
}

// DepthSnapshot represents a snapshot of the order book depth.
type DepthSnapshot struct {
	Symbol   string
	Asks     []DepthLevel
	Bids     []DepthLevel
	Version  int64
	SendTime *int64
}

func mapProtoLimitDepthSnapshot(msg *PublicLimitDepthsV3Api, symbol string, sendTime *int64) (*DepthSnapshot, error) {
	version, err := strconv.ParseInt(msg.Version, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid version %q", msg.Version)
	}

	asks, err := mapProtoDepthLevelFromLimitDepth(msg.Asks)
	if err != nil {
		return nil, fmt.Errorf("invalid asks: %v", err)
	}

	bids, err := mapProtoDepthLevelFromLimitDepth(msg.Bids)
	if err != nil {
		return nil, fmt.Errorf("invalid bids: %v", err)
	}

	return &DepthSnapshot{
		Symbol:   symbol,
		Asks:     asks,
		Bids:     bids,
		Version:  version,
		SendTime: sendTime,
	}, nil
}

func mapProtoDepthLevelFromLimitDepth(items []*PublicLimitDepthV3ApiItem) ([]DepthLevel, error) {
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
