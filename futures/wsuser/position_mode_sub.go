package wsuser

import (
	"encoding/json"
	"fmt"
)

// PositionMode represents the position mode.
type PositionMode string

const (
	PositionModeHedge  PositionMode = "HEDGE"
	PositionModeOneWay PositionMode = "ONE-WAY"
)

type positionModeSub struct {
	onInvalid func(error)
	onData    func(PositionModeEvent)
}

// NewPositionModeSub creates a new subscription for position mode updates.
//
// onData is the callback function to handle the position mode update.
//
// Panics:
//   - onData is nil
func NewPositionModeSub(
	onData func(PositionModeEvent),
) Subscription {
	if onData == nil {
		panic("NewPositionModeSub: onData function is nil")
	}

	return &positionModeSub{
		onData: onData,
	}
}

// SetOnInvalid sets a callback invoked when the subscription becomes invalid
// (e.g. malformed message, server rejection).
func (p *positionModeSub) SetOnInvalid(f func(error)) Subscription {
	p.onInvalid = f
	return p
}

func (p *positionModeSub) acceptEvent(msg *message) bool {
	return msg.Channel == "push.personal."+p.channel()
}

func (p *positionModeSub) handleEvent(msg *message) {
	var mode PositionModeEvent

	if err := json.Unmarshal(msg.Data, &mode); err != nil {
		err = fmt.Errorf("failed to unmarshal data: %v, raw: %s", err, string(msg.Data))
		if p.onInvalid != nil {
			p.onInvalid(err)
		}
		return
	}

	mode.SendTime = &msg.Ts
	p.onData(mode)
}

func (p *positionModeSub) id() string {
	return p.channel()
}

func (p *positionModeSub) channel() string {
	return "position.mode"
}

// PositionModeEvent represents a position mode update event.
type PositionModeEvent struct {
	PositionMode PositionMode
	SendTime     *int64
}

func (p *PositionModeEvent) UnmarshalJSON(data []byte) error {
	var tmp struct {
		PositionMode int `json:"positionMode"`
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	mode, err := parsePositionMode(tmp.PositionMode)
	if err != nil {
		return err
	}

	p.PositionMode = mode
	return nil
}

func parsePositionMode(code int) (PositionMode, error) {
	switch code {
	case 1:
		return PositionModeHedge, nil
	case 2:
		return PositionModeOneWay, nil
	default:
		return "", fmt.Errorf("unknown position mode code: %d", code)
	}
}
