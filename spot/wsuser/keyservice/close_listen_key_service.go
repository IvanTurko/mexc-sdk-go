package keyservice

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/IvanTurko/mexc-sdk-go/internal/httpx"
	"github.com/IvanTurko/mexc-sdk-go/internal/signature"
	"github.com/IvanTurko/mexc-sdk-go/internal/timeutil"
	"github.com/IvanTurko/mexc-sdk-go/sdkerr"
	"github.com/IvanTurko/mexc-sdk-go/transport"
)

// CloseListenKeyService closes a listen key.
type CloseListenKeyService struct {
	client     transport.HTTPClient
	reqBuilder *requestBuilder
	secretKey  string
	listenKey  string
	recvWindow *int64
	timestamp  func() int64
}

// NewCloseListenKeyService creates a new CloseListenKeyService.
func NewCloseListenKeyService(apiKey, secretKey string) *CloseListenKeyService {
	return &CloseListenKeyService{
		client:     httpx.NewDefaultHTTPClient(),
		reqBuilder: newRequestBuilder(apiKey),
		timestamp:  timeutil.NowMillis,
		secretKey:  secretKey,
	}
}

// WithClient sets the HTTP client for the service.
func (c *CloseListenKeyService) WithClient(client transport.HTTPClient) *CloseListenKeyService {
	c.client = client
	return c
}

// ListenKey sets the listen key to be closed.
func (c *CloseListenKeyService) ListenKey(key string) *CloseListenKeyService {
	c.listenKey = key
	return c
}

// RecvWindow sets the receive window for the request.
func (c *CloseListenKeyService) RecvWindow(ms int64) *CloseListenKeyService {
	c.recvWindow = &ms
	return c
}

// Validate validates the service parameters.
func (c *CloseListenKeyService) Validate() error {
	if err := c.validate(); err != nil {
		return sdkerr.NewSDKError().
			WithSubsys(subsys).
			WithOp("CloseListenKeyService.Validate").
			WithKind(sdkerr.ErrValidation).
			WithMessage(err.Error())
	}
	return nil
}

// Do executes the service.
func (c *CloseListenKeyService) Do(ctx context.Context) (string, error) {
	req := c.reqBuilder.
		WithMethod(http.MethodDelete).
		WithPath("/api/v3/userDataStream").
		WithQuery(c.buildQuery()).
		Build()

	op := "CloseListenKeyService.Do"
	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return "", sdkerr.NewSDKError().
			WithSubsys(subsys).
			WithOp(op).
			WithKind(sdkerr.ErrRequestFailed).
			WithCause(err)
	}

	if err := checkResponseError(resp.StatusCode, resp.Body); err != nil {
		return "", sdkerr.NewSDKError().
			WithSubsys(subsys).
			WithOp(op).
			WithKind(sdkerr.ErrAPIError).
			WithCause(err)
	}

	respObj, err := decodeResponse[listenKeySingle](resp.Body, op)
	if err != nil {
		return "", err
	}
	return respObj.ListenKey, nil
}

func (c *CloseListenKeyService) validate() error {
	if c.listenKey == "" {
		return errors.New("listenKey is required")
	}
	return nil
}

func (c *CloseListenKeyService) buildQuery() url.Values {
	q := make(url.Values)

	q.Add("listenKey", c.listenKey)

	if c.recvWindow != nil {
		q.Add("recvWindow", strconv.FormatInt(*c.recvWindow, 10))
	}

	q.Add("timestamp", strconv.FormatInt(c.timestamp(), 10))

	queryStr := q.Encode()
	sig := signature.HMACSHA256(queryStr, c.secretKey)
	q.Add("signature", sig)

	return q
}
