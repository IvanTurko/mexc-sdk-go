package testutil

import "context"

type MockClient struct {
	ConnectErr error
	CloseErr   error
	Closed     bool
	ReadFunc   func() ([]byte, error)
	WriteFunc  func(msg []byte) error
}

func (m *MockClient) Connect(ctx context.Context) error {
	return m.ConnectErr
}

func (m *MockClient) Close() error {
	m.Closed = true
	return m.CloseErr
}

func (m *MockClient) ReadMessage() ([]byte, error) {
	if nil == m.ReadFunc {
		return nil, nil
	}
	return m.ReadFunc()
}

func (m *MockClient) WriteMessage(msg []byte) error {
	if nil == m.WriteFunc {
		return nil
	}
	return m.WriteFunc(msg)
}
