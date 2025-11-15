package wsuser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/IvanTurko/mexc-sdk-go/internal/signature"
	"github.com/IvanTurko/mexc-sdk-go/internal/wsutil"
	"github.com/IvanTurko/mexc-sdk-go/sdkerr"
	"github.com/IvanTurko/mexc-sdk-go/ws"
)

const (
	subsys             = "futures/wsuser"
	defaultBaseURL     = "wss://contract.mexc.com/edge"
	maxCountSubscribes = 6
)

var (
	// ErrDuplicateSubscription is returned when a subscription with the same key already exists.
	ErrDuplicateSubscription = errors.New("duplicate subscription")
	// ErrMaxCountSubscribes is returned when the subscription limit is reached.
	ErrMaxCountSubscribes = errors.New("maximum number of subscribers exceeded")
)

// Options configures WSUser.
type Options = func(*WSUser)

// Subscription describes a WebSocket subscription (orders, positions, etc.).
// Instances are normally created via constructors like NewOrderSub().
type Subscription interface {
	// SetOnInvalid registers a callback fired when the subscription becomes invalid
	// (malformed message, server rejection, etc.).
	SetOnInvalid(f func(error)) Subscription
	subscriptionSpec
}

// SubscriptionHandle represents an active subscription.
type SubscriptionHandle interface {
	// Unsubscribe stops this subscription. Safe to call multiple times.
	Unsubscribe(ctx context.Context) error
}

// WSUser is a WebSocket client for MEXC FUTURES user streams.
type WSUser struct {
	client         ws.Client
	apiKey         string
	secretKey      string
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

// NewWSUser creates a WSUser using the default WebSocket client.
// Additional configuration can be supplied through Options.
// Panics if apiKey or secretKey are empty.
func NewWSUser(apiKey, secretKey string, opts ...Options) *WSUser {
	if apiKey == "" {
		panic("NewWSUser: apiKey is required")
	}
	if secretKey == "" {
		panic("NewWSUser: secretKey is required")
	}
	factory := func(url string) ws.Client {
		return ws.NewClient(url)
	}
	return NewWSUserWithFactory(apiKey, secretKey, factory, opts...)
}

// NewWSUserWithFactory is like NewWSUser but uses the provided ws.Client factory.
// Useful for tests and custom connection setups.
// Panics if apiKey, secretKey are empty or factory is nil.
func NewWSUserWithFactory(apiKey, secretKey string, factory func(url string) ws.Client, opts ...Options) *WSUser {
	if apiKey == "" {
		panic("NewWSUserWithFactory: apiKey is required")
	}
	if secretKey == "" {
		panic("NewWSUserWithFactory: secretKey is required")
	}
	if factory == nil {
		panic("NewWSUserWithFactory: factory must not be nil")
	}

	w := &WSUser{
		client:         factory(defaultBaseURL),
		apiKey:         apiKey,
		secretKey:      secretKey,
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
	return func(w *WSUser) {
		w.onDisconnect = f
	}
}

// WithPingLatencyHandler registers a callback receiving RTT ping/pong measurements.
func WithPingLatencyHandler(f func(time.Duration)) Options {
	return func(w *WSUser) {
		w.onLatency = f
	}
}

// Connect opens the WebSocket connection, authenticates, and starts internal workers.
func (w *WSUser) Connect(ctx context.Context) error {
	err := w.client.Connect(ctx)
	if err != nil {
		return w.errFactory("Connect", sdkerr.ErrWSConnection, err)
	}
	w.readingMessage(ctx)
	w.startPinger(ctx)

	if err := w.login(); err != nil {
		w.Close()
		return err
	}
	return nil
}

// Close shuts down the connection and internal workers. Safe to call multiple times.
func (w *WSUser) Close() error {
	w.closeOnce.Do(func() {
		err := w.client.Close()
		if err != nil {
			w.closeErr = w.errFactory("Close", sdkerr.ErrWSClose, err)
		}
	})
	return w.closeErr
}

// Subscribe registers a new subscription.
//
// Returns a SubscriptionHandle used for safe unsubscription.
//
// Errors:
//   - ErrDuplicateSubscription: subscription with the same key already exists.
//   - ErrMaxCountSubscribes: subscription limit exceeded (max 6 for user streams).
//
// Panics if sub is nil.
func (w *WSUser) Subscribe(ctx context.Context, sub Subscription) (SubscriptionHandle, error) {
	if nil == sub {
		panic("WSUser.Subscribe: subscribe must not be nil")
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

	wrapper := &subscriptionWrapper{
		ws:    w,
		inner: sub,
	}

	w.activeSubs[subID] = wrapper
	w.router.Register(sub)
	return wrapper, nil
}

func (w *WSUser) errFactory(op string, kind error, cause error) *sdkerr.SDKError {
	return sdkerr.NewSDKError().
		WithSubsys(subsys).
		WithOp(fmt.Sprintf("WSUser.%s", op)).
		WithKind(kind).
		WithCause(cause)
}

type subscriptionWrapper struct {
	ws    *WSUser
	inner subscriptionSpec
	once  sync.Once
}

// Unsubscribe cancels the subscription. Safe to call multiple times.
func (s *subscriptionWrapper) Unsubscribe(ctx context.Context) error {
	s.once.Do(func() {
		delete(s.ws.activeSubs, s.inner.id())
		s.ws.router.Unregister(s.inner)
	})
	return nil
}

type message struct {
	Channel string          `json:"channel"`
	Data    json.RawMessage `json:"data"`
	Symbol  string          `json:"symbol"`
	Ts      int64           `json:"ts"`
}

func (w *WSUser) login() error {
	startTime := w.now()

	ctx, cancel := context.WithDeadline(context.Background(), startTime.Add(w.internalTimeout))
	defer cancel()

	w.activeSubsMu.Lock()
	defer w.activeSubsMu.Unlock()

	req := &loginRequest{
		apiKey:    w.apiKey,
		secretKey: w.secretKey,
		now:       w.now,
	}

	if err := w.sendAndAwaitResponse(ctx, req); err != nil {
		return err
	}

	return nil
}

type loginRequest struct {
	apiKey    string
	secretKey string
	now       func() time.Time
}

func (l *loginRequest) MatchFunc() matchFunc {
	return func(msg *message) (bool, error) {
		if msg.Channel == "rs.login" {
			var s string
			if err := json.Unmarshal(msg.Data, &s); err != nil {
				return false, fmt.Errorf("invalid success payload: %s", string(msg.Data))
			}
			if s == "success" {
				return true, nil
			}
		}

		if msg.Channel == "rs.error" {
			return true, fmt.Errorf("auth failed: %s", string(msg.Data))
		}

		return false, nil
	}
}

func (l *loginRequest) Message() ([]byte, error) {
	timestamp := strconv.FormatInt(l.now().UnixMilli(), 10)
	target := l.apiKey + timestamp

	sig := signature.HMACSHA256(target, l.secretKey)

	payload := wsRequestPayload{
		Method: "login",
		Param: map[string]any{
			"apiKey":    l.apiKey,
			"signature": sig,
			"reqTime":   timestamp,
		},
	}

	return json.Marshal(payload)
}

func (w *WSUser) startPinger(ctx context.Context) {
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

func (w *WSUser) sendPing() error {
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

func (w *WSUser) sendAndAwaitResponse(
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

func (w *WSUser) readingMessage(ctx context.Context) {
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

func (w *WSUser) handleMessage(data []byte) {
	var msg *message
	if err := json.Unmarshal(data, &msg); err != nil {
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
