package wsmarket

import (
	"context"

	"github.com/IvanTurko/mexc-sdk-go/internal/wsutil"
)

type mockHandlerRouter struct {
	len   int
	count int
}

func (m *mockHandlerRouter) Register(h wsHandler) {
	m.count++
}

func (m *mockHandlerRouter) Unregister(h wsHandler) {
	m.count--
}

func (m *mockHandlerRouter) Route(msg *message) {
}

func (m *mockHandlerRouter) Len() int {
	return m.len
}

type mockSubscription struct {
	StreamName string
}

func (m *mockSubscription) SetOnInvalid(f func(error)) Subscription {
	return m
}

func (m *mockSubscription) matches(msg *message) (bool, error) {
	return false, nil
}

func (m *mockSubscription) acceptEvent(msg *message) bool {
	return false
}

func (m *mockSubscription) handleEvent(msg *message) {}

func (m *mockSubscription) id() string {
	return m.StreamName
}

func (m *mockSubscription) payload(op subscriptionOp) any {
	return m.StreamName
}

type mockWSRequest struct {
	msgErr error
}

func (s *mockWSRequest) ID() uint64 {
	return 0
}

func (s *mockWSRequest) Message() ([]byte, error) {
	return nil, s.msgErr
}

func (s *mockWSRequest) MatchFunc() matchFunc {
	return func(msg *message) (bool, error) {
		return false, nil
	}
}

type promisesFunc func(matchFn func(*message) (bool, error)) wsutil.Promise[message]

func newMockPromise(
	awaitFunc func() (*message, error),
) promisesFunc {
	return func(_ func(*message) (bool, error)) wsutil.Promise[message] {
		return &mockPromise{
			awaitFunc: awaitFunc,
		}
	}
}

var fakePromiseFunc = func(matchFn func(*message) (bool, error)) wsutil.Promise[message] {
	return &mockPromise{}
}

type mockPromise struct {
	awaitFunc func() (*message, error)
}

func (m *mockPromise) Match(msg *message) (bool, error) {
	return false, nil
}

func (m *mockPromise) Resolve(msg *message) {
}

func (m *mockPromise) Reject(err error) {
}

func (m *mockPromise) Await(ctx context.Context) (*message, error) {
	if nil == m.awaitFunc {
		return nil, nil
	}
	return m.awaitFunc()
}

var (
	_ handlerRouter = (*mockHandlerRouter)(nil)
	_ wsRequest     = (*mockWSRequest)(nil)
)
