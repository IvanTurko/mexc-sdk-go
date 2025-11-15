package ws

import (
	"context"
	"sync"
	"time"
)

// Client is the interface for a websocket client.
type Client interface {
	Connect(context.Context) error
	ReadMessage() ([]byte, error)
	WriteMessage([]byte) error
	Close() error
}

// Conn is an interface for a websocket connection.
type Conn interface {
	Close() error
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	SetWriteDeadline(t time.Time) error
}

// Option is a function type for client options.
type Option func(*clientImp)

// Logger is an interface for logging.
type Logger interface {
	Debugf(format string, args ...any)
	Errorf(format string, args ...any)
}

type clientImp struct {
	conn   Conn
	logger Logger

	url              string
	mu               sync.Mutex
	outCh            chan msgResult
	userWriteTimeout time.Duration

	closeOnce sync.Once
	closeErr  error
}

// NewClient creates a new websocket client for the given URL.
//
// By default:
//   - The client uses a buffered output channel of size 1000.
//   - The write timeout is 300 milliseconds.
//
// Options can be passed to override defaults (e.g., custom logger, write timeout, or channel size).
//
// Panics if the URL is empty.
func NewClient(url string, opts ...Option) Client {
	if url == "" {
		panic("url must not be empty")
	}

	c := &clientImp{
		url:              url,
		userWriteTimeout: 300 * time.Millisecond,
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.outCh == nil {
		c.outCh = make(chan msgResult, 1000)
	}
	return c
}

// WithLogger sets the logger for the client.
func WithLogger(l Logger) Option {
	return func(c *clientImp) {
		c.logger = l
	}
}

// WithWriteTimeout sets the write timeout for the client.
// The default is 300 milliseconds.
func WithWriteTimeout(d time.Duration) Option {
	return func(c *clientImp) {
		c.userWriteTimeout = d
	}
}

// WithBufferedOutCh sets the buffered output channel size.
// If n is not positive, it defaults to 1000.
func WithBufferedOutCh(n int) Option {
	return func(c *clientImp) {
		if n <= 0 {
			n = 1000
		}
		c.outCh = make(chan msgResult, n)
	}
}
