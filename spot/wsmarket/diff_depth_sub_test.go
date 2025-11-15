package wsmarket

import (
	"fmt"
	"testing"
	"time"

	"github.com/IvanTurko/mexc-sdk-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func Test_NewDiffDepthSub(t *testing.T) {
	t.Run("normal flow", func(t *testing.T) {
		symbol := "BTCUSDT"
		interval := Update10ms
		expCh := fmt.Sprintf("spot@public.aggre.depth.v3.api.pb@%s@%s", interval, symbol)

		sub := NewDiffDepthSub(symbol, interval, func(*DepthDelta) {}).(*diffDepthSub)
		assert.Equal(t, expCh, sub.streamName)
	})
	t.Run("invalid symbol name", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "invalid symbol name")
		}()

		NewDiffDepthSub("", Update10ms, func(*DepthDelta) {})
	})
	t.Run("invalid update interval", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "invalid update interval")
		}()

		NewDiffDepthSub("BTCUSDT", UpdateInterval("hello"), func(*DepthDelta) {})
	})
	t.Run("onData is nil", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "onData function is nil")
		}()

		NewDiffDepthSub("BTCUSDT", Update10ms, nil)
	})
}

func TestDiffDepthSub_matches(t *testing.T) {
	sub := NewDiffDepthSub("BTCUSDT", Update10ms, func(*DepthDelta) {}).(*diffDepthSub)

	msg := &message{
		Msg: sub.streamName,
	}

	ok, _ := sub.matches(msg)
	assert.True(t, ok)
}

func TestDiffDepthSub_acceptEvent(t *testing.T) {
	sub := NewDiffDepthSub("BTCUSDT", Update10ms, func(*DepthDelta) {}).(*diffDepthSub)

	msg := &PushDataV3MarketWrapper{
		Channel: sub.streamName,
	}
	assert.True(t, sub.acceptEvent(msg))
}

func TestDiffDepthSub_handleEvent(t *testing.T) {
	t.Run("calls onData for valid input", func(t *testing.T) {
		var called bool

		sub := NewDiffDepthSub("BTCUSDT", Update10ms, func(*DepthDelta) {
			called = true
		}).(*diffDepthSub)

		msg := &PushDataV3MarketWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3MarketWrapper_PublicAggreDepths{
				PublicAggreDepths: &PublicAggreDepthsV3Api{
					Asks: []*PublicAggreDepthV3ApiItem{
						{
							Price:    "38923.382",
							Quantity: "893829.383",
						},
					},
					Bids: []*PublicAggreDepthV3ApiItem{
						{
							Price:    "90392.3930",
							Quantity: "9302.0392",
						},
					},
					EventType:   "public",
					FromVersion: "13093",
					ToVersion:   "338919",
				},
			},
		}

		sub.handleEvent(msg)
		assert.True(t, called, "onData should be called")
	})

	t.Run("PublicAggreDepths is nil", func(t *testing.T) {
		sub := NewDiffDepthSub("BTCUSDT", Update10ms, func(*DepthDelta) {
			t.Fatal("unexpected call to onData")
		}).(*diffDepthSub)

		sub.SetOnInvalid(func(error) {
			t.Fatal("unexpected call to onInvalid")
		})

		msg := &PushDataV3MarketWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3MarketWrapper_PublicAggreDepths{
				PublicAggreDepths: nil,
			},
		}

		sub.handleEvent(msg)
	})

	t.Run("symbol is nil", func(t *testing.T) {
		sub := NewDiffDepthSub("BTCUSDT", Update10ms, func(*DepthDelta) {
			t.Fatal("unexpected call to onData")
		}).(*diffDepthSub)

		sub.SetOnInvalid(func(error) {
			t.Fatal("unexpected call to onInvalid")
		})

		msg := &PushDataV3MarketWrapper{
			Symbol: nil,
			Body: &PushDataV3MarketWrapper_PublicAggreDepths{
				PublicAggreDepths: &PublicAggreDepthsV3Api{},
			},
		}

		sub.handleEvent(msg)
	})

	t.Run("calls onInvalid for invalid input", func(t *testing.T) {
		var invalidCalled bool

		sub := NewDiffDepthSub("BTCUSDT", Update10ms, func(*DepthDelta) {}).(*diffDepthSub)
		sub.SetOnInvalid(func(error) {
			invalidCalled = true
		})

		msg := &PushDataV3MarketWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3MarketWrapper_PublicAggreDepths{
				PublicAggreDepths: &PublicAggreDepthsV3Api{
					Asks: []*PublicAggreDepthV3ApiItem{
						{
							Price:    "38923.382",
							Quantity: "893829.383",
						},
					},
					Bids: []*PublicAggreDepthV3ApiItem{
						{
							Price:    "INVALID",
							Quantity: "9302.0392",
						},
					},
					EventType:   "public",
					FromVersion: "89182",
					ToVersion:   "338919",
				},
			},
		}

		sub.handleEvent(msg)
		assert.True(t, invalidCalled, "onInvalid should be called")
	})
}

func Test_mapProtoDepthDelta(t *testing.T) {
	sendTime := time.Now().Unix()
	symbol := "BTCUSDT"

	t.Run("valid depth delta", func(t *testing.T) {
		msg := &PublicAggreDepthsV3Api{
			FromVersion: "123",
			ToVersion:   "124",
			EventType:   "depthUpdate",
			Asks: []*PublicAggreDepthV3ApiItem{
				{Price: "100.0", Quantity: "0.5"},
			},
			Bids: []*PublicAggreDepthV3ApiItem{
				{Price: "99.5", Quantity: "1.0"},
			},
		}

		depth, err := mapProtoDepthDelta(msg, symbol, &sendTime)
		require.NoError(t, err)
		require.NotNil(t, depth)
		require.Len(t, depth.Asks, 1)
		require.Len(t, depth.Bids, 1)

		assert.Equal(t, symbol, depth.Symbol)
		assert.Equal(t, "depthUpdate", depth.EventType)
		assert.Equal(t, uint64(123), depth.FromVersion)
		assert.Equal(t, uint64(124), depth.ToVersion)
		assert.Equal(t, &sendTime, depth.SendTime)

		testutil.AssertDecimalEqual(t, depth.Asks[0].Price, "100.0")
		testutil.AssertDecimalEqual(t, depth.Asks[0].Quantity, "0.5")
		testutil.AssertDecimalEqual(t, depth.Bids[0].Price, "99.5")
		testutil.AssertDecimalEqual(t, depth.Bids[0].Quantity, "1.0")
	})

	t.Run("invalid fromVersion", func(t *testing.T) {
		msg := &PublicAggreDepthsV3Api{
			FromVersion: "INVALID",
			ToVersion:   "124",
			EventType:   "depthUpdate",
			Asks: []*PublicAggreDepthV3ApiItem{
				{Price: "100.0", Quantity: "0.5"},
			},
			Bids: []*PublicAggreDepthV3ApiItem{
				{Price: "99.5", Quantity: "1.0"},
			},
		}

		_, err := mapProtoDepthDelta(msg, symbol, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid fromVersion")
	})

	t.Run("invalid toVersion", func(t *testing.T) {
		msg := &PublicAggreDepthsV3Api{
			FromVersion: "123",
			ToVersion:   "INVALID",
			EventType:   "depthUpdate",
			Asks: []*PublicAggreDepthV3ApiItem{
				{Price: "100.0", Quantity: "0.5"},
			},
			Bids: []*PublicAggreDepthV3ApiItem{
				{Price: "99.5", Quantity: "1.0"},
			},
		}

		_, err := mapProtoDepthDelta(msg, symbol, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid toVersion")
	})

	t.Run("invalid asks", func(t *testing.T) {
		msg := &PublicAggreDepthsV3Api{
			FromVersion: "123",
			ToVersion:   "124",
			EventType:   "depthUpdate",
			Asks: []*PublicAggreDepthV3ApiItem{
				{Price: "INVALID", Quantity: "0.5"},
			},
			Bids: []*PublicAggreDepthV3ApiItem{
				{Price: "99.5", Quantity: "1.0"},
			},
		}

		_, err := mapProtoDepthDelta(msg, symbol, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid asks")
	})

	t.Run("invalid bids", func(t *testing.T) {
		msg := &PublicAggreDepthsV3Api{
			FromVersion: "123",
			ToVersion:   "124",
			EventType:   "depthUpdate",
			Asks: []*PublicAggreDepthV3ApiItem{
				{Price: "100.0", Quantity: "0.5"},
			},
			Bids: []*PublicAggreDepthV3ApiItem{
				{Price: "INVALID", Quantity: "1.0"},
			},
		}

		_, err := mapProtoDepthDelta(msg, symbol, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid bids")
	})
}

func Test_mapProtoDepthLevelFromAggreDepth(t *testing.T) {
	t.Run("valid levels", func(t *testing.T) {
		items := []*PublicAggreDepthV3ApiItem{
			{Price: "100.5", Quantity: "1.23"},
		}

		levels, err := mapProtoDepthLevelFromAggreDepth(items)
		require.NoError(t, err)

		testutil.AssertDecimalEqual(t, levels[0].Price, "100.5")
		testutil.AssertDecimalEqual(t, levels[0].Quantity, "1.23")
	})

	t.Run("invalid price", func(t *testing.T) {
		items := []*PublicAggreDepthV3ApiItem{
			{Price: "INVALID", Quantity: "1.23"},
		}

		_, err := mapProtoDepthLevelFromAggreDepth(items)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid price")
	})

	t.Run("invalid quantity", func(t *testing.T) {
		items := []*PublicAggreDepthV3ApiItem{
			{Price: "100.5", Quantity: "INVALID"},
		}

		_, err := mapProtoDepthLevelFromAggreDepth(items)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid quantity")
	})
}
