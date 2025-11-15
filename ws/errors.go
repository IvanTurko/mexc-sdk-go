package ws

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/gorilla/websocket"
)

var (
	// ErrWriteTimeout is returned when a write operation times out.
	ErrWriteTimeout = errors.New("write timeout")
	// ErrWriteFailed is returned when a write operation fails.
	ErrWriteFailed = errors.New("write failed")
	// ErrWSNormalClosure is returned when a websocket connection is closed normally.
	ErrWSNormalClosure = errors.New("connection closed normally")
	// ErrWSAbnormalClosure is returned when a websocket connection is closed abnormally.
	ErrWSAbnormalClosure = errors.New("connection closed abnormally")
	// ErrWSNetworkIssue is returned when a network issue occurs.
	ErrWSNetworkIssue = errors.New("network issue")
	// ErrWSReadInterrupted is returned when a read operation is interrupted.
	ErrWSReadInterrupted = errors.New("read interrupted")
	// ErrWSPayloadCorrupted is returned when a payload is invalid or corrupted.
	ErrWSPayloadCorrupted = errors.New("invalid/corrupted payload")
	// ErrWSUnexpectedEOF is returned when an unexpected EOF is encountered.
	ErrWSUnexpectedEOF = errors.New("unexpected EOF")
	// ErrWSInternalError is returned for internal client errors.
	ErrWSInternalError = errors.New("internal client error")
)

func classifyWSError(err error) error {
	switch {
	case websocket.IsCloseError(err, websocket.CloseNormalClosure):
		return fmt.Errorf("%w: %v", ErrWSNormalClosure, err)

	case websocket.IsCloseError(err, websocket.CloseAbnormalClosure):
		return fmt.Errorf("%w: %v", ErrWSAbnormalClosure, err)

	case isReadInterrupted(err):
		return fmt.Errorf("%w: %v", ErrWSReadInterrupted, err)

	case isUnexpectedEOF(err):
		return fmt.Errorf("%w: %v", ErrWSUnexpectedEOF, err)

	case isNetError(err):
		return fmt.Errorf("%w: %v", ErrWSNetworkIssue, err)

	case isPayloadCorrupted(err):
		return fmt.Errorf("%w: %v", ErrWSPayloadCorrupted, err)

	default:
		return fmt.Errorf("%w: %v", ErrWSInternalError, err)
	}
}

func isReadInterrupted(err error) bool {
	return errors.Is(err, net.ErrClosed) ||
		errors.Is(err, io.ErrClosedPipe) ||
		strings.Contains(err.Error(), "use of closed network connection") ||
		strings.Contains(err.Error(), "close sent")
}

func isUnexpectedEOF(err error) bool {
	return errors.Is(err, io.EOF) ||
		strings.Contains(err.Error(), "unexpected EOF")
}

func isNetError(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr)
}

func isPayloadCorrupted(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "invalid UTF-8") ||
		strings.Contains(errMsg, "malformed") ||
		strings.Contains(errMsg, "unexpected opcode") ||
		strings.Contains(errMsg, "read limit exceeded")
}
