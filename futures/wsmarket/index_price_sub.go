package wsmarket

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

type indexPriceSub struct {
	symbol    string
	onInvalid func(error)
	onData    func(IndexPrice)
}

// NewIndexPriceSub creates a new subscription for index price updates.
//
// symbol is the trading pair, e.g., "BTC_USDT".
// onData is the callback function to handle the index price update.
//
// Panics:
//   - symbol is empty
//   - onData is nil
func NewIndexPriceSub(
	symbol string,
	onData func(IndexPrice),
) Subscription {
	if symbol == "" {
		panic("NewIndexPriceSub: invalid symbol name")
	}
	if onData == nil {
		panic("NewIndexPriceSub: onData function is nil")
	}

	return &indexPriceSub{
		symbol: symbol,
		onData: onData,
	}
}

// SetOnInvalid sets a callback invoked when the subscription becomes invalid
// (e.g. malformed message, server rejection).
func (s *indexPriceSub) SetOnInvalid(f func(error)) Subscription {
	s.onInvalid = f
	return s
}

func (s *indexPriceSub) matches(msg *message) (bool, error) {
	if msg.Channel == "rs.sub."+s.channel() {
		var r string
		if err := json.Unmarshal(msg.Data, &r); err != nil {
			return false, fmt.Errorf("invalid success payload: %s", string(msg.Data))
		}
		if r == "success" {
			return true, nil
		}
	}

	if msg.Channel == "rs.error" {
		return true, fmt.Errorf("sub failed: %s", string(msg.Data))
	}

	return false, nil
}

func (s *indexPriceSub) acceptEvent(msg *message) bool {
	return msg.Channel == "push."+s.channel() && msg.Symbol == s.symbol
}

func (s *indexPriceSub) handleEvent(msg *message) {
	var indexPrice IndexPrice

	if err := json.Unmarshal(msg.Data, &indexPrice); err != nil {
		err = fmt.Errorf("failed to unmarshal data: %v, raw: %s", err, string(msg.Data))
		if s.onInvalid != nil {
			s.onInvalid(err)
		}
		return
	}

	indexPrice.SendTime = &msg.Ts
	s.onData(indexPrice)
}

func (s *indexPriceSub) id() string {
	return fmt.Sprintf("%s@%s", s.channel(), s.symbol)
}

func (s *indexPriceSub) channel() string {
	return "index.price"
}

func (s *indexPriceSub) payload(op subscriptionOp) any {
	return wsRequestPayload{
		Method: fmt.Sprintf("%s.%s", op, s.channel()),
		Param: map[string]any{
			"symbol": s.symbol,
		},
	}
}

// IndexPrice represents an index price update.
type IndexPrice struct {
	Symbol   string          `json:"symbol"`
	Price    decimal.Decimal `json:"price"`
	SendTime *int64
}
