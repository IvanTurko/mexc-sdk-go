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

// KeepAliveListenKeyService keeps a listen key alive.
type KeepAliveListenKeyService struct {
	client     transport.HTTPClient
	reqBuilder *requestBuilder
	secretKey  string
	listenKey  string
	recvWindow *int64
	timestamp  func() int64
}

// NewKeepAliveListenKeyService creates a new KeepAliveListenKeyService.
func NewKeepAliveListenKeyService(apiKey, secretKey string) *KeepAliveListenKeyService {
	return &KeepAliveListenKeyService{
		client:     httpx.NewDefaultHTTPClient(),
		reqBuilder: newRequestBuilder(apiKey),
		timestamp:  timeutil.NowMillis,
		secretKey:  secretKey,
	}
}

// WithClient sets the HTTP client for the service.
func (k *KeepAliveListenKeyService) WithClient(client transport.HTTPClient) *KeepAliveListenKeyService {
	k.client = client
	return k
}

// ListenKey sets the listen key to be kept alive.
func (k *KeepAliveListenKeyService) ListenKey(key string) *KeepAliveListenKeyService {
	k.listenKey = key
	return k
}

// RecvWindow sets the receive window for the request.
func (k *KeepAliveListenKeyService) RecvWindow(ms int64) *KeepAliveListenKeyService {
	k.recvWindow = &ms
	return k
}

// Validate validates the service parameters.
func (k *KeepAliveListenKeyService) Validate() error {
	if err := k.validate(); err != nil {
		return sdkerr.NewSDKError().
			WithSubsys(subsys).
			WithOp("KeepAliveListenKeyService.Validate").
			WithKind(sdkerr.ErrValidation).
			WithMessage(err.Error())
	}
	return nil
}

// Do executes the service.
func (k *KeepAliveListenKeyService) Do(ctx context.Context) (string, error) {
	req := k.reqBuilder.
		WithMethod(http.MethodPut).
		WithPath("/api/v3/userDataStream").
		WithQuery(k.buildQuery()).
		Build()

	op := "KeepAliveListenKeyService.Do"
	resp, err := k.client.Do(ctx, req)
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

func (k *KeepAliveListenKeyService) validate() error {
	if k.listenKey == "" {
		return errors.New("listenKey is required")
	}
	return nil
}

func (k *KeepAliveListenKeyService) buildQuery() url.Values {
	q := make(url.Values)

	q.Add("listenKey", k.listenKey)

	if k.recvWindow != nil {
		q.Add("recvWindow", strconv.FormatInt(*k.recvWindow, 10))
	}

	q.Add("timestamp", strconv.FormatInt(k.timestamp(), 10))

	queryStr := q.Encode()
	sig := signature.HMACSHA256(queryStr, k.secretKey)
	q.Add("signature", sig)

	return q
}
