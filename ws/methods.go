package ws

import (
	"context"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/gorilla/websocket"
)

type msgResult struct {
	msg []byte
	err error
}

// Connect establishes a websocket connection.
func (c *clientImp) Connect(ctx context.Context) error {
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.url, nil)
	if err != nil {
		c.errorf("connect failed: %v", err)
		return classifyWSError(err)
	}

	c.conn = conn
	c.debugf("connected to %s", c.url)

	go c.startReader(ctx)
	return nil
}

// ReadMessage reads a message from the websocket connection.
func (c *clientImp) ReadMessage() ([]byte, error) {
	res := <-c.outCh
	return res.msg, res.err
}

func (c *clientImp) startReader(ctx context.Context) {
	defer c.Close()
	c.debugf("reader started")

	for {
		select {
		case <-ctx.Done():
			c.debugf("reader stopped by ctx")
			return
		default:
			if !c.readAndHandle() {
				return
			}
		}
	}
}

func (c *clientImp) readAndHandle() bool {
	msgType, buf, err := c.conn.ReadMessage()
	msgTypeStr := typeMsg(msgType)

	logWSMessage(c, msgTypeStr, buf)

	if err != nil {
		c.errorf("recv [%s] error: %v", msgTypeStr, err)
		err = classifyWSError(err)
		c.sendOrDrop(buf, err)
		return false
	}

	c.sendOrDrop(buf, nil)
	return true
}

func (c *clientImp) sendOrDrop(buf []byte, err error) {
	select {
	case c.outCh <- msgResult{msg: buf, err: err}:
	default:
		c.errorf("WS message dropped — increase buffer or read faster")
	}
}

func logWSMessage(c *clientImp, msgTypeStr string, buf []byte) {
	if len(buf) > 0 {
		if utf8.Valid(buf) {
			c.debugf("recv [%s]: %s", msgTypeStr, string(buf))
		} else {
			c.debugf("recv [%s]: <binary> %x", msgTypeStr, buf)
		}
	} else {
		c.debugf("recv [%s]: <empty>", msgTypeStr)
	}
}

// WriteMessage writes a message to the websocket connection.
func (c *clientImp) WriteMessage(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.conn.SetWriteDeadline(time.Now().Add(c.userWriteTimeout)); err != nil {
		c.errorf("set write deadline failed: %v", err)
		return fmt.Errorf("%w: %v", ErrWriteTimeout, err)
	}

	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		c.errorf("write failed: %v", err)
		return fmt.Errorf("%w: %v", ErrWriteFailed, err)
	}

	return nil
}

// Close closes the websocket connection.
func (c *clientImp) Close() error {
	c.closeOnce.Do(func() {
		if nil == c.conn {
			panic("Close: cannot call Close before successful Connect")
		}

		err := c.conn.Close()
		if err != nil {
			c.errorf("connection close failed: %v", err)
			c.closeErr = classifyWSError(err)
		}
	})
	return c.closeErr
}

func (c *clientImp) debugf(format string, args ...any) {
	if c.logger != nil {
		c.logger.Debugf(format, args...)
	}
}

func (c *clientImp) errorf(format string, args ...any) {
	if c.logger != nil {
		c.logger.Errorf(format, args...)
	}
}

func typeMsg(code int) string {
	switch code {
	case websocket.TextMessage:
		return "Text"
	case websocket.BinaryMessage:
		return "Binary"
	case websocket.CloseMessage:
		return "Close"
	case websocket.PingMessage:
		return "Ping"
	case websocket.PongMessage:
		return "Pong"
	default:
		return fmt.Sprintf("Unknown(%d)", code)
	}
}
