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

func Test_NewKlineSub(t *testing.T) {
	t.Run("normal flow", func(t *testing.T) {
		symbol := "BTCUSDT"
		interval := Kline5Min
		expCh := fmt.Sprintf("spot@public.kline.v3.api.pb@%s@%s", symbol, interval)

		sub := NewKlineSub(symbol, interval, func(Kline) {}).(*klineSub)
		assert.Equal(t, expCh, sub.streamName)
	})
	t.Run("invalid symbol name", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "invalid symbol name")
		}()

		NewKlineSub("", Kline5Min, func(Kline) {})
	})
	t.Run("invalid kline interval", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "invalid kline interval")
		}()

		NewKlineSub("BTCUSDT", KlineInterval("hello"), func(Kline) {})
	})
	t.Run("onData is nil", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "onData function is nil")
		}()

		NewKlineSub("BTCUSDT", Kline5Min, nil)
	})
}

func TestKlineSub_matches(t *testing.T) {
	sub := NewKlineSub("BTCUSDT", Kline5Min, func(Kline) {}).(*klineSub)

	msg := &message{
		Msg: sub.streamName,
	}

	ok, _ := sub.matches(msg)
	assert.True(t, ok)
}

func TestKlineSub_acceptEvent(t *testing.T) {
	sub := NewKlineSub("BTCUSDT", Kline5Min, func(Kline) {}).(*klineSub)

	msg := &PushDataV3MarketWrapper{
		Channel: sub.streamName,
	}
	assert.True(t, sub.acceptEvent(msg))
}

func TestKlineSub_handleEvent(t *testing.T) {
	t.Run("calls onData for valid input", func(t *testing.T) {
		var called bool

		sub := NewKlineSub("BTCUSDT", Kline5Min, func(Kline) {
			called = true
		}).(*klineSub)

		msg := &PushDataV3MarketWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3MarketWrapper_PublicSpotKline{
				PublicSpotKline: &PublicSpotKlineV3Api{
					Interval:     "Min5",
					WindowStart:  122930,
					OpeningPrice: "9303.49",
					ClosingPrice: "3920.30",
					HighestPrice: "10000.90",
					LowestPrice:  "830.39",
					Volume:       "12902",
					Amount:       "182928",
					WindowEnd:    1290290,
				},
			},
		}

		sub.handleEvent(msg)
		assert.True(t, called, "onData should be called")
	})

	t.Run("PublicSpotKline is nil", func(t *testing.T) {
		sub := NewKlineSub("BTCUSDT", Kline5Min, func(Kline) {
			t.Fatal("unexpected call to onData")
		}).(*klineSub)

		sub.SetOnInvalid(func(error) {
			t.Fatal("unexpected call to onInvalid")
		})

		msg := &PushDataV3MarketWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3MarketWrapper_PublicSpotKline{
				PublicSpotKline: nil,
			},
		}

		sub.handleEvent(msg)
	})

	t.Run("symbol is nil", func(t *testing.T) {
		sub := NewKlineSub("BTCUSDT", Kline5Min, func(Kline) {
			t.Fatal("unexpected call to onData")
		}).(*klineSub)

		sub.SetOnInvalid(func(error) {
			t.Fatal("unexpected call to onInvalid")
		})

		msg := &PushDataV3MarketWrapper{
			Symbol: nil,
			Body: &PushDataV3MarketWrapper_PublicSpotKline{
				PublicSpotKline: &PublicSpotKlineV3Api{},
			},
		}

		sub.handleEvent(msg)
	})

	t.Run("calls onInvalid for invalid input", func(t *testing.T) {
		var invalidCalled bool

		sub := NewKlineSub("BTCUSDT", Kline5Min, func(Kline) {}).(*klineSub)
		sub.SetOnInvalid(func(error) {
			invalidCalled = true
		})

		msg := &PushDataV3MarketWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3MarketWrapper_PublicSpotKline{
				PublicSpotKline: &PublicSpotKlineV3Api{
					Interval:     "INVALID",
					WindowStart:  122930,
					OpeningPrice: "9303.49",
					ClosingPrice: "3920.30",
					HighestPrice: "10000.90",
					LowestPrice:  "830.39",
					Volume:       "12902",
					Amount:       "182928",
					WindowEnd:    1290290,
				},
			},
		}

		sub.handleEvent(msg)
		assert.True(t, invalidCalled, "onInvalid should be called")
	})
}

func Test_mapProtoKline(t *testing.T) {
	sendTime := time.Now().Unix()
	symbol := "BTCUSDT"

	t.Run("valid kline", func(t *testing.T) {
		msg := &PublicSpotKlineV3Api{
			Interval:     "Min5",
			WindowStart:  111,
			WindowEnd:    222,
			OpeningPrice: "100.1",
			ClosingPrice: "102.5",
			HighestPrice: "103.0",
			LowestPrice:  "99.9",
			Volume:       "1000.0",
			Amount:       "100000.0",
		}

		kline, err := mapProtoKline(msg, symbol, &sendTime)
		require.NoError(t, err)

		assert.Equal(t, symbol, kline.Symbol)
		assert.Equal(t, Kline5Min, kline.Interval)
		assert.Equal(t, int64(111), kline.WindowStart)
		assert.Equal(t, int64(222), kline.WindowEnd)
		assert.Equal(t, &sendTime, kline.SendTime)

		testutil.AssertDecimalEqual(t, kline.OpeningPrice, "100.1")
		testutil.AssertDecimalEqual(t, kline.ClosingPrice, "102.5")
		testutil.AssertDecimalEqual(t, kline.HighestPrice, "103.0")
		testutil.AssertDecimalEqual(t, kline.LowestPrice, "99.9")
		testutil.AssertDecimalEqual(t, kline.Volume, "1000.0")
		testutil.AssertDecimalEqual(t, kline.Amount, "100000.0")
	})

	t.Run("invalid interval", func(t *testing.T) {
		msg := &PublicSpotKlineV3Api{
			Interval:     "INVALID",
			WindowStart:  111,
			WindowEnd:    222,
			OpeningPrice: "100.1",
			ClosingPrice: "102.5",
			HighestPrice: "103.0",
			LowestPrice:  "99.9",
			Volume:       "1000.0",
			Amount:       "100000.0",
		}

		_, err := mapProtoKline(msg, symbol, &sendTime)

		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid interval")
	})

	t.Run("invalid openingPrice", func(t *testing.T) {
		msg := &PublicSpotKlineV3Api{
			Interval:     "Min5",
			WindowStart:  111,
			WindowEnd:    222,
			OpeningPrice: "INVALID",
			ClosingPrice: "102.5",
			HighestPrice: "103.0",
			LowestPrice:  "99.9",
			Volume:       "1000.0",
			Amount:       "100000.0",
		}

		_, err := mapProtoKline(msg, symbol, &sendTime)

		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid openingPrice")
	})

	t.Run("invalid closingPrice", func(t *testing.T) {
		msg := &PublicSpotKlineV3Api{
			Interval:     "Min5",
			WindowStart:  111,
			WindowEnd:    222,
			OpeningPrice: "100.1",
			ClosingPrice: "INVALID",
			HighestPrice: "103.0",
			LowestPrice:  "99.9",
			Volume:       "1000.0",
			Amount:       "100000.0",
		}

		_, err := mapProtoKline(msg, symbol, &sendTime)

		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid closingPrice")
	})

	t.Run("invalid highestPrice", func(t *testing.T) {
		msg := &PublicSpotKlineV3Api{
			Interval:     "Min5",
			WindowStart:  111,
			WindowEnd:    222,
			OpeningPrice: "100.1",
			ClosingPrice: "102.5",
			HighestPrice: "INVALID",
			LowestPrice:  "99.9",
			Volume:       "1000.0",
			Amount:       "100000.0",
		}

		_, err := mapProtoKline(msg, symbol, &sendTime)

		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid highestPrice")
	})

	t.Run("invalid lowestPrice", func(t *testing.T) {
		msg := &PublicSpotKlineV3Api{
			Interval:     "Min5",
			WindowStart:  111,
			WindowEnd:    222,
			OpeningPrice: "100.1",
			ClosingPrice: "102.5",
			HighestPrice: "103.0",
			LowestPrice:  "INVALID",
			Volume:       "1000.0",
			Amount:       "100000.0",
		}

		_, err := mapProtoKline(msg, symbol, &sendTime)

		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid lowestPrice")
	})

	t.Run("invalid volume", func(t *testing.T) {
		msg := &PublicSpotKlineV3Api{
			Interval:     "Min5",
			WindowStart:  111,
			WindowEnd:    222,
			OpeningPrice: "100.1",
			ClosingPrice: "102.5",
			HighestPrice: "103.0",
			LowestPrice:  "99.9",
			Volume:       "INVALID",
			Amount:       "100000.0",
		}

		_, err := mapProtoKline(msg, symbol, &sendTime)

		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid volume")
	})

	t.Run("invalid amount", func(t *testing.T) {
		msg := &PublicSpotKlineV3Api{
			Interval:     "Min5",
			WindowStart:  111,
			WindowEnd:    222,
			OpeningPrice: "100.1",
			ClosingPrice: "102.5",
			HighestPrice: "103.0",
			LowestPrice:  "99.9",
			Volume:       "1000.0",
			Amount:       "INVALID",
		}

		_, err := mapProtoKline(msg, symbol, &sendTime)

		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid amount")
	})
}
