package wsuser

import (
	"encoding/json"
	"fmt"
)

// AdlLevel represents the auto-deleveraging level.
type AdlLevel int

const (
	_ AdlLevel = iota
	AdlLevel1
	AdlLevel2
	AdlLevel3
	AdlLevel4
	AdlLevel5
)

func (l AdlLevel) String() string {
	return [...]string{"0", "1", "2", "3", "4", "5"}[l]
}

func (l *AdlLevel) UnmarshalJSON(data []byte) error {
	var level int
	if err := json.Unmarshal(data, &level); err != nil {
		return err
	}

	if level < 0 || level > 5 {
		return fmt.Errorf("invalid adl level code: %d", level)
	}
	*l = AdlLevel(level)
	return nil
}

// AdlLevelSub is a subscription for ADL level updates.
type AdlLevelSub struct {
	onInvalid func(error)
	onData    func(AdlLevelEvent)
}

// NewAdlLevelSub creates a new subscription for ADL level updates.
//
// onData is the callback function to handle the ADL level update.
//
// Panics:
//   - onData is nil
func NewAdlLevelSub(
	onData func(AdlLevelEvent),
) Subscription {
	if onData == nil {
		panic("NewAdlLevelSub: onData function is nil")
	}

	return &AdlLevelSub{
		onData: onData,
	}
}

// SetOnInvalid sets a callback invoked when the subscription becomes invalid
// (e.g. malformed message, server rejection).
func (a *AdlLevelSub) SetOnInvalid(f func(error)) Subscription {
	a.onInvalid = f
	return a
}

func (a *AdlLevelSub) acceptEvent(msg *message) bool {
	return msg.Channel == "push.personal."+a.channel()
}

func (a *AdlLevelSub) handleEvent(msg *message) {
	var event AdlLevelEvent

	if err := json.Unmarshal(msg.Data, &event); err != nil {
		err = fmt.Errorf("failed to unmarshal data: %v, raw: %s", err, string(msg.Data))
		if a.onInvalid != nil {
			a.onInvalid(err)
		}
		return
	}

	event.SendTime = &msg.Ts
	a.onData(event)
}

func (a *AdlLevelSub) id() string {
	return a.channel()
}

func (a *AdlLevelSub) channel() string {
	return "adl.level"
}

// AdlLevelEvent represents an ADL level update event.
type AdlLevelEvent struct {
	AdlLevel   AdlLevel `json:"adlLevel"`
	PositionId int64    `json:"positionId"`
	SendTime   *int64
}
