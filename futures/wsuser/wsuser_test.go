package wsuser

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/IvanTurko/mexc-sdk-go/internal/testutil"
	"github.com/IvanTurko/mexc-sdk-go/internal/wsutil"
	"github.com/IvanTurko/mexc-sdk-go/sdkerr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fakeErr = errors.New("fake")

func TestWSUser_Connect_Failure(t *testing.T) {
	ws := &WSUser{
		client: &testutil.MockClient{
			ConnectErr: fakeErr,
		},
	}

	err := ws.Connect(context.Background())
	require.Error(t, err)
	require.ErrorContains(t, err, err.Error())
}

func TestWSUser_Close_Idempotent(t *testing.T) {
	client := &testutil.MockClient{}
	ws := &WSUser{client: client}

	_ = ws.Close()
	err := ws.Close()
	require.NoError(t, err)
	require.True(t, client.Closed)
}

func TestWSUser_Subscribe(t *testing.T) {
	const streamName = "test-id"

	t.Run("successful subscription", func(t *testing.T) {
		expCountHadlers := 2
		router := &mockHandlerRouter{count: expCountHadlers - 1}

		ws := &WSUser{
			client:      &testutil.MockClient{},
			router:      router,
			activeSubs:  make(map[string]SubscriptionHandle),
			promiseFunc: fakePromiseFunc,
		}

		handle, err := ws.Subscribe(context.Background(), &mockSubscription{StreamName: streamName})
		require.NoError(t, err)
		assert.NotNil(t, handle)
		assert.Equal(t, expCountHadlers, router.count)
		assert.Len(t, ws.activeSubs, 1)
	})

	t.Run("sub is nil", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "subscribe must not be nil")
		}()

		ws := &WSUser{}
		ws.Subscribe(context.Background(), nil)
	})

	t.Run("duplicate subscription", func(t *testing.T) {
		ws := &WSUser{
			client:     &testutil.MockClient{},
			router:     &mockHandlerRouter{},
			activeSubs: map[string]SubscriptionHandle{streamName: nil},
		}

		handle, err := ws.Subscribe(context.Background(), &mockSubscription{StreamName: streamName})
		require.Error(t, err)
		assert.Nil(t, handle)
		assert.True(t, errors.Is(err, ErrDuplicateSubscription))
	})

	t.Run("subscription limit exceeded", func(t *testing.T) {
		ws := &WSUser{
			client:     &testutil.MockClient{},
			router:     &mockHandlerRouter{len: maxCountSubscribes},
			activeSubs: make(map[string]SubscriptionHandle),
		}

		handle, err := ws.Subscribe(context.Background(), &mockSubscription{StreamName: streamName})
		require.Error(t, err)
		assert.Nil(t, handle)
		assert.True(t, errors.Is(err, ErrMaxCountSubscribes))
	})
}

func Test_subscriptionWrapper_Unsubscribe(t *testing.T) {
	t.Run("successful unsubscription", func(t *testing.T) {
		expCountHadlers := 2
		router := &mockHandlerRouter{count: expCountHadlers + 1}

		ws := &WSUser{
			router:     router,
			activeSubs: map[string]SubscriptionHandle{},
		}

		sub := &subscriptionWrapper{
			ws:    ws,
			inner: &mockSubscription{},
		}

		err := sub.Unsubscribe(context.Background())
		require.NoError(t, err)
		assert.Equal(t, expCountHadlers, router.count)
		assert.Len(t, ws.activeSubs, 0)
	})
}

func Test_subscriptionWrapper_Idempotent(t *testing.T) {
	ws := &WSUser{
		client:     &testutil.MockClient{},
		router:     &mockHandlerRouter{},
		activeSubs: map[string]SubscriptionHandle{},
	}

	sub := &subscriptionWrapper{
		ws:    ws,
		inner: &mockSubscription{},
	}

	_ = sub.Unsubscribe(context.Background())
	err := sub.Unsubscribe(context.Background())
	assert.NoError(t, err)
}

func TestWSUser_sendAndAwaitResponse(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		ws := &WSUser{
			client:      &testutil.MockClient{},
			promiseFunc: fakePromiseFunc,
		}

		err := ws.sendAndAwaitResponse(context.Background(), &mockWSRequest{})
		require.NoError(t, err)
	})

	t.Run("write error from request", func(t *testing.T) {
		ws := &WSUser{
			promiseFunc: fakePromiseFunc,
		}

		err := ws.sendAndAwaitResponse(context.Background(), &mockWSRequest{msgErr: fakeErr})
		require.Error(t, err)
		require.ErrorIs(t, err, sdkerr.ErrWSWrite)
		require.ErrorIs(t, err, fakeErr)
	})

	t.Run("write error from server", func(t *testing.T) {
		client := &testutil.MockClient{
			WriteFunc: func(msg []byte) error {
				return fakeErr
			},
		}

		ws := &WSUser{
			client:      client,
			promiseFunc: fakePromiseFunc,
		}

		err := ws.sendAndAwaitResponse(context.Background(), &mockWSRequest{})
		require.Error(t, err)
		require.ErrorIs(t, err, sdkerr.ErrWSWrite)
		require.ErrorIs(t, err, fakeErr)
	})

	t.Run("timeout error", func(t *testing.T) {
		awaitFunc := func() (*message, error) {
			return nil, &wsutil.PromiseError{Source: wsutil.FromContext, Err: fakeErr}
		}

		ws := &WSUser{
			client:      &testutil.MockClient{},
			promiseFunc: newMockPromise(awaitFunc),
		}

		err := ws.sendAndAwaitResponse(context.Background(), &mockWSRequest{})
		require.Error(t, err)
		require.ErrorIs(t, err, sdkerr.ErrWSMessageTimeout)
	})

	t.Run("server error", func(t *testing.T) {
		awaitFunc := func() (*message, error) {
			return nil, &wsutil.PromiseError{Source: wsutil.FromServer, Err: fakeErr}
		}

		ws := &WSUser{
			client:      &testutil.MockClient{},
			promiseFunc: newMockPromise(awaitFunc),
		}

		err := ws.sendAndAwaitResponse(context.Background(), &mockWSRequest{})
		require.Error(t, err)
		require.ErrorIs(t, err, sdkerr.ErrWSServerError)
	})
}

func TestWSUser_Subscribe_Sequential_NoRace(t *testing.T) {
	ws := &WSUser{
		client: &testutil.MockClient{},
	}

	raceCount := 0
	awaitFunc := func() (*message, error) {
		raceCount++
		return nil, nil
	}

	ws.promiseFunc = newMockPromise(awaitFunc)

	n := 10
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()

			// subscribe operation like
			ws.activeSubsMu.Lock()
			defer ws.activeSubsMu.Unlock()

			err := ws.sendAndAwaitResponse(context.Background(), &mockWSRequest{})
			if err != nil {
				t.Error(err)
			}
		}()
	}

	wg.Wait()
}
