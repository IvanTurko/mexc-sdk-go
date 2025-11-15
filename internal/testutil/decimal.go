package testutil

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func AssertDecimalEqual(t *testing.T, actual decimal.Decimal, expectedStr string, msgAndArgs ...any) {
	expected, err := decimal.NewFromString(expectedStr)
	assert.NoError(t, err, msgAndArgs...)
	assert.True(t, actual.Equal(expected), msgAndArgs...)
}
