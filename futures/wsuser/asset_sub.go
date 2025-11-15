package wsuser

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

type assetSub struct {
	onInvalid func(error)
	onData    func(Asset)
}

// NewAssetSub creates a new subscription for asset updates.
//
// onData is the callback function to handle the asset update.
//
// Panics:
//   - onData is nil
func NewAssetSub(
	onData func(Asset),
) Subscription {
	if onData == nil {
		panic("NewAssetSub: onData function is nil")
	}

	return &assetSub{
		onData: onData,
	}
}

// SetOnInvalid sets a callback invoked when the subscription becomes invalid
// (e.g. malformed message, server rejection).
func (a *assetSub) SetOnInvalid(f func(error)) Subscription {
	a.onInvalid = f
	return a
}

func (a *assetSub) acceptEvent(msg *message) bool {
	return msg.Channel == "push.personal."+a.channel()
}

func (a *assetSub) handleEvent(msg *message) {
	var event Asset

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

func (a *assetSub) id() string {
	return a.channel()
}

func (a *assetSub) channel() string {
	return "asset"
}

// Asset represents an asset update.
type Asset struct {
	Currency         string          `json:"currency"`
	PositionMargin   decimal.Decimal `json:"positionMargin"`
	FrozenBalance    decimal.Decimal `json:"frozenBalance"`
	AvailableBalance decimal.Decimal `json:"availableBalance"`
	CashBalance      decimal.Decimal `json:"cashBalance"`
	SendTime         *int64
}
