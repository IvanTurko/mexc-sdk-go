package wsmarket

import (
	"fmt"
	"testing"

	"github.com/IvanTurko/mexc-sdk-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func Test_NewBookDepthSub(t *testing.T) {
	t.Run("normal flow", func(t *testing.T) {
		symbol := "BTCUSDT"
		level := DepthLevel10
		expCh := fmt.Sprintf("spot@public.limit.depth.v3.api.pb@%s@%d", symbol, level)

		sub := NewBookDepthSub(symbol, DepthLevel10, func(*DepthSnapshot) {}).(*bookDepthSub)
		assert.Equal(t, expCh, sub.streamName)
	})
	t.Run("invalid symbol name", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "invalid symbol name")
		}()

		NewBookDepthSub("", DepthLevel10, func(*DepthSnapshot) {})
	})
	t.Run("invalid depth level", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "invalid depth level")
		}()

		NewBookDepthSub("BTCUSDT", DepthSize(11), func(*DepthSnapshot) {})
	})
	t.Run("onData is nil", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "onData function is nil")
		}()

		NewBookDepthSub("BTCUSDT", DepthLevel10, nil)
	})
}

func TestBookDepthSub_matches(t *testing.T) {
	sub := NewBookDepthSub("BTCUSDT", DepthLevel10, func(*DepthSnapshot) {}).(*bookDepthSub)

	msg := &message{
		Msg: sub.streamName,
	}

	ok, _ := sub.matches(msg)
	assert.True(t, ok)
}

func TestBookDepthSub_acceptEvent(t *testing.T) {
	sub := NewBookDepthSub("BTCUSDT", DepthLevel10, func(*DepthSnapshot) {}).(*bookDepthSub)

	msg := &PushDataV3MarketWrapper{
		Channel: sub.streamName,
	}
	assert.True(t, sub.acceptEvent(msg))
}

func TestBookDepthSub_handleEvent(t *testing.T) {
	t.Run("calls onData for valid input", func(t *testing.T) {
		var called bool

		sub := NewBookDepthSub("BTCUSDT", DepthLevel10, func(*DepthSnapshot) {
			called = true
		}).(*bookDepthSub)

		msg := &PushDataV3MarketWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3MarketWrapper_PublicLimitDepths{
				PublicLimitDepths: &PublicLimitDepthsV3Api{
					Asks: []*PublicLimitDepthV3ApiItem{
						{
							Price:    "9939.39",
							Quantity: "3939",
						},
					},
					Bids: []*PublicLimitDepthV3ApiItem{
						{
							Price:    "2929",
							Quantity: "9393.1",
						},
					},
					EventType: "public",
					Version:   "2929",
				},
			},
		}

		sub.handleEvent(msg)
		assert.True(t, called, "onData should be called")
	})

	t.Run("PublicLimitDepths is nil", func(t *testing.T) {
		sub := NewBookDepthSub("BTCUSDT", DepthLevel10, func(*DepthSnapshot) {
			t.Fatal("unexpected call to onData")
		}).(*bookDepthSub)

		sub.SetOnInvalid(func(error) {
			t.Fatal("unexpected call to onInvalid")
		})

		msg := &PushDataV3MarketWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3MarketWrapper_PublicLimitDepths{
				PublicLimitDepths: nil,
			},
		}

		sub.handleEvent(msg)
	})

	t.Run("symbol is nil", func(t *testing.T) {
		sub := NewBookDepthSub("BTCUSDT", DepthLevel10, func(*DepthSnapshot) {
			t.Fatal("unexpected call to onData")
		}).(*bookDepthSub)

		sub.SetOnInvalid(func(error) {
			t.Fatal("unexpected call to onInvalid")
		})

		msg := &PushDataV3MarketWrapper{
			Symbol: nil,
			Body: &PushDataV3MarketWrapper_PublicLimitDepths{
				PublicLimitDepths: &PublicLimitDepthsV3Api{},
			},
		}

		sub.handleEvent(msg)
	})

	t.Run("calls onInvalid for invalid input", func(t *testing.T) {
		var invalidCalled bool

		sub := NewBookDepthSub("BTCUSDT", DepthLevel10, func(*DepthSnapshot) {}).(*bookDepthSub)
		sub.SetOnInvalid(func(error) {
			invalidCalled = true
		})

		msg := &PushDataV3MarketWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3MarketWrapper_PublicLimitDepths{
				PublicLimitDepths: &PublicLimitDepthsV3Api{
					Asks: []*PublicLimitDepthV3ApiItem{
						{
							Price:    "9939.39",
							Quantity: "3939",
						},
					},
					Bids: []*PublicLimitDepthV3ApiItem{
						{
							Price:    "2929",
							Quantity: "INVALID",
						},
					},
					EventType: "public",
					Version:   "2929",
				},
			},
		}

		sub.handleEvent(msg)
		assert.True(t, invalidCalled, "onInvalid should be called")
	})
}

func Test_mapProtoLimitDepthSnapshot(t *testing.T) {
	t.Run("valid depth snapshot", func(t *testing.T) {
		sendTime := int64(111)

		msg := &PublicLimitDepthsV3Api{
			Version: "123",
			Asks: []*PublicLimitDepthV3ApiItem{
				{Price: "100.1", Quantity: "1.5"},
			},
			Bids: []*PublicLimitDepthV3ApiItem{
				{Price: "99.9", Quantity: "2.5"},
			},
		}

		got, err := mapProtoLimitDepthSnapshot(msg, "BTCUSDT", &sendTime)
		require.NoError(t, err)
		require.Len(t, got.Asks, 1)
		require.Len(t, got.Bids, 1)

		assert.Equal(t, "BTCUSDT", got.Symbol)
		assert.Equal(t, int64(123), got.Version)
		assert.Equal(t, &sendTime, got.SendTime)

		testutil.AssertDecimalEqual(t, got.Asks[0].Price, "100.1")
		testutil.AssertDecimalEqual(t, got.Asks[0].Quantity, "1.5")
		testutil.AssertDecimalEqual(t, got.Bids[0].Price, "99.9")
		testutil.AssertDecimalEqual(t, got.Bids[0].Quantity, "2.5")
	})

	t.Run("invalid version", func(t *testing.T) {
		msg := &PublicLimitDepthsV3Api{
			Version: "INVALID",
			Asks: []*PublicLimitDepthV3ApiItem{
				{Price: "100.1", Quantity: "1.5"},
			},
			Bids: []*PublicLimitDepthV3ApiItem{
				{Price: "99.9", Quantity: "2.5"},
			},
		}

		_, err := mapProtoLimitDepthSnapshot(msg, "BTCUSDT", nil)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid version")
	})

	t.Run("invalid asks", func(t *testing.T) {
		msg := &PublicLimitDepthsV3Api{
			Version: "123",
			Asks: []*PublicLimitDepthV3ApiItem{
				{Price: "INVALID", Quantity: "1.5"},
			},
			Bids: []*PublicLimitDepthV3ApiItem{
				{Price: "99.9", Quantity: "2.5"},
			},
		}

		_, err := mapProtoLimitDepthSnapshot(msg, "BTCUSDT", nil)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid asks")
	})

	t.Run("invalid bids", func(t *testing.T) {
		msg := &PublicLimitDepthsV3Api{
			Version: "123",
			Asks: []*PublicLimitDepthV3ApiItem{
				{Price: "100.1", Quantity: "1.5"},
			},
			Bids: []*PublicLimitDepthV3ApiItem{
				{Price: "INVALID", Quantity: "2.5"},
			},
		}

		_, err := mapProtoLimitDepthSnapshot(msg, "BTCUSDT", nil)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid bids")
	})
}

func Test_mapProtoDepthLevelFromLimitDepth(t *testing.T) {
	t.Run("valid depth level", func(t *testing.T) {
		items := []*PublicLimitDepthV3ApiItem{
			{Price: "101.5", Quantity: "2.3"},
		}

		got, err := mapProtoDepthLevelFromLimitDepth(items)
		require.NoError(t, err)
		require.Len(t, got, 1)
		testutil.AssertDecimalEqual(t, got[0].Price, "101.5")
		testutil.AssertDecimalEqual(t, got[0].Quantity, "2.3")
	})

	t.Run("invalid price", func(t *testing.T) {
		items := []*PublicLimitDepthV3ApiItem{
			{Price: "INVALID", Quantity: "2.3"},
		}

		_, err := mapProtoDepthLevelFromLimitDepth(items)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid price")
	})

	t.Run("invalid quantity", func(t *testing.T) {
		items := []*PublicLimitDepthV3ApiItem{
			{Price: "101.5", Quantity: "INVALID"},
		}

		_, err := mapProtoDepthLevelFromLimitDepth(items)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid quantity")
	})
}
