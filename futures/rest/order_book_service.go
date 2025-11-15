package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/IvanTurko/mexc-sdk-go/internal/httpx"
	"github.com/IvanTurko/mexc-sdk-go/sdkerr"
	"github.com/IvanTurko/mexc-sdk-go/transport"
	"github.com/shopspring/decimal"
)

// OrderBookService gets the order book for a symbol.
type OrderBookService struct {
	client     transport.HTTPClient
	reqBuilder *requestBuilder
	symbol     string
	limit      *int
}

// NewOrderBookService creates a new OrderBookService.
func NewOrderBookService() *OrderBookService {
	return &OrderBookService{
		client:     httpx.NewDefaultHTTPClient(),
		reqBuilder: newRequestBuilder("", ""),
	}
}

// WithClient sets the HTTP client for the service.
func (s *OrderBookService) WithClient(client transport.HTTPClient) *OrderBookService {
	s.client = client
	return s
}

// Symbol sets the symbol for the order book.
func (s *OrderBookService) Symbol(symbol string) *OrderBookService {
	s.symbol = symbol
	return s
}

// Limit sets the limit for the order book depths.
func (s *OrderBookService) Limit(n int) *OrderBookService {
	s.limit = &n
	return s
}

// Validate validates the service parameters.
func (s *OrderBookService) Validate() error {
	if err := s.validate(); err != nil {
		return sdkerr.NewSDKError().
			WithSubsys(subsys).
			WithOp("OrderBookService.Validate").
			WithKind(sdkerr.ErrValidation).
			WithMessage(err.Error())
	}
	return nil
}

// Do executes the service.
func (s *OrderBookService) Do(ctx context.Context) (*OrderBookDepths, error) {
	path := fmt.Sprintf("/api/v1/contract/depth/%s", s.symbol)
	req := s.reqBuilder.
		WithMethod(http.MethodGet).
		WithPath(path).
		WithQuery(s.buildQuery()).
		Build()

	op := "OrderBookService.Do"
	resp, err := s.client.Do(ctx, req)
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

	return decodeResponse[OrderBookDepths](resp.Body, op)
}

func (s *OrderBookService) validate() error {
	var errs []string
	if s.symbol == "" {
		errs = append(errs, "symbol is required")
	}
	if s.limit != nil && *s.limit < 1 {
		errs = append(errs, "limit must be above 0")
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func (s *OrderBookService) buildQuery() url.Values {
	q := make(url.Values)
	if s.limit != nil {
		q.Add("limit", strconv.Itoa(*s.limit))
	}
	return q
}

// OrderBookDepths represents the order book depths.
type OrderBookDepths struct {
	Asks      []BookLevel `json:"asks"`
	Bids      []BookLevel `json:"bids"`
	Version   int64       `json:"version"`
	Timestamp int64       `json:"timestamp"`
}

// BookLevel represents a level in the order book.
type BookLevel struct {
	Price    decimal.Decimal
	Quantity decimal.Decimal
	Orders   decimal.Decimal
}

func (b *BookLevel) UnmarshalJSON(data []byte) error {
	var tmp [3]json.Number
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	price, err := decimal.NewFromString(tmp[0].String())
	if err != nil {
		return err
	}

	quantity, err := decimal.NewFromString(tmp[1].String())
	if err != nil {
		return err
	}

	orders, err := decimal.NewFromString(tmp[2].String())
	if err != nil {
		return err
	}

	b.Price = price
	b.Quantity = quantity
	b.Orders = orders
	return nil
}
