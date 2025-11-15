package wsmarket

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

type bookDepthSub struct {
	symbol    string
	onInvalid func(error)
	onData    func(*DepthSnapshot)
}

// NewBookDepthSub creates a subscription for order book depth updates.
//
// symbol is the trading pair, e.g. "BTC_USDT".
// onData is called for each depth snapshot.
//
// Panics:
//   - symbol is empty
//   - onData is nil
func NewBookDepthSub(
	symbol string,
	onData func(*DepthSnapshot),
) Subscription {
	if symbol == "" {
		panic("NewBookDepthSub: invalid symbol name")
	}
	if onData == nil {
		panic("NewBookDepthSub: onData function is nil")
	}

	return &bookDepthSub{
		symbol: symbol,
		onData: onData,
	}
}

// SetOnInvalid sets a callback invoked when the subscription becomes invalid
// (e.g. malformed message, server rejection).
func (d *bookDepthSub) SetOnInvalid(f func(error)) Subscription {
	d.onInvalid = f
	return d
}

func (d *bookDepthSub) matches(msg *message) (bool, error) {
	if msg.Channel == "rs.sub."+d.channel() {
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

func (d *bookDepthSub) acceptEvent(msg *message) bool {
	return msg.Channel == "push."+d.channel() && msg.Symbol == d.symbol
}

func (d *bookDepthSub) handleEvent(msg *message) {
	var depth *DepthSnapshot

	if err := json.Unmarshal(msg.Data, &depth); err != nil {
		err = fmt.Errorf("failed to unmarshal data: %v, raw: %s", err, string(msg.Data))
		if d.onInvalid != nil {
			d.onInvalid(err)
		}
		return
	}

	depth.Symbol = msg.Symbol
	depth.SendTime = &msg.Ts
	d.onData(depth)
}

func (d *bookDepthSub) id() string {
	return fmt.Sprintf("%s@%s", d.channel(), d.symbol)
}

func (d *bookDepthSub) channel() string {
	return "depth"
}

func (d *bookDepthSub) payload(op subscriptionOp) any {
	return wsRequestPayload{
		Method: fmt.Sprintf("%s.%s", op, d.channel()),
		Param: map[string]any{
			"symbol":   d.symbol,
			"compress": true,
		},
	}
}

// DepthSnapshot represents a snapshot of the order book depth.
type DepthSnapshot struct {
	Symbol   string
	Asks     []DepthLevel
	Bids     []DepthLevel
	Version  int64
	SendTime *int64
}

// DepthLevel represents a level in the order book.
type DepthLevel struct {
	Price      decimal.Decimal
	OrderCount decimal.Decimal
	Quantity   decimal.Decimal
}

type depthSnapshotJSON struct {
	Asks    [][]any `json:"asks"`
	Bids    [][]any `json:"bids"`
	Version int64   `json:"version"`
}

func parseDepthLevels(raw [][]any) ([]DepthLevel, error) {
	levels := make([]DepthLevel, 0, len(raw))

	for _, item := range raw {
		if len(item) != 3 {
			return nil, fmt.Errorf("invalid depth level format %v", item)
		}

		price, ok := item[0].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid price %v", item[0])
		}

		orderCount, ok := item[1].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid order count %v", item[1])
		}

		quantity, ok := item[2].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid quantity %v", item[2])
		}

		level := DepthLevel{
			Price:      decimal.NewFromFloat(price),
			OrderCount: decimal.NewFromFloat(orderCount),
			Quantity:   decimal.NewFromFloat(quantity),
		}
		levels = append(levels, level)
	}

	return levels, nil
}

func (d *DepthSnapshot) UnmarshalJSON(data []byte) error {
	var tmp depthSnapshotJSON
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	asks, err := parseDepthLevels(tmp.Asks)
	if err != nil {
		return fmt.Errorf("invalid asks: %v", err)
	}

	bids, err := parseDepthLevels(tmp.Bids)
	if err != nil {
		return fmt.Errorf("invalid bids: %v", err)
	}

	d.Asks = asks
	d.Bids = bids
	d.Version = tmp.Version

	return nil
}
