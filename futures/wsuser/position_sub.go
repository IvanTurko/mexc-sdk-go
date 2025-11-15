package wsuser

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

// PositionState represents the state of a position.
type PositionState int

const (
	_ PositionState = iota
	PositionStateHolding
	PositionStateSystemHolding
	PositionStateClosed
)

type positionSub struct {
	onInvalid func(error)
	onData    func(*PositionEvent)
}

// NewPositionSub creates a new subscription for position updates.
//
// onData is the callback function to handle the position update.
//
// Panics:
//   - onData is nil
func NewPositionSub(
	onData func(*PositionEvent),
) Subscription {
	if onData == nil {
		panic("NewPositionSub: onData function is nil")
	}

	return &positionSub{
		onData: onData,
	}
}

// SetOnInvalid sets a callback invoked when the subscription becomes invalid
// (e.g. malformed message, server rejection).
func (p *positionSub) SetOnInvalid(f func(error)) Subscription {
	p.onInvalid = f
	return p
}

func (p *positionSub) acceptEvent(msg *message) bool {
	return msg.Channel == "push.personal."+p.channel()
}

func (p *positionSub) handleEvent(msg *message) {
	var event *PositionEvent

	if err := json.Unmarshal(msg.Data, &event); err != nil {
		err = fmt.Errorf("failed to unmarshal data: %v, raw: %s", err, string(msg.Data))
		if p.onInvalid != nil {
			p.onInvalid(err)
		}
		return
	}

	event.SendTime = &msg.Ts
	p.onData(event)
}

func (p *positionSub) id() string {
	return p.channel()
}

func (p *positionSub) channel() string {
	return "position"
}

// PositionEvent represents a position update event.
type PositionEvent struct {
	PositionId     int64
	Symbol         string
	HoldVol        decimal.Decimal
	PositionType   PositionType
	OpenType       OpenType
	State          PositionState
	FrozenVol      decimal.Decimal
	CloseVol       decimal.Decimal
	HoldAvgPrice   decimal.Decimal
	CloseAvgPrice  decimal.Decimal
	OpenAvgPrice   decimal.Decimal
	LiquidatePrice decimal.Decimal
	Oim            decimal.Decimal
	AdlLevel       int
	Im             decimal.Decimal
	HoldFee        decimal.Decimal
	Realised       decimal.Decimal
	AutoAddIm      bool
	Leverage       int
	SendTime       *int64
}

func (p *PositionEvent) UnmarshalJSON(data []byte) error {
	var tmp positionJSON

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	positionType, err := parsePositionType(tmp.PositionType)
	if err != nil {
		return err
	}

	openType, err := parseOpenType(tmp.OpenType)
	if err != nil {
		return err
	}

	state, err := parsePositionState(tmp.State)
	if err != nil {
		return err
	}

	p.PositionId = tmp.PositionId
	p.Symbol = tmp.Symbol
	p.HoldVol = tmp.HoldVol
	p.PositionType = positionType
	p.OpenType = openType
	p.State = state
	p.FrozenVol = tmp.FrozenVol
	p.CloseVol = tmp.CloseVol
	p.HoldAvgPrice = tmp.HoldAvgPrice
	p.CloseAvgPrice = tmp.CloseAvgPrice
	p.OpenAvgPrice = tmp.OpenAvgPrice
	p.LiquidatePrice = tmp.LiquidatePrice
	p.Oim = tmp.Oim
	p.AdlLevel = tmp.AdlLevel
	p.Im = tmp.Im
	p.HoldFee = tmp.HoldFee
	p.Realised = tmp.Realised
	p.AutoAddIm = tmp.AutoAddIm
	p.Leverage = tmp.Leverage

	return nil
}

type positionJSON struct {
	PositionId     int64           `json:"positionId"`
	Symbol         string          `json:"symbol"`
	HoldVol        decimal.Decimal `json:"holdVol"`
	PositionType   int             `json:"positionType"`
	OpenType       int             `json:"openType"`
	State          int             `json:"state"`
	FrozenVol      decimal.Decimal `json:"frozenVol"`
	CloseVol       decimal.Decimal `json:"closeVol"`
	HoldAvgPrice   decimal.Decimal `json:"holdAvgPrice"`
	CloseAvgPrice  decimal.Decimal `json:"closeAvgPrice"`
	OpenAvgPrice   decimal.Decimal `json:"openAvgPrice"`
	LiquidatePrice decimal.Decimal `json:"liquidatePrice"`
	Oim            decimal.Decimal `json:"oim"`
	AdlLevel       int             `json:"adlLevel"`
	Im             decimal.Decimal `json:"im"`
	HoldFee        decimal.Decimal `json:"holdFee"`
	Realised       decimal.Decimal `json:"realised"`
	AutoAddIm      bool            `json:"autoAddIm"`
	Leverage       int             `json:"leverage"`
}



func parsePositionState(code int) (PositionState, error) {
	switch code {
	case 1:
		return PositionStateHolding, nil
	case 2:
		return PositionStateSystemHolding, nil
	case 3:
		return PositionStateClosed, nil
	default:
		return 0, fmt.Errorf("unknown position state code: %d", code)
	}
}
