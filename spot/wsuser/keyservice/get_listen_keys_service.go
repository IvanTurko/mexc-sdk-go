package keyservice

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/IvanTurko/mexc-sdk-go/internal/httpx"
	"github.com/IvanTurko/mexc-sdk-go/internal/signature"
	"github.com/IvanTurko/mexc-sdk-go/internal/timeutil"
	"github.com/IvanTurko/mexc-sdk-go/sdkerr"
	"github.com/IvanTurko/mexc-sdk-go/transport"
)

// GetListenKeysService gets listen keys.
type GetListenKeysService struct {
	client     transport.HTTPClient
	reqBuilder *requestBuilder
	secretKey  string
	recvWindow *int64
	timestamp  func() int64
}

// NewGetListenKeysService creates a new GetListenKeysService.
func NewGetListenKeysService(apiKey, secretKey string) *GetListenKeysService {
	return &GetListenKeysService{
		client:     httpx.NewDefaultHTTPClient(),
		reqBuilder: newRequestBuilder(apiKey),
		timestamp:  timeutil.NowMillis,
		secretKey:  secretKey,
	}
}

// WithClient sets the HTTP client for the service.
func (g *GetListenKeysService) WithClient(client transport.HTTPClient) *GetListenKeysService {
	g.client = client
	return g
}

// RecvWindow sets the receive window for the request.
func (g *GetListenKeysService) RecvWindow(ms int64) *GetListenKeysService {
	g.recvWindow = &ms
	return g
}

// Do executes the service.
func (g *GetListenKeysService) Do(ctx context.Context) (*ListenKeys, error) {
	req := g.reqBuilder.
		WithMethod(http.MethodGet).
		WithPath("/api/v3/userDataStream").
		WithQuery(g.buildQuery()).
		Build()

	op := "GetListenKeysService.Do"
	resp, err := g.client.Do(ctx, req)
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

	return decodeResponse[ListenKeys](resp.Body, op)
}

func (g *GetListenKeysService) buildQuery() url.Values {
	q := make(url.Values)

	if g.recvWindow != nil {
		q.Add("recvWindow", strconv.FormatInt(*g.recvWindow, 10))
	}

	q.Add("timestamp", strconv.FormatInt(g.timestamp(), 10))

	queryStr := q.Encode()
	sig := signature.HMACSHA256(queryStr, g.secretKey)
	q.Add("signature", sig)

	return q
}
