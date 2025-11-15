package wsuser

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// TradeSide indicates whether the trade was buyer or seller initiated.
type TradeSide string

const (
	TradeSideBuy  TradeSide = "BUY"
	TradeSideSell TradeSide = "SELL"
)

// OrderType represents the type of an order.
type OrderType string

const (
	OrderTypeLimitOrder         OrderType = "LIMIT_ORDER"
	OrderTypePostOnly           OrderType = "POST_ONLY"
	OrderTypeImmediateOrCancel  OrderType = "IMMEDIATE_OR_CANCEL"
	OrderTypeFillOrKill         OrderType = "FILL_OR_KILL"
	OrderTypeMarketOrder        OrderType = "MARKET_ORDER"
	OrderTypeStopLossTakeProfit OrderType = "STOP_LOSS_TAKE_PROFIT"
)

// OrderStatus represents the current status of an order.
type OrderStatus string

const (
	OrderStatusNotTraded         OrderStatus = "NOT_TRADED"
	OrderStatusFullyTraded       OrderStatus = "FULLY_TRADED"
	OrderStatusPartiallyTraded   OrderStatus = "PARTIALLY_TRADED"
	OrderStatusCanceled          OrderStatus = "CANCELED"
	OrderStatusPartiallyCanceled OrderStatus = "PARTIALLY_CANCELED"
)

type spotAccountOrdersSub struct {
	streamName string
	onData     func(*PrivateOrder)
	onInvalid  func(error)
}

// NewSpotAccountOrdersSub creates a new subscription for private order updates.
//
// onData is the callback function to handle the private order update.
//
// Panics:
//   - onData is nil
func NewSpotAccountOrdersSub(
	onData func(*PrivateOrder),
) Subscription {
	if onData == nil {
		panic("NewSpotAccountOrdersSub: onData function is nil")
	}
	return &spotAccountOrdersSub{
		streamName: "spot@private.orders.v3.api.pb",
		onData:     onData,
	}
}

// SetOnInvalid sets a callback invoked when the subscription becomes invalid
// (e.g. malformed message, server rejection).
func (s *spotAccountOrdersSub) SetOnInvalid(f func(error)) Subscription {
	s.onInvalid = f
	return s
}

func (s *spotAccountOrdersSub) matches(msg *message) (bool, error) {
	return msg.Msg == s.streamName, nil
}

func (s *spotAccountOrdersSub) acceptEvent(msg *PushDataV3UserWrapper) bool {
	return msg.GetChannel() == s.streamName
}

func (s *spotAccountOrdersSub) handleEvent(msg *PushDataV3UserWrapper) {
	v := msg.GetPrivateOrders()
	if v == nil {
		return
	}

	res, err := mapProtoPrivateOrder(v, msg.SendTime)
	if err != nil {
		if s.onInvalid != nil {
			s.onInvalid(err)
		}
		return
	}

	s.onData(res)
}

func (s *spotAccountOrdersSub) id() string {
	return s.streamName
}

func (s *spotAccountOrdersSub) params() any {
	return s.streamName
}

func parseOrderType(code int32) (OrderType, error) {
	switch code {
	case 1:
		return OrderTypeLimitOrder, nil
	case 2:
		return OrderTypePostOnly, nil
	case 3:
		return OrderTypeImmediateOrCancel, nil
	case 4:
		return OrderTypeFillOrKill, nil
	case 5:
		return OrderTypeMarketOrder, nil
	case 100:
		return OrderTypeStopLossTakeProfit, nil
	default:
		return "", fmt.Errorf("unknown order type: %d", code)
	}
}

func parseOrderStatus(code int32) (OrderStatus, error) {
	switch code {
	case 1:
		return OrderStatusNotTraded, nil
	case 2:
		return OrderStatusFullyTraded, nil
	case 3:
		return OrderStatusPartiallyTraded, nil
	case 4:
		return OrderStatusCanceled, nil
	case 5:
		return OrderStatusPartiallyCanceled, nil
	default:
		return "", fmt.Errorf("unknown order status code: %d", code)
	}
}

func parseTradeSide(code int32) (TradeSide, error) {
	switch code {
	case 1:
		return TradeSideBuy, nil
	case 2:
		return TradeSideSell, nil
	default:
		return "", fmt.Errorf("unknown trade side code: %d", code)
	}
}

// PrivateOrder represents a private order update.
type PrivateOrder struct {
	Id                 string
	ClientId           string // from proto
	Price              decimal.Decimal
	Quantity           decimal.Decimal
	Amount             decimal.Decimal
	AvgPrice           decimal.Decimal
	OrderType          OrderType
	TradeSide          TradeSide
	IsMaker            bool
	RemainAmount       decimal.Decimal
	RemainQuantity     decimal.Decimal
	LastDealQuantity   *decimal.Decimal
	CumulativeQuantity decimal.Decimal
	CumulativeAmount   decimal.Decimal
	Status             OrderStatus
	CreateTime         int64

	// Additional fields from proto
	Market           *string
	TriggerType      *int32
	TriggerPrice     *string
	State            *int32
	OcoId            *string
	RouteFactor      *string
	SymbolId         *string
	MarketId         *string
	MarketCurrencyId *string
	CurrencyId       *string

	SendTime *int64
}

func mapProtoPrivateOrder(msg *PrivateOrdersV3Api, sendTime *int64) (*PrivateOrder, error) {
	parseDecimal := func(fieldName, val string) (decimal.Decimal, error) {
		d, err := decimal.NewFromString(val)
		if err != nil {
			return decimal.Decimal{}, fmt.Errorf("invalid %s %q", fieldName, val)
		}
		return d, nil
	}

	price, err := parseDecimal("price", msg.Price)
	if err != nil {
		return nil, err
	}
	quantity, err := parseDecimal("quantity", msg.Quantity)
	if err != nil {
		return nil, err
	}
	amount, err := parseDecimal("amount", msg.Amount)
	if err != nil {
		return nil, err
	}
	avgPrice, err := parseDecimal("avgPrice", msg.AvgPrice)
	if err != nil {
		return nil, err
	}
	remainAmount, err := parseDecimal("remainAmount", msg.RemainAmount)
	if err != nil {
		return nil, err
	}
	remainQuantity, err := parseDecimal("remainQuantity", msg.RemainQuantity)
	if err != nil {
		return nil, err
	}
	cumulativeQuantity, err := parseDecimal("cumulativeQuantity", msg.CumulativeQuantity)
	if err != nil {
		return nil, err
	}
	cumulativeAmount, err := parseDecimal("cumulativeAmount", msg.CumulativeAmount)
	if err != nil {
		return nil, err
	}

	var lastDealQuantity *decimal.Decimal
	if msg.LastDealQuantity != nil {
		d, err := decimal.NewFromString(*msg.LastDealQuantity)
		if err != nil {
			return nil, fmt.Errorf("invalid lastDealQuantity %q", *msg.LastDealQuantity)
		} else {
			lastDealQuantity = &d
		}
	}

	orderType, err := parseOrderType(msg.OrderType)
	if err != nil {
		return nil, err
	}

	tradeSide, err := parseTradeSide(msg.TradeType)
	if err != nil {
		return nil, err
	}

	status, err := parseOrderStatus(msg.Status)
	if err != nil {
		return nil, err
	}

	return &PrivateOrder{
		Id:                 msg.Id,
		ClientId:           msg.ClientId,
		Price:              price,
		Quantity:           quantity,
		Amount:             amount,
		AvgPrice:           avgPrice,
		OrderType:          orderType,
		TradeSide:          tradeSide,
		IsMaker:            msg.IsMaker,
		RemainAmount:       remainAmount,
		RemainQuantity:     remainQuantity,
		LastDealQuantity:   lastDealQuantity,
		CumulativeQuantity: cumulativeQuantity,
		CumulativeAmount:   cumulativeAmount,
		Status:             status,
		CreateTime:         msg.CreateTime,
		Market:             msg.Market,
		TriggerType:        msg.TriggerType,
		TriggerPrice:       msg.TriggerPrice,
		State:              msg.State,
		OcoId:              msg.OcoId,
		RouteFactor:        msg.RouteFactor,
		SymbolId:           msg.SymbolId,
		MarketId:           msg.MarketId,
		MarketCurrencyId:   msg.MarketCurrencyId,
		CurrencyId:         msg.CurrencyId,
		SendTime:           sendTime,
	}, nil
}
