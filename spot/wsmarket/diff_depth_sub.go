package wsmarket

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/shopspring/decimal"
)

type diffDepthSub struct {
	streamName string
	onData     func(*DepthDelta)
	onInvalid  func(error)
}

// NewDiffDepthSub creates a subscription for aggregated incremental depth updates.
//
// symbol is the trading pair (e.g. "BTCUSDT").
// interval defines how frequently updates are aggregated on the server side.
// onData is invoked for each received DepthDelta.
//
// Panics:
//   - symbol is empty
//   - interval is invalid
//   - onData is nil
func NewDiffDepthSub(
	symbol string,
	interval UpdateInterval,
	onData func(*DepthDelta),
) Subscription {
	if symbol == "" {
		panic("NewDiffDepthSub: invalid symbol name")
	}
	if !interval.isValid() {
		panic("NewDiffDepthSub: invalid update interval: " + string(interval))
	}
	if onData == nil {
		panic("NewDiffDepthSub: onData function is nil")
	}
	stream := fmt.Sprintf("spot@public.aggre.depth.v3.api.pb@%s@%s", interval, symbol)

	return &diffDepthSub{
		streamName: stream,
		onData:     onData,
	}
}

// SetOnInvalid registers a callback that triggers if the subscription becomes
// invalid (malformed update, server-side error, rejected channel, etc.).
func (d *diffDepthSub) SetOnInvalid(f func(error)) Subscription {
	d.onInvalid = f
	return d
}

func (d *diffDepthSub) matches(msg *message) (bool, error) {
	return msg.Msg == d.streamName, nil
}

func (d *diffDepthSub) acceptEvent(msg *PushDataV3MarketWrapper) bool {
	return msg.GetChannel() == d.streamName
}

func (d *diffDepthSub) handleEvent(msg *PushDataV3MarketWrapper) {
	v := msg.GetPublicAggreDepths()
	if v == nil || msg.Symbol == nil {
		return
	}

	res, err := mapProtoDepthDelta(v, *msg.Symbol, msg.SendTime)
	if err != nil {
		if d.onInvalid != nil {
			d.onInvalid(err)
		}
		return
	}

	d.onData(res)
}

func (d *diffDepthSub) id() string {
	return d.streamName
}

func (d *diffDepthSub) params() any {
	return d.streamName
}

// DepthDelta represents a single incremental depth update.
type DepthDelta struct {
	Symbol      string
	Asks        []DepthLevel
	Bids        []DepthLevel
	EventType   string
	FromVersion uint64
	ToVersion   uint64
	SendTime    *int64
}

// DepthLevel represents a single price level delta in the order book.
type DepthLevel struct {
	Price    decimal.Decimal
	Quantity decimal.Decimal
}

func mapProtoDepthDelta(msg *PublicAggreDepthsV3Api, symbol string, sendTime *int64) (*DepthDelta, error) {
	fromVersion, err := strconv.ParseUint(msg.FromVersion, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid fromVersion %q", msg.FromVersion)
	}

	toVersion, err := strconv.ParseUint(msg.ToVersion, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid toVersion %q", msg.ToVersion)
	}

	asks, err := mapProtoDepthLevelFromAggreDepth(msg.Asks)
	if err != nil {
		return nil, fmt.Errorf("invalid asks: %v", err)
	}

	bids, err := mapProtoDepthLevelFromAggreDepth(msg.Bids)
	if err != nil {
		return nil, fmt.Errorf("invalid bids: %v", err)
	}

	return &DepthDelta{
		Symbol:      symbol,
		Asks:        asks,
		Bids:        bids,
		EventType:   msg.EventType,
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		SendTime:    sendTime,
	}, nil
}

func mapProtoDepthLevelFromAggreDepth(items []*PublicAggreDepthV3ApiItem) ([]DepthLevel, error) {
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
