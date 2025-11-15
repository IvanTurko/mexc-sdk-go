package wsuser

import (
	"testing"
	"time"

	"github.com/IvanTurko/mexc-sdk-go/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewSpotAccountUpdateSub(t *testing.T) {
	t.Run("normal flow", func(t *testing.T) {
		expCh := "spot@private.account.v3.api.pb"

		sub := NewSpotAccountUpdateSub(func(PrivateAccountUpdate) {}).(*spotAccountUpdateSub)
		assert.Equal(t, expCh, sub.streamName)
	})
	t.Run("onData is nil", func(t *testing.T) {
		defer func() {
			r := recover()
			assert.Contains(t, r, "onData function is nil")
		}()

		NewSpotAccountUpdateSub(nil)
	})
}

func TestSpotAccountUpdateSub_matches(t *testing.T) {
	sub := NewSpotAccountUpdateSub(func(PrivateAccountUpdate) {}).(*spotAccountUpdateSub)

	msg := &message{
		Msg: sub.streamName,
	}

	ok, _ := sub.matches(msg)
	assert.True(t, ok)
}

func TestSpotAccountUpdateSub_acceptEvent(t *testing.T) {
	sub := NewSpotAccountUpdateSub(func(PrivateAccountUpdate) {}).(*spotAccountUpdateSub)

	msg := &PushDataV3UserWrapper{
		Channel: sub.streamName,
	}
	assert.True(t, sub.acceptEvent(msg))
}

func TestSpotAccountUpdateSub_handleEvent(t *testing.T) {
	t.Run("calls onData for valid input", func(t *testing.T) {
		var called bool

		sub := NewSpotAccountUpdateSub(func(PrivateAccountUpdate) {
			called = true
		}).(*spotAccountUpdateSub)

		msg := &PushDataV3UserWrapper{
			Body: &PushDataV3UserWrapper_PrivateAccount{
				PrivateAccount: &PrivateAccountV3Api{
					VcoinName:           "BTC",
					CoinId:              "1",
					BalanceAmount:       "0.1234",
					BalanceAmountChange: "0.01",
					FrozenAmount:        "0.002",
					FrozenAmountChange:  "-0.001",
					Type:                "deposit",
					Time:                1234567890,
				},
			},
		}

		sub.handleEvent(msg)
		assert.True(t, called, "onData should be called")
	})

	t.Run("PrivateAccount is nil", func(t *testing.T) {
		sub := NewSpotAccountUpdateSub(func(PrivateAccountUpdate) {
			t.Fatal("unexpected call to onData")
		}).(*spotAccountUpdateSub)

		sub.SetOnInvalid(func(error) {
			t.Fatal("unexpected call to onInvalid")
		})

		msg := &PushDataV3UserWrapper{
			Body: &PushDataV3UserWrapper_PrivateAccount{
				PrivateAccount: nil,
			},
		}

		sub.handleEvent(msg)
	})

	t.Run("calls onInvalid for invalid input", func(t *testing.T) {
		var invalidCalled bool

		sub := NewSpotAccountUpdateSub(func(PrivateAccountUpdate) {}).(*spotAccountUpdateSub)
		sub.SetOnInvalid(func(err error) {
			invalidCalled = true
		})

		msg := &PushDataV3UserWrapper{
			Body: &PushDataV3UserWrapper_PrivateAccount{
				PrivateAccount: &PrivateAccountV3Api{
					VcoinName:           "BTC",
					CoinId:              "1",
					BalanceAmount:       "INVALID",
					BalanceAmountChange: "0.01",
					FrozenAmount:        "0.002",
					FrozenAmountChange:  "-0.001",
					Type:                "deposit",
					Time:                1234567890,
				},
			},
		}

		sub.handleEvent(msg)
		assert.True(t, invalidCalled, "onInvalid should be called")
	})
}

func Test_mapProtoPrivateAccount(t *testing.T) {
	sendTime := time.Now().Unix()

	t.Run("valid private account update", func(t *testing.T) {
		msg := &PrivateAccountV3Api{
			VcoinName:           "BTC",
			CoinId:              "1",
			BalanceAmount:       "123.45",
			BalanceAmountChange: "1.23",
			FrozenAmount:        "10.01",
			FrozenAmountChange:  "0.1",
			Type:                "test",
			Time:                1234567,
		}

		update, err := mapProtoPrivateAccountUpdate(msg, &sendTime)
		assert.NoError(t, err)

		assert.Equal(t, "BTC", update.VcoinName)
		assert.Equal(t, "1", update.CoinID)
		assert.Equal(t, "test", update.Type)
		assert.Equal(t, int64(1234567), update.Time)
		assert.Equal(t, &sendTime, update.SendTime)

		testutil.AssertDecimalEqual(t, update.BalanceAmount, "123.45")
		testutil.AssertDecimalEqual(t, update.BalanceAmountChange, "1.23")
		testutil.AssertDecimalEqual(t, update.FrozenAmount, "10.01")
		testutil.AssertDecimalEqual(t, update.FrozenAmountChange, "0.1")
	})

	t.Run("invalid balanceAmount", func(t *testing.T) {
		msg := &PrivateAccountV3Api{
			VcoinName:           "BTC",
			CoinId:              "1",
			BalanceAmount:       "INVALID",
			BalanceAmountChange: "1.23",
			FrozenAmount:        "10.01",
			FrozenAmountChange:  "0.1",
			Type:                "test",
			Time:                1234567,
		}

		_, err := mapProtoPrivateAccountUpdate(msg, &sendTime)

		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid balanceAmount")
	})

	t.Run("invalid balanceAmountChange", func(t *testing.T) {
		msg := &PrivateAccountV3Api{
			VcoinName:           "BTC",
			CoinId:              "1",
			BalanceAmount:       "123.45",
			BalanceAmountChange: "INVALID",
			FrozenAmount:        "10.01",
			FrozenAmountChange:  "0.1",
			Type:                "test",
			Time:                1234567,
		}

		_, err := mapProtoPrivateAccountUpdate(msg, &sendTime)

		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid balanceAmountChange")
	})

	t.Run("invalid frozenAmount", func(t *testing.T) {
		msg := &PrivateAccountV3Api{
			VcoinName:           "BTC",
			CoinId:              "1",
			BalanceAmount:       "123.45",
			BalanceAmountChange: "1.23",
			FrozenAmount:        "INVALID",
			FrozenAmountChange:  "0.1",
			Type:                "test",
			Time:                1234567,
		}

		_, err := mapProtoPrivateAccountUpdate(msg, &sendTime)

		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid frozenAmount")
	})

	t.Run("invalid frozenAmountChange", func(t *testing.T) {
		msg := &PrivateAccountV3Api{
			VcoinName:           "BTC",
			CoinId:              "1",
			BalanceAmount:       "123.45",
			BalanceAmountChange: "1.23",
			FrozenAmount:        "10.01",
			FrozenAmountChange:  "INVALID",
			Type:                "test",
			Time:                1234567,
		}

		_, err := mapProtoPrivateAccountUpdate(msg, &sendTime)

		require.Error(t, err)
		assert.ErrorContains(t, err, "invalid frozenAmountChange")
	})
}
