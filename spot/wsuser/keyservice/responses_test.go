package keyservice

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckResponseError_OK(t *testing.T) {
	err := checkResponseError(http.StatusOK, []byte(`{}`))
	assert.NoError(t, err)
}

func TestCheckResponseError_KnownErrorCode_ReturnsExpectedMessage(t *testing.T) {
	body := []byte(`{"code": -2011, "msg": "Unknown order sent"}`)
	err := checkResponseError(400, body)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Unknown order sent")
}

func TestCheckResponseError_UnknownErrorCode_ReturnsGenericMessage(t *testing.T) {
	body := []byte(`{"code": -1001, "msg": "some error"}`)
	err := checkResponseError(400, body)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown error code: -1001")
}

func TestCheckResponseError_HTTPErrorWithInvalidJSON(t *testing.T) {
	body := []byte(`not-json`)
	err := checkResponseError(500, body)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "http status 500")
}

func TestDecodeResponse_Success(t *testing.T) {
	type Result struct {
		Name string `json:"name"`
	}
	data := []byte(`{"name": "test"}`)
	result, err := decodeResponse[Result](data, "TestOp")
	assert.NoError(t, err)
	assert.Equal(t, "test", result.Name)
}

func TestDecodeResponse_InvalidJSON(t *testing.T) {
	data := []byte(`{"name":`) // malformed
	_, err := decodeResponse[struct{ Name string }](data, "DecodeFail")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DecodeFail")
}
