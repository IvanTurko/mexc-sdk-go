package wsmarket

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

type fundingRateSub struct {
	symbol    string
	onInvalid func(error)
	onData    func(FundingRate)
}

// NewFundingRateSub creates a new subscription for funding rate updates.
//
// symbol is the trading pair, e.g., "BTC_USDT".
// onData is the callback function to handle the funding rate update.
//
// Panics:
//   - symbol is empty
//   - onData is nil
func NewFundingRateSub(
	symbol string,
	onData func(FundingRate),
) Subscription {
	if symbol == "" {
		panic("NewFundingRateSub: invalid symbol name")
	}
	if onData == nil {
		panic("NewFundingRateSub: onData function is nil")
	}

	return &fundingRateSub{
		symbol: symbol,
		onData: onData,
	}
}

// SetOnInvalid sets a callback invoked when the subscription becomes invalid
// (e.g. malformed message, server rejection).
func (s *fundingRateSub) SetOnInvalid(f func(error)) Subscription {
	s.onInvalid = f
	return s
}

func (s *fundingRateSub) matches(msg *message) (bool, error) {
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

func (s *fundingRateSub) acceptEvent(msg *message) bool {
	return msg.Channel == "push."+s.channel() && msg.Symbol == s.symbol
}

func (s *fundingRateSub) handleEvent(msg *message) {
	var fundingRate FundingRate

	if err := json.Unmarshal(msg.Data, &fundingRate); err != nil {
		err = fmt.Errorf("failed to unmarshal data: %v, raw: %s", err, string(msg.Data))
		if s.onInvalid != nil {
			s.onInvalid(err)
		}
		return
	}

	fundingRate.SendTime = &msg.Ts
	s.onData(fundingRate)
}

func (s *fundingRateSub) id() string {
	return fmt.Sprintf("%s@%s", s.channel(), s.symbol)
}

func (s *fundingRateSub) channel() string {
	return "funding.rate"
}

func (s *fundingRateSub) payload(op subscriptionOp) any {
	return wsRequestPayload{
		Method: fmt.Sprintf("%s.%s", op, s.channel()),
		Param: map[string]any{
			"symbol": s.symbol,
		},
	}
}

// FundingRate represents a funding rate update.
type FundingRate struct {
	Symbol         string          `json:"symbol"`
	FundingRate    decimal.Decimal `json:"rate"`
	NextSettleTime int64           `json:"nextSettleTime"`
	SendTime       *int64
}
