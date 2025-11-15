package wsuser

import (
	"testing"
	"time"

	"github.com/IvanTurko/mexc-sdk-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func Test_NewSpotAccountDealsSub(t *testing.T) {
	t.Run("normal flow", func(t *testing.T) {
		expCh := "spot@private.deals.v3.api.pb"

		sub := NewSpotAccountDealsSub(func(PrivateDeal) {}).(*spotAccountDealsSub)
		assert.Equal(t, expCh, sub.streamName)
	})
	t.Run("onData is nil", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "onData function is nil")
		}()

		NewSpotAccountDealsSub(nil)
	})
}

func TestSpotAccountDealsSub_matches(t *testing.T) {
	sub := NewSpotAccountDealsSub(func(PrivateDeal) {}).(*spotAccountDealsSub)

	msg := &message{
		Msg: sub.streamName,
	}

	ok, _ := sub.matches(msg)
	assert.True(t, ok)
}

func TestSpotAccountDealsSub_acceptEvent(t *testing.T) {
	sub := NewSpotAccountDealsSub(func(PrivateDeal) {}).(*spotAccountDealsSub)

	msg := &PushDataV3UserWrapper{
		Channel: sub.streamName,
	}
	assert.True(t, sub.acceptEvent(msg))
}

func TestSpotAccountDealsSub_handleEvent(t *testing.T) {
	t.Run("calls onData for valid input", func(t *testing.T) {
		var called bool

		sub := NewSpotAccountDealsSub(func(PrivateDeal) {
			called = true
		}).(*spotAccountDealsSub)

		msg := &PushDataV3UserWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3UserWrapper_PrivateDeals{
				PrivateDeals: &PrivateDealsV3Api{
					Price:         "930293",
					Quantity:      "9393",
					Amount:        "19292",
					TradeType:     1,
					IsMaker:       false,
					IsSelfTrade:   false,
					TradeId:       "4949l",
					ClientOrderId: "01202l",
					OrderId:       "3010k",
					FeeAmount:     "49290.9",
					FeeCurrency:   "KIn",
					Time:          2939930,
				},
			},
		}

		sub.handleEvent(msg)
		assert.True(t, called, "onData should be called")
	})

	t.Run("PrivateDeals is nil", func(t *testing.T) {
		sub := NewSpotAccountDealsSub(func(PrivateDeal) {
			t.Fatal("unexpected call to onData")
		}).(*spotAccountDealsSub)

		sub.SetOnInvalid(func(error) {
			t.Fatal("unexpected call to onInvalid")
		})

		msg := &PushDataV3UserWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3UserWrapper_PrivateDeals{
				PrivateDeals: nil,
			},
		}

		sub.handleEvent(msg)
	})

	t.Run("symbol is nil", func(t *testing.T) {
		sub := NewSpotAccountDealsSub(func(PrivateDeal) {}).(*spotAccountDealsSub)

		sub.SetOnInvalid(func(error) {
			t.Fatal("unexpected call to onInvalid")
		})

		msg := &PushDataV3UserWrapper{
			Symbol: nil,
			Body:   &PushDataV3UserWrapper_PrivateDeals{},
		}

		sub.handleEvent(msg)
	})

	t.Run("calls onInvalid for invalid input", func(t *testing.T) {
		var invalidCalled bool

		sub := NewSpotAccountDealsSub(func(PrivateDeal) {}).(*spotAccountDealsSub)
		sub.SetOnInvalid(func(error) {
			invalidCalled = true
		})

		msg := &PushDataV3UserWrapper{
			Symbol: proto.String("BTCUSDT"),
			Body: &PushDataV3UserWrapper_PrivateDeals{
				PrivateDeals: &PrivateDealsV3Api{
					Price:         "930293",
					Quantity:      "9393",
					Amount:        "INVALID",
					TradeType:     1,
					IsMaker:       false,
					IsSelfTrade:   false,
					TradeId:       "4949l",
					ClientOrderId: "01202l",
					OrderId:       "3010k",
					FeeAmount:     "49290.9",
					FeeCurrency:   "KIn",
					Time:          2939930,
				},
			},
		}

		sub.handleEvent(msg)
		assert.True(t, invalidCalled, "onInvalid should be called")
	})
}

func Test_mapProtoPrivateDeal_Valid(t *testing.T) {
	sendTime := time.Now().Unix()
	symbol := "BTCUSDT"

	msg := &PrivateDealsV3Api{
		Price:         "12345.67",
		Quantity:      "0.01",
		Amount:        "123.4567",
		TradeType:     1,
		TradeId:       "trade123",
		OrderId:       "order456",
		ClientOrderId: "client789",
		FeeAmount:     "0.1",
		FeeCurrency:   "USDT",
		Time:          1234567890,
		IsMaker:       true,
		IsSelfTrade:   false,
	}

	deal, err := mapProtoPrivateDeal(msg, symbol, &sendTime)
	assert.NoError(t, err)

	assert.Equal(t, symbol, deal.Symbol)
	assert.Equal(t, "trade123", deal.TradeId)
	assert.Equal(t, "order456", deal.OrderId)
	assert.Equal(t, "client789", deal.ClientOrderId)
	assert.Equal(t, "USDT", deal.FeeCurrency)
	assert.Equal(t, int64(1234567890), deal.Time)
	assert.Equal(t, true, deal.IsMaker)
	assert.Equal(t, false, deal.IsSelfTrade)
	assert.Equal(t, &sendTime, deal.SendTime)

	assert.Equal(t, TradeSideBuy, deal.TradeSide)

	testutil.AssertDecimalEqual(t, deal.Price, "12345.67")
	testutil.AssertDecimalEqual(t, deal.Quantity, "0.01")
	testutil.AssertDecimalEqual(t, deal.Amount, "123.4567")
	testutil.AssertDecimalEqual(t, deal.FeeAmount, "0.1")
}

func Test_mapProtoPrivateDeal_InvalidFields(t *testing.T) {
	sendTime := int64(1234567890)
	symbol := "BTCUSDT"

	validMsg := func() *PrivateDealsV3Api {
		return &PrivateDealsV3Api{
			Price:         "12345.67",
			Quantity:      "0.01",
			Amount:        "123.4567",
			TradeType:     1,
			TradeId:       "trade123",
			OrderId:       "order456",
			ClientOrderId: "client789",
			FeeAmount:     "0.1",
			FeeCurrency:   "USDT",
			Time:          1234567890,
			IsMaker:       true,
			IsSelfTrade:   false,
		}
	}

	t.Run("invalid price", func(t *testing.T) {
		msg := validMsg()
		msg.Price = "INVALID"
		_, err := mapProtoPrivateDeal(msg, symbol, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid price")
	})

	t.Run("invalid quantity", func(t *testing.T) {
		msg := validMsg()
		msg.Quantity = "INVALID"
		_, err := mapProtoPrivateDeal(msg, symbol, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid quantity")
	})

	t.Run("invalid amount", func(t *testing.T) {
		msg := validMsg()
		msg.Amount = "INVALID"
		_, err := mapProtoPrivateDeal(msg, symbol, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid amount")
	})

	t.Run("unknown trade side", func(t *testing.T) {
		msg := validMsg()
		msg.TradeType = 999
		_, err := mapProtoPrivateDeal(msg, symbol, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "unknown trade side")
	})

	t.Run("invalid feeAmount", func(t *testing.T) {
		msg := validMsg()
		msg.FeeAmount = "INVALID"
		_, err := mapProtoPrivateDeal(msg, symbol, &sendTime)
		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid feeAmount")
	})
}
