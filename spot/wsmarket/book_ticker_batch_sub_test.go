package wsmarket

import (
	"fmt"
	"testing"

	"github.com/IvanTurko/mexc-sdk-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func Test_NewBookTickerBatchSub(t *testing.T) {
	t.Run("normal flow", func(t *testing.T) {
		symbol := "BTCUSDT"
		expCh := fmt.Sprintf("spot@public.bookTicker.batch.v3.api.pb@%s", symbol)

		sub := NewBookTickerBatchSub(symbol, func(*BookTickerBatch) {}).(*bookTickerBatchSub)
		assert.Equal(t, expCh, sub.streamName)
	})
	t.Run("invalid symbol name", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "invalid symbol name")
		}()

		NewBookTickerBatchSub("", func(*BookTickerBatch) {})
	})
	t.Run("onData is nil", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "onData function is nil")
		}()

		NewBookTickerBatchSub("BTCUSDT", nil)
	})
}

func TestBookTickerBatchSub_matches(t *testing.T) {
	sub := NewBookTickerBatchSub("BTCUSDT", func(*BookTickerBatch) {}).(*bookTickerBatchSub)

	msg := &message{
		Msg: sub.streamName,
	}

	ok, _ := sub.matches(msg)
	assert.True(t, ok)
}

func TestBookTickerBatchSub_acceptEvent(t *testing.T) {
	sub := NewBookTickerBatchSub("BTCUSDT", func(*BookTickerBatch) {}).(*bookTickerBatchSub)

	msg := &PushDataV3MarketWrapper{
		Channel: sub.streamName,
	}
	assert.True(t, sub.acceptEvent(msg))
}

func TestBookTickerBatchSub_handleEvent(t *testing.T) {
	t.Run("calls onData for valid input", func(t *testing.T) {
		var called bool

		sub := NewBookTickerBatchSub("BTCUSDT", func(*BookTickerBatch) {
			called = true
		}).(*bookTickerBatchSub)

		msg := &PushDataV3MarketWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3MarketWrapper_PublicBookTickerBatch{
				PublicBookTickerBatch: &PublicBookTickerBatchV3Api{
					Items: []*PublicBookTickerV3Api{
						{
							BidPrice:    "3902.90",
							BidQuantity: "9302.9",
							AskPrice:    "9329.93",
							AskQuantity: "2939.93",
						},
					},
				},
			},
		}

		sub.handleEvent(msg)
		assert.True(t, called, "onData should be called")
	})

	t.Run("PublicBookTickerBatch is nil", func(t *testing.T) {
		sub := NewBookTickerBatchSub("BTCUSDT", func(*BookTickerBatch) {
			t.Fatal("unexpected call to onData")
		}).(*bookTickerBatchSub)

		sub.SetOnInvalid(func(error) {
			t.Fatal("unexpected call to onInvalid")
		})

		msg := &PushDataV3MarketWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3MarketWrapper_PublicBookTickerBatch{
				PublicBookTickerBatch: nil,
			},
		}

		sub.handleEvent(msg)
	})

	t.Run("symbol is nil", func(t *testing.T) {
		sub := NewBookTickerBatchSub("BTCUSDT", func(*BookTickerBatch) {
			t.Fatal("unexpected call to onData")
		}).(*bookTickerBatchSub)

		sub.SetOnInvalid(func(error) {
			t.Fatal("unexpected call to onInvalid")
		})

		msg := &PushDataV3MarketWrapper{
			Symbol: nil,
			Body: &PushDataV3MarketWrapper_PublicBookTickerBatch{
				PublicBookTickerBatch: &PublicBookTickerBatchV3Api{},
			},
		}

		sub.handleEvent(msg)
	})

	t.Run("calls onInvalid for invalid input", func(t *testing.T) {
		var invalidCalled bool

		sub := NewBookTickerBatchSub("BTCUSDT", func(*BookTickerBatch) {}).(*bookTickerBatchSub)
		sub.SetOnInvalid(func(error) {
			invalidCalled = true
		})

		msg := &PushDataV3MarketWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3MarketWrapper_PublicBookTickerBatch{
				PublicBookTickerBatch: &PublicBookTickerBatchV3Api{
					Items: []*PublicBookTickerV3Api{
						{
							BidPrice:    "INVALID",
							BidQuantity: "9302.9",
							AskPrice:    "9329.93",
							AskQuantity: "2939.93",
						},
					},
				},
			},
		}

		sub.handleEvent(msg)
		assert.True(t, invalidCalled, "onInvalid should be called")
	})
}

func Test_mapProtoBookTickerBatch(t *testing.T) {
	t.Run("valid book ticker batch", func(t *testing.T) {
		sendTime := int64(1234567890)

		msg := &PublicBookTickerBatchV3Api{
			Items: []*PublicBookTickerV3Api{
				{
					BidPrice:    "100.1",
					BidQuantity: "5.5",
					AskPrice:    "100.2",
					AskQuantity: "6.6",
				},
			},
		}

		got, err := mapProtoBookTickerBatch(msg, "BTCUSDT", &sendTime)
		require.NoError(t, err)
		require.Len(t, got.Tickers, 1)

		assert.Equal(t, "BTCUSDT", got.Symbol)
		assert.Equal(t, &sendTime, got.SendTime)

		ticker := got.Tickers[0]
		testutil.AssertDecimalEqual(t, ticker.BidPrice, "100.1")
		testutil.AssertDecimalEqual(t, ticker.BidQuantity, "5.5")
		testutil.AssertDecimalEqual(t, ticker.AskPrice, "100.2")
		testutil.AssertDecimalEqual(t, ticker.AskQuantity, "6.6")
	})

	t.Run("invalid bidPrice", func(t *testing.T) {
		msg := &PublicBookTickerBatchV3Api{
			Items: []*PublicBookTickerV3Api{
				{
					BidPrice:    "INVALID",
					BidQuantity: "5.5",
					AskPrice:    "100.2",
					AskQuantity: "6.6",
				},
			},
		}

		_, err := mapProtoBookTickerBatch(msg, "BTCUSDT", nil)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid bidPrice")
	})

	t.Run("invalid bidQuantity", func(t *testing.T) {
		msg := &PublicBookTickerBatchV3Api{
			Items: []*PublicBookTickerV3Api{
				{
					BidPrice:    "100.1",
					BidQuantity: "INVALID",
					AskPrice:    "100.2",
					AskQuantity: "6.6",
				},
			},
		}

		_, err := mapProtoBookTickerBatch(msg, "BTCUSDT", nil)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid bidQuantity")
	})

	t.Run("invalid askPrice", func(t *testing.T) {
		msg := &PublicBookTickerBatchV3Api{
			Items: []*PublicBookTickerV3Api{
				{
					BidPrice:    "100.1",
					BidQuantity: "5.5",
					AskPrice:    "INVALID",
					AskQuantity: "6.6",
				},
			},
		}

		_, err := mapProtoBookTickerBatch(msg, "BTCUSDT", nil)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid askPrice")
	})

	t.Run("invalid askQuantity", func(t *testing.T) {
		msg := &PublicBookTickerBatchV3Api{
			Items: []*PublicBookTickerV3Api{
				{
					BidPrice:    "100.1",
					BidQuantity: "5.5",
					AskPrice:    "100.2",
					AskQuantity: "INVALID",
				},
			},
		}

		_, err := mapProtoBookTickerBatch(msg, "BTCUSDT", nil)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid askQuantity")
	})
}
