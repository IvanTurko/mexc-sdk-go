package wsuser

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

// OrderSide represents the side of an order.
type OrderSide string

const (
	OrderSideOpenLong   OrderSide = "OPEN_LONG"
	OrderSideCloseShort OrderSide = "CLOSE_SHORT"
	OrderSideOpenShort  OrderSide = "OPEN_SHORT"
	OrderSideCloseLong  OrderSide = "CLOSE_LONG"
)

// OrderType represents the type of an order.
type OrderType string

const (
	OrderTypeLimit             OrderType = "LIMIT"
	OrderTypeMarket            OrderType = "MARKET"
	OrderTypeLimitMaker        OrderType = "LIMIT_MAKER"
	OrderTypeImmediateOrCancel OrderType = "IMMEDIATE_OR_CANCEL"
	OrderTypeFillOrKill        OrderType = "FILL_OR_KILL"
	OrderTypeMarketToLimit     OrderType = "MARKET_TO_LIMIT"
)

// OrderCategory represents the category of an order.
type OrderCategory string

const (
	OrderCategoryLimitOrder             OrderCategory = "LIMIT_ORDER"
	OrderCategorySystemTakeOverDelegate OrderCategory = "SYSTEM_TAKE_OVER_DELEGATE"
	OrderCategoryCloseDelegate          OrderCategory = "CLOSE_DELEGATE"
	OrderCategoryAdlReduction           OrderCategory = "ADL_REDUCTION"
)

// OrderState represents the state of an order.
type OrderState string

const (
	OrderStateUninformed  OrderState = "UNINFORMED"
	OrderStateUncompleted OrderState = "UNCOMPLETED"
	OrderStateCompleted   OrderState = "COMPLETED"
	OrderStateCancelled   OrderState = "CANCELLED"
	OrderStateInvalid     OrderState = "INVALID"
)

// OrderErrorCode represents the error code of an order.
type OrderErrorCode string

const (
	OrderErrorCodeNormal                 OrderErrorCode = "NORMAL"
	OrderErrorCodeParamInvalid           OrderErrorCode = "PARAM_INVALID"
	OrderErrorCodeInsufficientBalance    OrderErrorCode = "INSUFFICIENT_BALANCE"
	OrderErrorCodePositionNotExists      OrderErrorCode = "POSITION_NOT_EXISTS"
	OrderErrorCodePositionNotEnough      OrderErrorCode = "POSITION_NOT_ENOUGH"
	OrderErrorCodePositionLiq            OrderErrorCode = "POSITION_LIQ"
	OrderErrorCodeOrderLiq               OrderErrorCode = "ORDER_LIQ"
	OrderErrorCodeRiskLevelLimit         OrderErrorCode = "RISK_LEVEL_LIMIT"
	OrderErrorCodeSysCancel              OrderErrorCode = "SYS_CANCEL"
	OrderErrorCodePositionModeNotMatch   OrderErrorCode = "POSITION_MODE_NOT_MATCH"
	OrderErrorCodeReduceOnlyLiq          OrderErrorCode = "REDUCE_ONLY_LIQ"
	OrderErrorCodeContractNotEnable      OrderErrorCode = "CONTRACT_NOT_ENABLE"
	OrderErrorCodeDeliveryCancel         OrderErrorCode = "DELIVERY_CANCEL"
	OrderErrorCodePositionLiqCancel      OrderErrorCode = "POSITION_LIQ_CANCEL"
	OrderErrorCodeAdlCancel              OrderErrorCode = "ADL_CANCEL"
	OrderErrorCodeBlackUserCancel        OrderErrorCode = "BLACK_USER_CANCEL"
	OrderErrorCodeSettleFundingCancel    OrderErrorCode = "SETTLE_FUNDING_CANCEL"
	OrderErrorCodePositionImChangeCancel OrderErrorCode = "POSITION_IM_CHANGE_CANCEL"
	OrderErrorCodeIocCancel              OrderErrorCode = "IOC_CANCEL"
	OrderErrorCodeFokCancel              OrderErrorCode = "FOK_CANCEL"
	OrderErrorCodePostOnlyCancel         OrderErrorCode = "POST_ONLY_CANCEL"
	OrderErrorCodeMarketCancel           OrderErrorCode = "MARKET_CANCEL"
)

type orderSub struct {
	onInvalid func(error)
	onData    func(*Order)
}

// NewOrderSub creates a new subscription for order updates.
//
// onData is the callback function to handle the order update.
//
// Panics:
//   - onData is nil
func NewOrderSub(
	onData func(*Order),
) Subscription {
	if onData == nil {
		panic("NewOrderSub: onData function is nil")
	}

	return &orderSub{
		onData: onData,
	}
}

// SetOnInvalid sets a callback invoked when the subscription becomes invalid
// (e.g. malformed message, server rejection).
func (s *orderSub) SetOnInvalid(f func(error)) Subscription {
	s.onInvalid = f
	return s
}

func (s *orderSub) acceptEvent(msg *message) bool {
	return msg.Channel == "push.personal."+s.channel()
}

func (s *orderSub) handleEvent(msg *message) {
	var event *Order

	if err := json.Unmarshal(msg.Data, &event); err != nil {
		err = fmt.Errorf("failed to unmarshal data: %v, raw: %s", err, string(msg.Data))
		if s.onInvalid != nil {
			s.onInvalid(err)
		}
		return
	}

	event.SendTime = &msg.Ts
	s.onData(event)
}

func (s *orderSub) id() string {
	return s.channel()
}

func (s *orderSub) channel() string {
	return "order"
}

// Order represents an order update.
type Order struct {
	OrderId      string
	Symbol       string
	PositionId   int64
	Price        decimal.Decimal
	Vol          decimal.Decimal
	Leverage     decimal.Decimal
	Side         OrderSide
	Category     OrderCategory
	OrderType    OrderType
	DealAvgPrice decimal.Decimal
	DealVol      decimal.Decimal
	OrderMargin  decimal.Decimal
	UsedMargin   decimal.Decimal
	TakerFee     decimal.Decimal
	MakerFee     decimal.Decimal
	Profit       decimal.Decimal
	FeeCurrency  string
	OpenType     OpenType
	State        OrderState
	ErrorCode    OrderErrorCode
	ExternalOid  string
	CreateTime   int64
	UpdateTime   int64
	SendTime     *int64
}

func (s *Order) UnmarshalJSON(data []byte) error {
	var tmp orderJSON

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	side, err := parseOrderSide(tmp.Side)
	if err != nil {
		return err
	}
	category, err := parseOrderCategory(tmp.Category)
	if err != nil {
		return err
	}
	orderType, err := parseOrderType(tmp.OrderType)
	if err != nil {
		return err
	}
	openType, err := parseOpenType(tmp.OpenType)
	if err != nil {
		return err
	}
	state, err := parseOrderState(tmp.State)
	if err != nil {
		return err
	}
	errorCode, err := parseOrderErrorCode(tmp.ErrorCode)
	if err != nil {
		return err
	}

	s.OrderId = tmp.OrderId
	s.Symbol = tmp.Symbol
	s.PositionId = tmp.PositionId
	s.Price = tmp.Price
	s.Vol = tmp.Vol
	s.Leverage = tmp.Leverage
	s.Side = side
	s.Category = category
	s.OrderType = orderType
	s.DealAvgPrice = tmp.DealAvgPrice
	s.DealVol = tmp.DealVol
	s.OrderMargin = tmp.OrderMargin
	s.UsedMargin = tmp.UsedMargin
	s.TakerFee = tmp.TakerFee
	s.MakerFee = tmp.MakerFee
	s.Profit = tmp.Profit
	s.FeeCurrency = tmp.FeeCurrency
	s.OpenType = openType
	s.State = state
	s.ErrorCode = errorCode
	s.ExternalOid = tmp.ExternalOid
	s.CreateTime = tmp.CreateTime
	s.UpdateTime = tmp.UpdateTime

	return nil
}

type orderJSON struct {
	OrderId      string          `json:"orderId"`
	Symbol       string          `json:"symbol"`
	PositionId   int64           `json:"positionId"`
	Price        decimal.Decimal `json:"price"`
	Vol          decimal.Decimal `json:"vol"`
	Leverage     decimal.Decimal `json:"leverage"`
	Side         int             `json:"side"`
	Category     int             `json:"category"`
	OrderType    int             `json:"orderType"`
	DealAvgPrice decimal.Decimal `json:"dealAvgPrice"`
	DealVol      decimal.Decimal `json:"dealVol"`
	OrderMargin  decimal.Decimal `json:"orderMargin"`
	UsedMargin   decimal.Decimal `json:"usedMargin"`
	TakerFee     decimal.Decimal `json:"takerFee"`
	MakerFee     decimal.Decimal `json:"makerFee"`
	Profit       decimal.Decimal `json:"profit"`
	FeeCurrency  string          `json:"feeCurrency"`
	OpenType     int             `json:"openType"`
	State        int             `json:"state"`
	ErrorCode    int             `json:"errorCode"`
	ExternalOid  string          `json:"externalOid"`
	CreateTime   int64           `json:"createTime"`
	UpdateTime   int64           `json:"updateTime"`
}

func parseOrderSide(code int) (OrderSide, error) {
	switch code {
	case 1:
		return OrderSideOpenLong, nil
	case 2:
		return OrderSideCloseShort, nil
	case 3:
		return OrderSideOpenShort, nil
	case 4:
		return OrderSideCloseLong, nil
	default:
		return "", fmt.Errorf("unknown order side code: %d", code)
	}
}

func parseOrderType(code int) (OrderType, error) {
	switch code {
	case 1:
		return OrderTypeLimit, nil
	case 2:
		return OrderTypeLimitMaker, nil
	case 3:
		return OrderTypeImmediateOrCancel, nil
	case 4:
		return OrderTypeFillOrKill, nil
	case 5:
		return OrderTypeMarket, nil
	case 6:
		return OrderTypeMarketToLimit, nil
	default:
		return "", fmt.Errorf("unknown order type code: %d", code)
	}
}

func parseOrderCategory(code int) (OrderCategory, error) {
	switch code {
	case 1:
		return OrderCategoryLimitOrder, nil
	case 2:
		return OrderCategorySystemTakeOverDelegate, nil
	case 3:
		return OrderCategoryCloseDelegate, nil
	case 4:
		return OrderCategoryAdlReduction, nil
	default:
		return "", fmt.Errorf("unknown order category code: %d", code)
	}
}



func parseOrderState(code int) (OrderState, error) {
	switch code {
	case 1:
		return OrderStateUninformed, nil
	case 2:
		return OrderStateUncompleted, nil
	case 3:
		return OrderStateCompleted, nil
	case 4:
		return OrderStateCancelled, nil
	case 5:
		return OrderStateInvalid, nil
	default:
		return "", fmt.Errorf("unknown order state code: %d", code)
	}
}

var orderErrorCodeMap = map[int]OrderErrorCode{
	0:  OrderErrorCodeNormal,
	1:  OrderErrorCodeParamInvalid,
	2:  OrderErrorCodeInsufficientBalance,
	3:  OrderErrorCodePositionNotExists,
	4:  OrderErrorCodePositionNotEnough,
	5:  OrderErrorCodePositionLiq,
	6:  OrderErrorCodeOrderLiq,
	7:  OrderErrorCodeRiskLevelLimit,
	8:  OrderErrorCodeSysCancel,
	9:  OrderErrorCodePositionModeNotMatch,
	10: OrderErrorCodeReduceOnlyLiq,
	11: OrderErrorCodeContractNotEnable,
	12: OrderErrorCodeDeliveryCancel,
	13: OrderErrorCodePositionLiqCancel,
	14: OrderErrorCodeAdlCancel,
	15: OrderErrorCodeBlackUserCancel,
	16: OrderErrorCodeSettleFundingCancel,
	17: OrderErrorCodePositionImChangeCancel,
	18: OrderErrorCodeIocCancel,
	19: OrderErrorCodeFokCancel,
	20: OrderErrorCodePostOnlyCancel,
	21: OrderErrorCodeMarketCancel,
}

func parseOrderErrorCode(code int) (OrderErrorCode, error) {
	if val, ok := orderErrorCodeMap[code]; ok {
		return val, nil
	}
	return "", fmt.Errorf("unknown order error code: %d", code)
}
