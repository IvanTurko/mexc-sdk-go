package wsuser

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type spotAccountDealsSub struct {
	streamName string
	onData     func(PrivateDeal)
	onInvalid  func(error)
}

// NewSpotAccountDealsSub creates a new subscription for private deal updates.
//
// onData is the callback function to handle the private deal update.
//
// Panics:
//   - onData is nil
func NewSpotAccountDealsSub(
	onData func(PrivateDeal),
) Subscription {
	if onData == nil {
		panic("NewSpotAccountDealsSub: onData function is nil")
	}
	return &spotAccountDealsSub{
		streamName: "spot@private.deals.v3.api.pb",
		onData:     onData,
	}
}

// SetOnInvalid sets a callback invoked when the subscription becomes invalid
// (e.g. malformed message, server rejection).
func (s *spotAccountDealsSub) SetOnInvalid(f func(error)) Subscription {
	s.onInvalid = f
	return s
}

func (s *spotAccountDealsSub) matches(msg *message) (bool, error) {
	return msg.Msg == s.streamName, nil
}

func (s *spotAccountDealsSub) acceptEvent(msg *PushDataV3UserWrapper) bool {
	return msg.GetChannel() == s.streamName
}

func (s *spotAccountDealsSub) handleEvent(msg *PushDataV3UserWrapper) {
	v := msg.GetPrivateDeals()
	if v == nil || msg.Symbol == nil {
		return
	}

	res, err := mapProtoPrivateDeal(v, *msg.Symbol, msg.SendTime)
	if err != nil {
		if s.onInvalid != nil {
			s.onInvalid(err)
		}
		return
	}

	s.onData(res)
}

func (s *spotAccountDealsSub) id() string {
	return s.streamName
}

func (s *spotAccountDealsSub) params() any {
	return s.streamName
}

// PrivateDeal represents a private deal update.
type PrivateDeal struct {
	Symbol        string
	Price         decimal.Decimal
	Quantity      decimal.Decimal
	Amount        decimal.Decimal
	TradeSide     TradeSide
	TradeId       string
	OrderId       string
	ClientOrderId string
	FeeAmount     decimal.Decimal
	FeeCurrency   string
	Time          int64
	IsMaker       bool
	IsSelfTrade   bool
	SendTime      *int64
}

func mapProtoPrivateDeal(msg *PrivateDealsV3Api, symbol string, sendTime *int64) (PrivateDeal, error) {
	price, err := decimal.NewFromString(msg.Price)
	if err != nil {
		return PrivateDeal{}, fmt.Errorf("invalid price %q", msg.Price)
	}

	quantity, err := decimal.NewFromString(msg.Quantity)
	if err != nil {
		return PrivateDeal{}, fmt.Errorf("invalid quantity %q", msg.Quantity)
	}

	amount, err := decimal.NewFromString(msg.Amount)
	if err != nil {
		return PrivateDeal{}, fmt.Errorf("invalid amount %q", msg.Amount)
	}

	feeAmount, err := decimal.NewFromString(msg.FeeAmount)
	if err != nil {
		return PrivateDeal{}, fmt.Errorf("invalid feeAmount %q", msg.FeeAmount)
	}

	tradeSide, err := parseTradeSide(msg.TradeType)
	if err != nil {
		return PrivateDeal{}, err
	}

	return PrivateDeal{
		Symbol:        symbol,
		Price:         price,
		Quantity:      quantity,
		Amount:        amount,
		TradeSide:     tradeSide,
		TradeId:       msg.TradeId,
		OrderId:       msg.OrderId,
		ClientOrderId: msg.ClientOrderId,
		FeeAmount:     feeAmount,
		FeeCurrency:   msg.FeeCurrency,
		Time:          msg.Time,
		IsMaker:       msg.IsMaker,
		IsSelfTrade:   msg.IsSelfTrade,
		SendTime:      sendTime,
	}, nil
}
