package rest

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/IvanTurko/mexc-sdk-go/internal/httpx"
	"github.com/IvanTurko/mexc-sdk-go/internal/signature"
	"github.com/IvanTurko/mexc-sdk-go/internal/timeutil"
	"github.com/IvanTurko/mexc-sdk-go/sdkerr"
	"github.com/IvanTurko/mexc-sdk-go/transport"
	"github.com/shopspring/decimal"
)

// PlacedOrder represents a successfully placed order.
type PlacedOrder struct {
	Symbol       string          `json:"symbol"`
	OrderID      string          `json:"orderId"`
	OrderListID  int64           `json:"orderListId"`
	Price        decimal.Decimal `json:"price"`
	OrigQty      decimal.Decimal `json:"origQty"`
	Type         OrderType       `json:"type"`
	Side         OrderSide       `json:"side"`
	TransactTime int64           `json:"transactTime"`
}

// CreateOrderService creates a new order.
type CreateOrderService struct {
	secretKey  string
	client     transport.HTTPClient
	reqBuilder *requestBuilder

	symbol           string
	side             OrderSide
	_type            OrderType
	quantity         *decimal.Decimal
	quoteOrderQty    *decimal.Decimal
	price            *decimal.Decimal
	newClientOrderId *string
	recvWindow       *int64
	timestamp        func() int64
}

// NewCreateOrderService creates a new CreateOrderService.
func NewCreateOrderService(apiKey, secretKey string) *CreateOrderService {
	return &CreateOrderService{
		client:     httpx.NewDefaultHTTPClient(),
		reqBuilder: newRequestBuilder(apiKey),
		timestamp:  timeutil.NowMillis,
		secretKey:  secretKey,
	}
}

// WithClient sets the HTTP client for the service.
func (c *CreateOrderService) WithClient(client transport.HTTPClient) *CreateOrderService {
	c.client = client
	return c
}

// Symbol sets the symbol for the order.
func (c *CreateOrderService) Symbol(symbol string) *CreateOrderService {
	c.symbol = symbol
	return c
}

// Side sets the side of the order.
func (c *CreateOrderService) Side(side OrderSide) *CreateOrderService {
	c.side = side
	return c
}

// Type sets the type of the order.
func (c *CreateOrderService) Type(orderType OrderType) *CreateOrderService {
	c._type = orderType
	return c
}

// Quantity sets the quantity of the order.
func (c *CreateOrderService) Quantity(quantity decimal.Decimal) *CreateOrderService {
	c.quantity = &quantity
	return c
}

// QuoteOrderQty sets the quote order quantity of the order.
func (c *CreateOrderService) QuoteOrderQty(qty decimal.Decimal) *CreateOrderService {
	c.quoteOrderQty = &qty
	return c
}

// Price sets the price of the order.
func (c *CreateOrderService) Price(price decimal.Decimal) *CreateOrderService {
	c.price = &price
	return c
}

// NewClientOrderId sets the new client order ID of the order.
func (c *CreateOrderService) NewClientOrderId(id string) *CreateOrderService {
	c.newClientOrderId = &id
	return c
}

// RecvWindow sets the receive window for the request.
func (c *CreateOrderService) RecvWindow(ms int64) *CreateOrderService {
	c.recvWindow = &ms
	return c
}

// Validate validates the service parameters.
func (c *CreateOrderService) Validate() error {
	if err := c.validate(); err != nil {
		return sdkerr.NewSDKError().
			WithSubsys(subsys).
			WithOp("CreateOrderService.Validate").
			WithKind(sdkerr.ErrValidation).
			WithMessage(err.Error())
	}
	return nil
}

// Do executes the service.
func (c *CreateOrderService) Do(ctx context.Context) (*PlacedOrder, error) {
	req := c.reqBuilder.
		WithMethod(http.MethodPost).
		WithPath("/api/v3/order").
		WithQuery(c.buildQuery()).
		Build()

	op := "CreateOrderService.Do"
	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return nil, sdkerr.NewSDKError().
			WithSubsys(subsys).
			WithOp(op).
			WithKind(sdkerr.ErrRequestFailed).
			WithCause(err)
	}

	if err := checkResponseError(resp.StatusCode, resp.Body); err != nil {
		return nil, sdkerr.NewSDKError().
			WithSubsys(subsys).
			WithOp(op).
			WithKind(sdkerr.ErrAPIError).
			WithCause(err)
	}

	return decodeResponse[PlacedOrder](resp.Body, op)
}

func (c *CreateOrderService) validate() error {
	var errs []string

	if c.symbol == "" {
		errs = append(errs, "symbol is required")
	}

	if !c.side.isValid() {
		errs = append(errs, "side is invalid")
	}

	if !c._type.isValid() {
		errs = append(errs, "type is invalid")
	}

	if c.quantity != nil && c.quantity.Cmp(decimal.Zero) <= 0 {
		errs = append(errs, "quantity must be greater than zero")
	}

	if c.quoteOrderQty != nil && c.quoteOrderQty.Cmp(decimal.Zero) <= 0 {
		errs = append(errs, "quoteOrderQty must be greater than zero")
	}

	if c.price != nil && c.price.Cmp(decimal.Zero) <= 0 {
		errs = append(errs, "price must be greater than zero")
	}

	if c.recvWindow != nil {
		if *c.recvWindow < 1 || *c.recvWindow > 60000 {
			errs = append(errs, "recvWindow must be between 1 and 60000")
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func (c *CreateOrderService) buildQuery() url.Values {
	q := make(url.Values)

	q.Add("symbol", c.symbol)
	q.Add("side", string(c.side))
	q.Add("type", string(c._type))

	if c.quantity != nil {
		q.Add("quantity", c.quantity.String())
	}
	if c.quoteOrderQty != nil {
		q.Add("quoteOrderQty", c.quoteOrderQty.String())
	}
	if c.price != nil {
		q.Add("price", c.price.String())
	}
	if c.newClientOrderId != nil {
		q.Add("newClientOrderId", *c.newClientOrderId)
	}
	if c.recvWindow != nil {
		q.Add("recvWindow", strconv.FormatInt(*c.recvWindow, 10))
	}

	q.Add("timestamp", strconv.FormatInt(c.timestamp(), 10))

	queryStr := q.Encode()
	sig := signature.HMACSHA256(queryStr, c.secretKey)
	q.Add("signature", sig)

	return q
}
