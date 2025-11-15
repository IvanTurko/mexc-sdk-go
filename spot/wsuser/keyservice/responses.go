package keyservice

import (
	"encoding/json"
	"fmt"

	"github.com/IvanTurko/mexc-sdk-go/sdkerr"
	"github.com/IvanTurko/mexc-sdk-go/spot/errs"
)

type responseErr struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

func checkResponseError(status int, body []byte) error {
	if status >= 200 && status < 300 {
		return nil
	}
	var respErr responseErr
	if err := json.Unmarshal(body, &respErr); err != nil {
		return fmt.Errorf("http status %d: %s", status, string(body))
	}
	return errs.ErrorCode(respErr.Code)
}

func decodeResponse[T any](data []byte, op string) (*T, error) {
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, sdkerr.NewSDKError().
			WithSubsys(subsys).
			WithOp(op).
			WithKind(sdkerr.ErrDecodeError).
			WithCause(err)
	}
	return &result, nil
}
