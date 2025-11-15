package wsmarket

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/IvanTurko/mexc-sdk-go/internal/wsutil"
	"github.com/IvanTurko/mexc-sdk-go/sdkerr"
	"github.com/IvanTurko/mexc-sdk-go/ws"
)

const (
	subsys             = "futures/wsmarket"
	defaultBaseURL     = "wss://contract.mexc.com/edge"
	maxCountSubscribes = 30
)

var (
	// ErrDuplicateSubscription is returned when a subscription with the same ID already exists.
	ErrDuplicateSubscription = errors.New("duplicate subscription")
	// ErrMaxCountSubscribes is returned when the maximum number of subscriptions is reached.
	ErrMaxCountSubscribes = errors.New("maximum number of subscribers exceeded")
)

// Options is a function type for WSMarket options.
type Options = func(*WSMarket)

// Subscription is an interface for a websocket subscription.
type Subscription interface {
	// SetOnInvalid registers a callback fired when the subscription becomes invalid
	// (malformed message, server rejection, etc.).
	SetOnInvalid(f func(error)) Subscription
	subscriptionSpec
}

// SubscriptionHandle is an interface to handle a subscription, e.g., to unsubscribe.
type SubscriptionHandle interface {
	Unsubscribe(ctx context.Context) error
}

// WSMarket is a WebSocket client for MEXC FUTURES market streams.
type WSMarket struct {
	client         ws.Client
	waitingTimeout time.Duration

	internalTimeout time.Duration
	pingInterval    time.Duration
	now             func() time.Time
	onDisconnect    func(err error)
	onLatency       func(latency time.Duration)

	activeSubs   map[string]SubscriptionHandle
	activeSubsMu sync.Mutex

	promisesMu  sync.Mutex
	promise     wsutil.Promise[message]
	promiseFunc func(matchFn func(*message) (bool, error)) wsutil.Promise[message]

	router    handlerRouter
	closeOnce sync.Once
	closeErr  error
}

// NewWSMarket creates a WSMarket using the default WebSocket client.
// Additional configuration can be supplied through Options.
func NewWSMarket(opts ...Options) *WSMarket {
	factory := func(url string) ws.Client {
		return ws.NewClient(url)
	}
	return NewWSMarketWithFactory(factory, opts...)
}

// NewWSMarketWithFactory is like NewWSMarket but uses the provided ws.Client factory.
// Useful for tests and custom connection setups.
func NewWSMarketWithFactory(factory func(url string) ws.Client, opts ...Options) *WSMarket {
	if factory == nil {
		panic("NewWSMarketWithFactory: factory must not be nil")
	}

	w := &WSMarket{
		client:         factory(defaultBaseURL),
		waitingTimeout: 1 * time.Second,

		internalTimeout: 1 * time.Second,
		pingInterval:    20 * time.Second,
		now:             time.Now,

		router:     newHandlerRouter(),
		activeSubs: make(map[string]SubscriptionHandle),

		promiseFunc: wsutil.NewPromise[message],
	}

	for _, opt := range opts {
		opt(w)
	}

	return w
}

// WithOnDisconnect registers a callback for unexpected disconnections.
func WithOnDisconnect(f func(err error)) Options {
	return func(w *WSMarket) {
		w.onDisconnect = f
	}
}

// WithPingLatencyHandler registers a callback receiving RTT ping/pong measurements.
func WithPingLatencyHandler(f func(time.Duration)) Options {
	return func(w *WSMarket) {
		w.onLatency = f
	}
}

// Connect opens the WebSocket connection and starts internal workers.
func (w *WSMarket) Connect(ctx context.Context) error {
	err := w.client.Connect(ctx)
	if err != nil {
		return w.errFactory("Connect", sdkerr.ErrWSConnection, err)
	}
	w.readingMessage(ctx)
	w.startPinger(ctx)
	return nil
}

// Close shuts down the connection and internal workers. Safe to call multiple times.
func (w *WSMarket) Close() error {
	w.closeOnce.Do(func() {
		err := w.client.Close()
		if err != nil {
			w.closeErr = w.errFactory("Close", sdkerr.ErrWSClose, err)
		}
	})
	return w.closeErr
}

// Subscribe registers a new subscription and waits for server acknowledgment.
//
// Returns a SubscriptionHandle used for safe unsubscription.
//
// Errors:
//   - ErrDuplicateSubscription: subscription with the same key already exists.
//   - ErrMaxCountSubscribes: subscription limit exceeded.
//
// Panics if sub is nil.
func (w *WSMarket) Subscribe(ctx context.Context, sub Subscription) (SubscriptionHandle, error) {
	if nil == sub {
		panic("WSMarket.Subscribe: subscribe must not be nil")
	}

	subID := sub.id()

	w.activeSubsMu.Lock()
	defer w.activeSubsMu.Unlock()

	if _, ok := w.activeSubs[subID]; ok {
		return nil, w.errFactory("Subscribe", nil, ErrDuplicateSubscription).
			WithMessage(fmt.Sprintf("already subscribed: %s", subID))
	}

	if w.router.Len() >= maxCountSubscribes {
		return nil, w.errFactory("Subscribe", nil, ErrMaxCountSubscribes).
			WithMessage(fmt.Sprintf("subscription limit exceeded (max allowed: %d)", maxCountSubscribes))
	}

	req := newSubscriptionRequest(subscribe, sub)

	ctx, cancel := ensureDeadline(ctx, w.waitingTimeout)
	defer cancel()

	if err := w.sendAndAwaitResponse(ctx, req); err != nil {
		return nil, err
	}

	wrapper := &subscriptionWrapper{
		ws:    w,
		inner: sub,
	}

	w.activeSubs[subID] = wrapper
	w.router.Register(sub)
	return wrapper, nil
}

func (w *WSMarket) errFactory(op string, kind error, cause error) *sdkerr.SDKError {
	return sdkerr.NewSDKError().
		WithSubsys(subsys).
		WithOp(fmt.Sprintf("WSMarket.%s", op)).
		WithKind(kind).
		WithCause(cause)
}

type subscriptionWrapper struct {
	ws    *WSMarket
	inner subscriptionSpec
	once  sync.Once
	err   error
}

// Unsubscribe cancels the subscription. Safe to call multiple times.
// Waits for server acknowledgment using the provided context.
func (s *subscriptionWrapper) Unsubscribe(ctx context.Context) error {
	s.once.Do(func() {
		ctx, cancel := ensureDeadline(ctx, s.ws.waitingTimeout)
		defer cancel()

		s.ws.activeSubsMu.Lock()
		defer s.ws.activeSubsMu.Unlock()

		req := newSubscriptionRequest(unsubscribe, s.inner)

		msg, err := req.Message()
		if err != nil {
			s.err = s.ws.errFactory("Unsubscribe", sdkerr.ErrWSWrite, err)
			delete(s.ws.activeSubs, s.inner.id())
			s.ws.router.Unregister(s.inner)
			return
		}

		writeDone := make(chan error, 1)

		go func() {
			writeDone <- s.ws.client.WriteMessage(msg)
		}()

		select {
		case <-ctx.Done():
			s.err = s.ws.errFactory("Unsubscribe", sdkerr.ErrWSMessageTimeout, nil)
		case err := <-writeDone:
			if err != nil {
				s.err = s.ws.errFactory("Unsubscribe", sdkerr.ErrWSWrite, err)
			}
		}

		delete(s.ws.activeSubs, s.inner.id())
		s.ws.router.Unregister(s.inner)
	})
	return s.err
}

type message struct {
	Channel string          `json:"channel"`
	Data    json.RawMessage `json:"data"`
	Symbol  string          `json:"symbol"`
	Ts      int64           `json:"ts"`
}

func (w *WSMarket) startPinger(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(w.pingInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := w.sendPing(); err != nil {
					_ = w.Close()
					return
				}
			}
		}
	}()
}

type pingRequest struct{}

func (p *pingRequest) Message() ([]byte, error) {
	payload := wsRequestPayload{
		Method: "ping",
	}
	return json.Marshal(payload)
}

func (p *pingRequest) MatchFunc() matchFunc {
	return func(m *message) (bool, error) {
		if m.Channel == "pong" {
			return true, nil
		}
		if m.Channel == "rs.error" {
			return true, fmt.Errorf("ping failed: %s", string(m.Data))
		}
		return false, nil
	}
}

func (w *WSMarket) sendPing() error {
	startTime := w.now()

	ctx, cancel := context.WithDeadline(context.Background(), startTime.Add(w.internalTimeout))
	defer cancel()

	w.activeSubsMu.Lock()
	defer w.activeSubsMu.Unlock()

	req := &pingRequest{}

	if err := w.sendAndAwaitResponse(ctx, req); err != nil {
		return err
	}

	if w.onLatency != nil {
		w.onLatency(time.Since(startTime))
	}
	return nil
}

type wsRequest interface {
	Message() ([]byte, error)
	MatchFunc() matchFunc
}

func (w *WSMarket) sendAndAwaitResponse(
	ctx context.Context,
	req wsRequest,
) error {
	promise := w.promiseFunc(req.MatchFunc())

	w.promisesMu.Lock()
	w.promise = promise
	w.promisesMu.Unlock()

	defer func() {
		w.promisesMu.Lock()
		w.promise = nil
		w.promisesMu.Unlock()
	}()

	msg, err := req.Message()
	if err != nil {
		return w.errFactory("sendAndAwaitResponse", sdkerr.ErrWSWrite, err)
	}

	if err := w.client.WriteMessage(msg); err != nil {
		return w.errFactory("sendAndAwaitResponse", sdkerr.ErrWSWrite, err)
	}

	_, err = promise.Await(ctx)
	if err != nil {
		var perr *wsutil.PromiseError
		if errors.As(err, &perr) {
			switch perr.Source {
			case wsutil.FromContext:
				return w.errFactory("sendAndAwaitResponse", sdkerr.ErrWSMessageTimeout, nil).
					WithMessage(perr.Err.Error())
			case wsutil.FromServer:
				return w.errFactory("sendAndAwaitResponse", sdkerr.ErrWSServerError, nil).
					WithMessage(perr.Err.Error())
			}
		}
		panic("unreachable")
	}
	return nil
}

func (w *WSMarket) readingMessage(ctx context.Context) {
	go func() {
		defer w.Close()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				data, err := w.client.ReadMessage()
				if err != nil {
					err = w.errFactory("readingMessage", sdkerr.ErrWSRead, err)
					if w.onDisconnect != nil {
						w.onDisconnect(err)
					}
					return
				}
				w.handleMessage(data)
			}
		}
	}()
}

func (w *WSMarket) handleMessage(data []byte) {
	var (
		jsonData []byte
		err      error
	)

	if isCompressed(data) {
		jsonData, err = unzipPayload(data)
		if err != nil {
			return
		}
	} else {
		jsonData = data
	}

	var msg *message
	if err := json.Unmarshal(jsonData, &msg); err != nil {
		return
	}

	w.promisesMu.Lock()
	defer w.promisesMu.Unlock()

	if w.promise != nil {
		ok, err := w.promise.Match(msg)
		if ok {
			if err == nil {
				w.promise.Resolve(msg)
			} else {
				w.promise.Reject(err)
			}
			return
		}
	}

	w.router.Route(msg)
}

func ensureDeadline(ctx context.Context, fallback time.Duration) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, fallback)
}

func isCompressed(data []byte) bool {
	return len(data) >= 2 && data[0] == 0x1F && data[1] == 0x8B
}

func unzipPayload(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var out bytes.Buffer
	if _, err := io.Copy(&out, r); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}
