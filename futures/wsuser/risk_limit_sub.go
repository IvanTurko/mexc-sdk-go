package wsuser

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

// RiskSource represents the source of a risk limit event.
type RiskSource string

const (
	RiskSourceOther              RiskSource = "OTHER"
	RiskSourceLiquidationService RiskSource = "LIQUIDATION SERVICE"
)

type riskLimitSub struct {
	onInvalid func(error)
	onData    func(RiskLimitEvent)
}

// NewRiskLimitSub creates a new subscription for risk limit updates.
//
// onData is the callback function to handle the risk limit update.
//
// Panics:
//   - onData is nil
func NewRiskLimitSub(
	onData func(RiskLimitEvent),
) Subscription {
	if onData == nil {
		panic("NewRiskLimitSub: onData function is nil")
	}

	return &riskLimitSub{
		onData: onData,
	}
}

// SetOnInvalid sets a callback invoked when the subscription becomes invalid
// (e.g. malformed message, server rejection).
func (r *riskLimitSub) SetOnInvalid(f func(error)) Subscription {
	r.onInvalid = f
	return r
}

func (r *riskLimitSub) acceptEvent(msg *message) bool {
	return msg.Channel == "push.personal."+r.channel()
}

func (r *riskLimitSub) handleEvent(msg *message) {
	var event RiskLimitEvent

	if err := json.Unmarshal(msg.Data, &event); err != nil {
		err = fmt.Errorf("failed to unmarshal data: %v, raw: %s", err, string(msg.Data))
		if r.onInvalid != nil {
			r.onInvalid(err)
		}
		return
	}

	event.SendTime = &msg.Ts
	r.onData(event)
}

func (r *riskLimitSub) id() string {
	return r.channel()
}

func (r *riskLimitSub) channel() string {
	return "risk.limit"
}

// RiskLimitEvent represents a risk limit update event.
type RiskLimitEvent struct {
	Symbol       string
	PositionType PositionType
	RiskSource   RiskSource
	Level        int
	MaxVol       decimal.Decimal
	MaxLeverage  int
	Mmr          decimal.Decimal
	Imr          decimal.Decimal
	SendTime     *int64
}

func (r *RiskLimitEvent) UnmarshalJSON(data []byte) error {
	var tmp riskLimitJSON

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	positionType, err := parsePositionType(tmp.PositionType)
	if err != nil {
		return err
	}

	riskSource, err := parseRiskSource(tmp.RiskSource)
	if err != nil {
		return err
	}

	r.Symbol = tmp.Symbol
	r.PositionType = positionType
	r.RiskSource = riskSource
	r.Level = tmp.Level
	r.MaxVol = tmp.MaxVol
	r.MaxLeverage = tmp.MaxLeverage
	r.Mmr = tmp.Mmr
	r.Imr = tmp.Imr

	return nil
}

type riskLimitJSON struct {
	Symbol       string          `json:"symbol"`
	PositionType int             `json:"positionType"`
	RiskSource   int             `json:"riskSource"`
	Level        int             `json:"level"`
	MaxVol       decimal.Decimal `json:"maxVol"`
	MaxLeverage  int             `json:"maxLeverage"`
	Mmr          decimal.Decimal `json:"mmr"`
	Imr          decimal.Decimal `json:"imr"`
}

func parseRiskSource(code int) (RiskSource, error) {
	switch code {
	case 0:
		return RiskSourceOther, nil
	case 1:
		return RiskSourceLiquidationService, nil
	default:
		return "", fmt.Errorf("unknown risk source code: %d", code)
	}
}
