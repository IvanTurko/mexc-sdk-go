package wsmarket

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

// TradeSide indicates whether the trade was buyer or seller initiated.
type TradeSide string

const (
	TradeSideBuy  TradeSide = "BUY"
	TradeSideSell TradeSide = "SELL"
)

// OpenType indicates whether the trade was an open or close position.
type OpenType string

const (
	OpenTypeOpen     OpenType = "OPEN"
	OpenTypeClose    OpenType = "CLOSE"
	OpenTypeNoChange OpenType = "NO_CHANGE"
)

// AutoTransact indicates whether the trade was an auto-transact.
type AutoTransact string

const (
	AutoTransactYes AutoTransact = "YES"
	AutoTransactNo  AutoTransact = "NO"
)

type tradeStreamsSub struct {
	symbol    string
	onInvalid func(error)
	onData    func(Trade)
}

// NewTradeStreamsSub creates a subscription for trade streams.
//
// symbol is the trading pair, e.g. "BTC_USDT".
// onData is called for every received Trade.
//
// Panics:
//   - symbol is empty
//   - onData is nil
func NewTradeStreamsSub(
	symbol string,
	onData func(Trade),
) Subscription {
	if symbol == "" {
		panic("NewTradeStreamsSub: invalid symbol name")
	}
	if onData == nil {
		panic("NewTradeStreamsSub: onData function is nil")
	}

	return &tradeStreamsSub{
		symbol: symbol,
		onData: onData,
	}
}

// SetOnInvalid registers a callback that is triggered if the subscription
// becomes invalid (invalid payload, server rejection, etc.).
func (t *tradeStreamsSub) SetOnInvalid(f func(error)) Subscription {
	t.onInvalid = f
	return t
}

func (t *tradeStreamsSub) matches(msg *message) (bool, error) {
	if msg.Channel == "rs.sub."+t.channel() {
		var s string
		if err := json.Unmarshal(msg.Data, &s); err != nil {
			return false, fmt.Errorf("invalid success payload: %s", string(msg.Data))
		}
		if s == "success" {
			return true, nil
		}
	}

	if msg.Channel == "rs.error" {
		return true, fmt.Errorf("sub failed: %s", string(msg.Data))
	}

	return false, nil
}

func (t *tradeStreamsSub) acceptEvent(msg *message) bool {
	return msg.Channel == "push."+t.channel() && msg.Symbol == t.symbol
}

func (t *tradeStreamsSub) handleEvent(msg *message) {
	var deal Trade

	if err := json.Unmarshal(msg.Data, &deal); err != nil {
		err = fmt.Errorf("failed to unmarshal data: %v, raw: %s", err, string(msg.Data))
		if t.onInvalid != nil {
			t.onInvalid(err)
		}
		return
	}

	deal.Symbol = msg.Symbol
	deal.SendTime = &msg.Ts
	t.onData(deal)
}

func (t *tradeStreamsSub) id() string {
	return fmt.Sprintf("%s@%s", t.channel(), t.symbol)
}

func (t *tradeStreamsSub) channel() string {
	return "deal"
}

func (t *tradeStreamsSub) payload(op subscriptionOp) any {
	return wsRequestPayload{
		Method: fmt.Sprintf("%s.%s", op, t.channel()),
		Param: map[string]any{
			"symbol": t.symbol,
		},
	}
}

// Trade represents a single executed trade.
type Trade struct {
	Symbol         string
	Price          decimal.Decimal
	Volume         decimal.Decimal
	Side           TradeSide
	OpenType       OpenType
	IsAutoTransact AutoTransact
	Timestamp      int64
	SendTime       *int64
}

func (t *Trade) UnmarshalJSON(data []byte) error {
	var tmp tradeJSON

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	side, err := parseTradeSide(tmp.Side)
	if err != nil {
		return err
	}
	openType, err := parseOpenType(tmp.OpenType)
	if err != nil {
		return err
	}
	autoTransact, err := parseAutoTransact(tmp.AutoTransact)
	if err != nil {
		return err
	}

	t.Price = tmp.Price
	t.Volume = tmp.Volume
	t.Side = side
	t.OpenType = openType
	t.IsAutoTransact = autoTransact
	t.Timestamp = tmp.Timestamp

	return nil
}

type tradeJSON struct {
	Price        decimal.Decimal `json:"p"`
	Volume       decimal.Decimal `json:"v"`
	Side         int             `json:"T"`
	OpenType     int             `json:"O"`
	AutoTransact int             `json:"M"`
	Timestamp    int64           `json:"t"`
}

func parseTradeSide(code int) (TradeSide, error) {
	switch code {
	case 1:
		return TradeSideBuy, nil
	case 2:
		return TradeSideSell, nil
	default:
		return "", fmt.Errorf("unknown trade side code: %d", code)
	}
}

func parseOpenType(code int) (OpenType, error) {
	switch code {
	case 1:
		return OpenTypeOpen, nil
	case 2:
		return OpenTypeClose, nil
	case 3:
		return OpenTypeNoChange, nil
	default:
		return "", fmt.Errorf("unknown open type code: %d", code)
	}
}

func parseAutoTransact(code int) (AutoTransact, error) {
	switch code {
	case 1:
		return AutoTransactYes, nil
	case 2:
		return AutoTransactNo, nil
	default:
		return "", fmt.Errorf("unknown auto transact code: %d", code)
	}
}
