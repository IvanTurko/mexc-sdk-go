package wsuser

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type spotAccountUpdateSub struct {
	streamName string
	onData     func(PrivateAccountUpdate)
	onInvalid  func(error)
}

// NewSpotAccountUpdateSub creates a new subscription for private account updates.
//
// onData is the callback function to handle the private account update.
//
// Panics:
//   - onData is nil
func NewSpotAccountUpdateSub(
	onData func(PrivateAccountUpdate),
) Subscription {
	if onData == nil {
		panic("NewSpotAccountUpdateSub: onData function is nil")
	}
	return &spotAccountUpdateSub{
		streamName: "spot@private.account.v3.api.pb",
		onData:     onData,
	}
}

// SetOnInvalid sets a callback invoked when the subscription becomes invalid
// (e.g. malformed message, server rejection).
func (s *spotAccountUpdateSub) SetOnInvalid(f func(error)) Subscription {
	s.onInvalid = f
	return s
}

func (s *spotAccountUpdateSub) matches(msg *message) (bool, error) {
	return msg.Msg == s.streamName, nil
}

func (s *spotAccountUpdateSub) acceptEvent(msg *PushDataV3UserWrapper) bool {
	return msg.GetChannel() == s.streamName
}

func (s *spotAccountUpdateSub) handleEvent(msg *PushDataV3UserWrapper) {
	v := msg.GetPrivateAccount()
	if v == nil {
		return
	}

	res, err := mapProtoPrivateAccountUpdate(v, msg.SendTime)
	if err != nil {
		if s.onInvalid != nil {
			s.onInvalid(err)
		}
		return
	}

	s.onData(res)
}

func (s *spotAccountUpdateSub) id() string {
	return s.streamName
}

func (s *spotAccountUpdateSub) params() any {
	return s.streamName
}

// PrivateAccountUpdate represents a private account update.
type PrivateAccountUpdate struct {
	VcoinName           string
	CoinID              string
	BalanceAmount       decimal.Decimal
	BalanceAmountChange decimal.Decimal
	FrozenAmount        decimal.Decimal
	FrozenAmountChange  decimal.Decimal
	Type                string
	Time                int64
	SendTime            *int64
}

func mapProtoPrivateAccountUpdate(msg *PrivateAccountV3Api, sendTime *int64) (PrivateAccountUpdate, error) {
	balanceAmount, err := decimal.NewFromString(msg.BalanceAmount)
	if err != nil {
		return PrivateAccountUpdate{}, fmt.Errorf("invalid balanceAmount %q", msg.BalanceAmount)
	}

	balanceAmountChange, err := decimal.NewFromString(msg.BalanceAmountChange)
	if err != nil {
		return PrivateAccountUpdate{}, fmt.Errorf("invalid balanceAmountChange %q", msg.BalanceAmountChange)
	}

	frozenAmount, err := decimal.NewFromString(msg.FrozenAmount)
	if err != nil {
		return PrivateAccountUpdate{}, fmt.Errorf("invalid frozenAmount %q", msg.FrozenAmount)
	}

	frozenAmountChange, err := decimal.NewFromString(msg.FrozenAmountChange)
	if err != nil {
		return PrivateAccountUpdate{}, fmt.Errorf("invalid frozenAmountChange %q", msg.FrozenAmountChange)
	}

	return PrivateAccountUpdate{
		VcoinName:           msg.VcoinName,
		CoinID:              msg.CoinId,
		BalanceAmount:       balanceAmount,
		BalanceAmountChange: balanceAmountChange,
		FrozenAmount:        frozenAmount,
		FrozenAmountChange:  frozenAmountChange,
		Type:                msg.Type,
		Time:                msg.Time,
		SendTime:            sendTime,
	}, nil
}
