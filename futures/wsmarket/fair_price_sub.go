package wsmarket

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

type fairPriceSub struct {
	symbol    string
	onInvalid func(error)
	onData    func(FairPrice)
}

// NewFairPriceSub creates a new subscription for fair price updates.
//
// symbol is the trading pair, e.g., "BTC_USDT".
// onData is the callback function to handle the fair price update.
//
// Panics:
//   - symbol is empty
//   - onData is nil
func NewFairPriceSub(
	symbol string,
	onData func(FairPrice),
) Subscription {
	if symbol == "" {
		panic("NewFairPriceSub: invalid symbol name")
	}
	if onData == nil {
		panic("NewFairPriceSub: onData function is nil")
	}

	return &fairPriceSub{
		symbol: symbol,
		onData: onData,
	}
}

// SetOnInvalid sets a callback invoked when the subscription becomes invalid
// (e.g. malformed message, server rejection).
func (s *fairPriceSub) SetOnInvalid(f func(error)) Subscription {
	s.onInvalid = f
	return s
}

func (s *fairPriceSub) matches(msg *message) (bool, error) {
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

func (s *fairPriceSub) acceptEvent(msg *message) bool {
	return msg.Channel == "push."+s.channel() && msg.Symbol == s.symbol
}

func (s *fairPriceSub) handleEvent(msg *message) {
	var fairPrice FairPrice

	if err := json.Unmarshal(msg.Data, &fairPrice); err != nil {
		err = fmt.Errorf("failed to unmarshal data: %v, raw: %s", err, string(msg.Data))
		if s.onInvalid != nil {
			s.onInvalid(err)
		}
		return
	}

	fairPrice.SendTime = &msg.Ts
	s.onData(fairPrice)
}

func (s *fairPriceSub) id() string {
	return fmt.Sprintf("%s@%s", s.channel(), s.symbol)
}

func (s *fairPriceSub) channel() string {
	return "fair.price"
}

func (s *fairPriceSub) payload(op subscriptionOp) any {
	return wsRequestPayload{
		Method: fmt.Sprintf("%s.%s", op, s.channel()),
		Param: map[string]any{
			"symbol": s.symbol,
		},
	}
}

// FairPrice represents a fair price update.
type FairPrice struct {
	Symbol   string          `json:"symbol"`
	Price    decimal.Decimal `json:"price"`
	SendTime *int64
}
