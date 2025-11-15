package ws

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

var errFake error = errors.New("fake error")

func Test_Close(t *testing.T) {
	t.Run("normal flow", func(t *testing.T) {
		t.Parallel()
		c := newFakeClient()
		_ = c.Close()
		if !c.conn.(*fakeConn).closed {
			t.Error("conn must be closed")
		}
	})
	t.Run("panic from conn=nil", func(t *testing.T) {
		t.Parallel()
		defer func() {
			_ = recover()
		}()
		c := newFakeClient(withConn(nil))
		_ = c.Close()
	})
	t.Run("panic from cancel=nil", func(t *testing.T) {
		t.Parallel()
		defer func() {
			_ = recover()
		}()
		c := newFakeClient()
		_ = c.Close()
	})
}

func Test_startReader(t *testing.T) {
	t.Run("normal message", func(t *testing.T) {
		t.Parallel()

		conn := newFakeConn()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client := &clientImp{
			conn:  conn,
			outCh: make(chan msgResult, 1),
		}

		conn.readCh <- readResp{
			msgType: websocket.TextMessage,
			data:    []byte("hello"),
			err:     nil,
		}

		go client.startReader(ctx)

		select {
		case res := <-client.outCh:
			assert.Equal(t, []byte("hello"), res.msg)
			assert.NoError(t, res.err)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for message")
		}
	})

	t.Run("read error", func(t *testing.T) {
		t.Parallel()

		conn := newFakeConn()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client := &clientImp{
			conn:  conn,
			outCh: make(chan msgResult, 1),
		}

		conn.readCh <- readResp{
			msgType: websocket.TextMessage,
			data:    []byte("bad"),
			err:     errFake,
		}

		go client.startReader(ctx)

		select {
		case res := <-client.outCh:
			assert.Equal(t, []byte("bad"), res.msg)
			assert.ErrorIs(t, res.err, ErrWSInternalError)
			assert.ErrorContains(t, res.err, errFake.Error())
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for error message")
		}
	})

	t.Run("normal close from server", func(t *testing.T) {
		t.Parallel()

		conn := newFakeConn()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client := &clientImp{
			conn:  conn,
			outCh: make(chan msgResult, 1),
		}

		conn.readCh <- readResp{
			msgType: websocket.TextMessage,
			data:    nil,
			err: &websocket.CloseError{
				Code: websocket.CloseNormalClosure,
				Text: "normal close",
			},
		}

		go client.startReader(ctx)

		select {
		case res := <-client.outCh:
			assert.Nil(t, res.msg)
			assert.ErrorIs(t, res.err, ErrWSNormalClosure)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for close message")
		}
	})

	t.Run("context cancel", func(t *testing.T) {
		t.Parallel()

		conn := newFakeConn()
		ctx, cancel := context.WithCancel(context.Background())

		client := &clientImp{
			conn:  conn,
			outCh: make(chan msgResult, 1),
		}

		cancel()

		go client.startReader(ctx)

		select {
		case <-client.outCh:
			t.Fatal("should not receive message after context cancel")
		case <-time.After(200 * time.Millisecond):
		}
	})
}

func newFakeClient(opts ...Option) *clientImp {
	c := &clientImp{
		conn:  newFakeConn(),
		outCh: make(chan msgResult, 10),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func withConn(conn *websocket.Conn) Option {
	return func(c *clientImp) {
		c.conn = conn
	}
}

type fakeConn struct {
	readCh      chan readResp
	writeCh     chan writeReq
	closed      bool
	writeCalled bool
}

type readResp struct {
	msgType int
	data    []byte
	err     error
}

type writeReq struct {
	msgType int
	data    []byte
}

func newFakeConn() *fakeConn {
	return &fakeConn{
		readCh:  make(chan readResp, 10),
		writeCh: make(chan writeReq, 10),
	}
}

func (c *fakeConn) ReadMessage() (int, []byte, error) {
	r := <-c.readCh
	return r.msgType, r.data, r.err
}

func (c *fakeConn) WriteMessage(messageType int, data []byte) error {
	c.writeCalled = true
	c.writeCh <- writeReq{msgType: messageType, data: data}
	return nil
}

func (c *fakeConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (c *fakeConn) Close() error {
	c.closed = true
	return nil
}

var _ Conn = (*fakeConn)(nil)
