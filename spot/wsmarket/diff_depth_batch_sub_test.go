package wsmarket

import (
	"fmt"
	"testing"

	"github.com/IvanTurko/mexc-sdk-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func Test_NewDiffDepthBatchSub(t *testing.T) {
	t.Run("normal flow", func(t *testing.T) {
		symbol := "BTCUSDT"
		expCh := fmt.Sprintf("spot@public.increase.depth.batch.v3.api.pb@%s", symbol)

		sub := NewDiffDepthBatchSub(symbol, func(*DepthBatch) {}).(*diffDepthBatchSub)
		assert.Equal(t, expCh, sub.streamName)
	})
	t.Run("invalid symbol name", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "invalid symbol name")
		}()

		NewDiffDepthBatchSub("", func(*DepthBatch) {})
	})
	t.Run("onData is nil", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "onData function is nil")
		}()

		NewDiffDepthBatchSub("BTCUSDT", nil)
	})
}

func TestDiffDepthBatchSub_matches(t *testing.T) {
	sub := NewDiffDepthBatchSub("BTCUSDT", func(*DepthBatch) {}).(*diffDepthBatchSub)

	msg := &message{
		Msg: sub.streamName,
	}

	ok, _ := sub.matches(msg)
	assert.True(t, ok)
}

func TestDiffDepthBatchSub_acceptEvent(t *testing.T) {
	sub := NewDiffDepthBatchSub("BTCUSDT", func(*DepthBatch) {}).(*diffDepthBatchSub)

	msg := &PushDataV3MarketWrapper{
		Channel: sub.streamName,
	}
	assert.True(t, sub.acceptEvent(msg))
}

func TestDiffDepthBatchSub_handleEvent(t *testing.T) {
	t.Run("calls onData for valid input", func(t *testing.T) {
		var called bool

		sub := NewDiffDepthBatchSub("BTCUSDT", func(*DepthBatch) {
			called = true
		}).(*diffDepthBatchSub)

		msg := &PushDataV3MarketWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3MarketWrapper_PublicIncreaseDepthsBatch{
				PublicIncreaseDepthsBatch: &PublicIncreaseDepthsBatchV3Api{
					Items: []*PublicIncreaseDepthsV3Api{
						{
							Asks: []*PublicIncreaseDepthV3ApiItem{
								{
									Price:    "8329.93",
									Quantity: "82938.9",
								},
							},
							Bids:    []*PublicIncreaseDepthV3ApiItem{},
							Version: "1289128",
						},
					},
					EventType: "public",
				},
			},
		}

		sub.handleEvent(msg)
		assert.True(t, called, "onData should be called")
	})

	t.Run("PublicIncreaseDepthsBatch is nil", func(t *testing.T) {
		sub := NewDiffDepthBatchSub("BTCUSDT", func(*DepthBatch) {
			t.Fatal("unexpected call to onData")
		}).(*diffDepthBatchSub)

		sub.SetOnInvalid(func(error) {
			t.Fatal("unexpected call to onInvalid")
		})

		msg := &PushDataV3MarketWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3MarketWrapper_PublicIncreaseDepthsBatch{
				PublicIncreaseDepthsBatch: nil,
			},
		}

		sub.handleEvent(msg)
	})

	t.Run("symbol is nil", func(t *testing.T) {
		sub := NewDiffDepthBatchSub("BTCUSDT", func(*DepthBatch) {
			t.Fatal("unexpected call to onData")
		}).(*diffDepthBatchSub)

		sub.SetOnInvalid(func(error) {
			t.Fatal("unexpected call to onInvalid")
		})

		msg := &PushDataV3MarketWrapper{
			Symbol: nil,
			Body: &PushDataV3MarketWrapper_PublicIncreaseDepthsBatch{
				PublicIncreaseDepthsBatch: &PublicIncreaseDepthsBatchV3Api{},
			},
		}

		sub.handleEvent(msg)
	})

	t.Run("calls onInvalid for invalid input", func(t *testing.T) {
		var invalidCalled bool

		sub := NewDiffDepthBatchSub("BTCUSDT", func(*DepthBatch) {}).(*diffDepthBatchSub)
		sub.SetOnInvalid(func(error) {
			invalidCalled = true
		})

		msg := &PushDataV3MarketWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3MarketWrapper_PublicIncreaseDepthsBatch{
				PublicIncreaseDepthsBatch: &PublicIncreaseDepthsBatchV3Api{
					Items: []*PublicIncreaseDepthsV3Api{
						{
							Asks: []*PublicIncreaseDepthV3ApiItem{
								{
									Price:    "8329.93",
									Quantity: "82938.9",
								},
							},
							Bids:    []*PublicIncreaseDepthV3ApiItem{},
							Version: "INVALID",
						},
					},
					EventType: "public",
				},
			},
		}

		sub.handleEvent(msg)
		assert.True(t, invalidCalled, "onInvalid should be called")
	})
}

func Test_mapProtoDepthBatch(t *testing.T) {
	t.Run("valid depth batch", func(t *testing.T) {
		sendTime := int64(1234567890)

		msg := &PublicIncreaseDepthsBatchV3Api{
			Items: []*PublicIncreaseDepthsV3Api{
				{
					Version: "123",
					Asks: []*PublicIncreaseDepthV3ApiItem{
						{Price: "100.2", Quantity: "1.1"},
					},
					Bids: []*PublicIncreaseDepthV3ApiItem{
						{Price: "100.1", Quantity: "2.2"},
					},
				},
			},
		}

		got, err := mapProtoDepthBatch(msg, "BTCUSDT", &sendTime)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Len(t, got.Items, 1)

		assert.Equal(t, "BTCUSDT", got.Symbol)
		assert.Equal(t, &sendTime, got.SendTime)

		item := got.Items[0]
		require.Len(t, item.Bids, 1)
		require.Len(t, item.Asks, 1)

		assert.Equal(t, int64(123), item.Version)

		testutil.AssertDecimalEqual(t, item.Asks[0].Price, "100.2")
		testutil.AssertDecimalEqual(t, item.Asks[0].Quantity, "1.1")
		testutil.AssertDecimalEqual(t, item.Bids[0].Price, "100.1")
		testutil.AssertDecimalEqual(t, item.Bids[0].Quantity, "2.2")
	})

	t.Run("invalid version", func(t *testing.T) {
		msg := &PublicIncreaseDepthsBatchV3Api{
			Items: []*PublicIncreaseDepthsV3Api{
				{
					Version: "INVALID",
					Asks: []*PublicIncreaseDepthV3ApiItem{
						{Price: "100.2", Quantity: "1.1"},
					},
					Bids: []*PublicIncreaseDepthV3ApiItem{
						{Price: "100.1", Quantity: "2.2"},
					},
				},
			},
		}

		_, err := mapProtoDepthBatch(msg, "BTCUSDT", nil)
		require.Error(t, err)

		assert.ErrorContains(t, err, "invalid version")
	})

	t.Run("invalid asks", func(t *testing.T) {
		msg := &PublicIncreaseDepthsBatchV3Api{
			Items: []*PublicIncreaseDepthsV3Api{
				{
					Version: "123",
					Asks: []*PublicIncreaseDepthV3ApiItem{
						{Price: "INVALID", Quantity: "1.1"},
					},
					Bids: []*PublicIncreaseDepthV3ApiItem{
						{Price: "100.1", Quantity: "2.2"},
					},
				},
			},
		}

		_, err := mapProtoDepthBatch(msg, "BTCUSDT", nil)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid asks")
	})

	t.Run("invalid bids", func(t *testing.T) {
		msg := &PublicIncreaseDepthsBatchV3Api{
			Items: []*PublicIncreaseDepthsV3Api{
				{
					Version: "123",
					Asks: []*PublicIncreaseDepthV3ApiItem{
						{Price: "100.2", Quantity: "1.1"},
					},
					Bids: []*PublicIncreaseDepthV3ApiItem{
						{Price: "INVALID", Quantity: "2.2"},
					},
				},
			},
		}

		_, err := mapProtoDepthBatch(msg, "BTCUSDT", nil)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid bids")
	})
}

func Test_mapProtoDepthLevelFromIncreaseDepth(t *testing.T) {
	t.Run("valid depth level", func(t *testing.T) {
		items := []*PublicIncreaseDepthV3ApiItem{
			{Price: "10000.1", Quantity: "0.5"},
		}

		levels, err := mapProtoDepthLevelFromIncreaseDepth(items)
		require.NoError(t, err)

		testutil.AssertDecimalEqual(t, levels[0].Price, "10000.1")
		testutil.AssertDecimalEqual(t, levels[0].Quantity, "0.5")
	})

	t.Run("invalid price", func(t *testing.T) {
		items := []*PublicIncreaseDepthV3ApiItem{
			{Price: "INVALID", Quantity: "1.23"},
		}

		_, err := mapProtoDepthLevelFromIncreaseDepth(items)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid price")
	})

	t.Run("invalid quantity", func(t *testing.T) {
		items := []*PublicIncreaseDepthV3ApiItem{
			{Price: "100.5", Quantity: "INVALID"},
		}

		_, err := mapProtoDepthLevelFromIncreaseDepth(items)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid quantity")
	})
}
