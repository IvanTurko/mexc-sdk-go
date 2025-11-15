package errs

import "fmt"

// ErrorCode represents an error code.
type ErrorCode int

// Common & HTTP-level errors (7)
const (
	ErrUnauthorized         = ErrorCode(401)
	ErrInternalError        = ErrorCode(500)
	ErrSystemBusy           = ErrorCode(501)
	ErrUnknownRequestSource = ErrorCode(506)
	ErrTooManyRequests      = ErrorCode(510)
	ErrEndpointInaccessible = ErrorCode(511)
	ErrPublicAbnormal       = ErrorCode(9999)
)

// Auth & Signature errors (8)
const (
	ErrAPIKeyExpired                   = ErrorCode(402)
	ErrIPNotInWhitelist                = ErrorCode(406)
	ErrInvalidRequestTime              = ErrorCode(513)
	ErrVerifyFailed                    = ErrorCode(602)
	ErrAccountReadPermissionRequired   = ErrorCode(701)
	ErrAccountModifyPermissionRequired = ErrorCode(702)
	ErrTradeReadPermissionRequired     = ErrorCode(703)
	ErrTradeModifyPermissionRequired   = ErrorCode(704)
)

// User & account errors (1)
const (
	ErrAccountDoesNotExist = ErrorCode(1000)
)

// Symbol, pair & market errors (6)
const (
	ErrContractDoesNotExist = ErrorCode(1001)
	ErrContractNotActivated = ErrorCode(1002)
	ErrUnsupportedCurrency  = ErrorCode(4001)
	ErrTradingForbidden     = ErrorCode(6001)
	ErrOpenForbidden        = ErrorCode(6002)
	ErrPairNotAvailable     = ErrorCode(6005)
)

// Balance & amount related errors (2)
const (
	ErrAmountError         = ErrorCode(1004)
	ErrBalanceInsufficient = ErrorCode(2005)
)

// Parameter & format errors (4)
const (
	ErrParameterError      = ErrorCode(600)
	ErrDataDecodingError   = ErrorCode(601)
	ErrRepeatedRequests    = ErrorCode(603)
	ErrPairAndStatusNeeded = ErrorCode(6004)
)

// Trading errors (43)
const (
	ErrRiskLimitLevelError             = ErrorCode(1003)
	ErrWrongOrderDirection             = ErrorCode(2001)
	ErrWrongOpeningType                = ErrorCode(2002)
	ErrOverpricedToPay                 = ErrorCode(2003)
	ErrLowPriceForSelling              = ErrorCode(2004)
	ErrLeverageRatioError              = ErrorCode(2006)
	ErrOrderPriceError                 = ErrorCode(2007)
	ErrQuantityInsufficient            = ErrorCode(2008)
	ErrPositionsDoNotExist             = ErrorCode(2009)
	ErrOrderQuantityError              = ErrorCode(2011)
	ErrCancelOrdersOverMaximumLimit    = ErrorCode(2013)
	ErrBatchOrderQuantityExceedsLimit  = ErrorCode(2014)
	ErrPriceOrQuantityAccuracyError    = ErrorCode(2015)
	ErrTriggerVolumeOverMaximum        = ErrorCode(2016)
	ErrExceedingMaxAvailableMargin     = ErrorCode(2018)
	ErrActiveOpenPositionExists        = ErrorCode(2019)
	ErrLeverageNotConsistent           = ErrorCode(2021)
	ErrWrongPositionType               = ErrorCode(2022)
	ErrPositionsOverMaxLeverage        = ErrorCode(2023)
	ErrOrdersWithLeverageOverMax       = ErrorCode(2024)
	ErrHoldingPositionsOverMax         = ErrorCode(2025)
	ErrLeverageModificationUnsupported = ErrorCode(2026)
	ErrOnlyOneCrossOrIsolatedAllowed   = ErrorCode(2027)
	ErrMaxOrderQuantityExceeded        = ErrorCode(2028)
	ErrErrorOrderType                  = ErrorCode(2029)
	ErrExternalOrderIDTooLong          = ErrorCode(2030)
	ErrPositionExceedsRiskLimit        = ErrorCode(2031)
	ErrPriceLessThanLiqPrice           = ErrorCode(2032)
	ErrPriceMoreThanLiqPrice           = ErrorCode(2033)
	ErrBatchQueryQuantityExceeded      = ErrorCode(2034)
	ErrUnsupportedMarketPriceTier      = ErrorCode(2035)
	ErrOrdersMoreThanLimit             = ErrorCode(2036)
	ErrFrequentTransactions            = ErrorCode(2037)
	ErrMaxAllowablePositionExceeded    = ErrorCode(2038)
	ErrTriggerPriceTypeError           = ErrorCode(3001)
	ErrTriggerTypeError                = ErrorCode(3002)
	ErrExecutiveCycleError             = ErrorCode(3003)
	ErrTriggerPriceError               = ErrorCode(3004)
	ErrSLTPPriceCannotBeNone           = ErrorCode(5001)
	ErrSLTPOrderDoesNotExist           = ErrorCode(5002)
	ErrSLTPPriceSettingWrong           = ErrorCode(5003)
	ErrSLTPVolumeMoreThanPosition      = ErrorCode(5004)
	ErrTimeRangeError                  = ErrorCode(6003)
)

var errorCodes = map[int]string{
	401:  "Unauthorized",
	402:  "Api_key expired",
	406:  "Accessed IP is not in the whitelist",
	500:  "Internal error",
	501:  "System busy",
	506:  "Unknown source of request",
	510:  "Excessive frequency of requests",
	511:  "Endpoint inaccessible",
	513:  "Invalid request(for open api serves time more or less than 10s)",
	600:  "Parameter error",
	601:  "Data decoding error",
	602:  "Verify failed",
	603:  "Repeated requests",
	701:  "Account read permission is required",
	702:  "Account modify permission is required",
	703:  "Trade information read permission is required",
	704:  "Transaction information modify permission is required",
	1000: "Account does not exist",
	1001: "Contract does not exist",
	1002: "Contract not activated",
	1003: "Error in risk limit level",
	1004: "Amount error",
	2001: "Wrong order direction",
	2002: "Wrong opening type",
	2003: "Overpriced to pay",
	2004: "Low-price for selling",
	2005: "Balance insufficient",
	2006: "Leverage ratio error",
	2007: "Order price error",
	2008: "The quantity is insufficient",
	2009: "Positions do not exist or have been closed",
	2011: "Order quantity error",
	2013: "Cancel orders over maximum limit",
	2014: "The quantity of batch order exceeds the limit",
	2015: "Price or quantity accuracy error",
	2016: "Trigger volume over the maximum",
	2018: "Exceeding the maximum available margin",
	2019: "There is an active open position",
	2021: "The single leverage is not consistent with the existing position leverage",
	2022: "Wrong position type",
	2023: "There are positions over the maximum leverage",
	2024: "There are orders with leverage over the maximum",
	2025: "The holding positions is over the maximum allowable positions",
	2026: "Modification of leverage is not supported for cross",
	2027: "There is only one cross or isolated in the same direction",
	2028: "The maximum order quantity is exceeded",
	2029: "Error order type",
	2030: "External order ID is too long (Max. 32 bits )",
	2031: "The allowable holding position exceed the current risk limit",
	2032: "Order price is less than long position force liquidate price",
	2033: "Order price is more than short position force liquidate price",
	2034: "The batch query quantity limit is exceeded",
	2035: "Unsupported market price tier",
	2036: "The orders more than the limit, please contact customer service",
	2037: "Frequent transactions, please try it later",
	2038: "The maximum allowable position quantity is exceeded, please contact customer service!",
	3001: "Trigger price type error",
	3002: "Trigger type error",
	3003: "Executive cycle error",
	3004: "Trigger price error",
	4001: "Unsupported currency",
	5001: "The take-price and the stop-loss price cannot be none at the same time",
	5002: "The Stop-Limit order does not exist or has closed",
	5003: "Take-profit and stop-loss price setting is wrong",
	5004: "The take-profit and stop-loss order volume is more than the holding positions can be liquidated",
	6001: "Trading forbidden",
	6002: "Open forbidden",
	6003: "Time range error",
	6004: "The trading pair and status should be fill in",
	6005: "The trading pair is not available",
	9999: "Public abnormal",
}

var commonErrors = map[ErrorCode]struct{}{
	ErrUnauthorized:         {},
	ErrInternalError:        {},
	ErrSystemBusy:           {},
	ErrUnknownRequestSource: {},
	ErrTooManyRequests:      {},
	ErrEndpointInaccessible: {},
	ErrPublicAbnormal:       {},
}

var authErrors = map[ErrorCode]struct{}{
	ErrAPIKeyExpired:                   {},
	ErrIPNotInWhitelist:                {},
	ErrInvalidRequestTime:              {},
	ErrVerifyFailed:                    {},
	ErrAccountReadPermissionRequired:   {},
	ErrAccountModifyPermissionRequired: {},
	ErrTradeReadPermissionRequired:     {},
	ErrTradeModifyPermissionRequired:   {},
}

var userErrors = map[ErrorCode]struct{}{
	ErrAccountDoesNotExist: {},
}

var marketErrors = map[ErrorCode]struct{}{
	ErrContractDoesNotExist: {},
	ErrContractNotActivated: {},
	ErrUnsupportedCurrency:  {},
	ErrTradingForbidden:     {},
	ErrOpenForbidden:        {},
	ErrPairNotAvailable:     {},
}

var balanceErrors = map[ErrorCode]struct{}{
	ErrAmountError:         {},
	ErrBalanceInsufficient: {},
}

var parameterErrors = map[ErrorCode]struct{}{
	ErrParameterError:      {},
	ErrDataDecodingError:   {},
	ErrRepeatedRequests:    {},
	ErrPairAndStatusNeeded: {},
}

var tradingErrors = map[ErrorCode]struct{}{
	ErrRiskLimitLevelError:             {},
	ErrWrongOrderDirection:             {},
	ErrWrongOpeningType:                {},
	ErrOverpricedToPay:                 {},
	ErrLowPriceForSelling:              {},
	ErrLeverageRatioError:              {},
	ErrOrderPriceError:                 {},
	ErrQuantityInsufficient:            {},
	ErrPositionsDoNotExist:             {},
	ErrOrderQuantityError:              {},
	ErrCancelOrdersOverMaximumLimit:    {},
	ErrBatchOrderQuantityExceedsLimit:  {},
	ErrPriceOrQuantityAccuracyError:    {},
	ErrTriggerVolumeOverMaximum:        {},
	ErrExceedingMaxAvailableMargin:     {},
	ErrActiveOpenPositionExists:        {},
	ErrLeverageNotConsistent:           {},
	ErrWrongPositionType:               {},
	ErrPositionsOverMaxLeverage:        {},
	ErrOrdersWithLeverageOverMax:       {},
	ErrHoldingPositionsOverMax:         {},
	ErrLeverageModificationUnsupported: {},
	ErrOnlyOneCrossOrIsolatedAllowed:   {},
	ErrMaxOrderQuantityExceeded:        {},
	ErrErrorOrderType:                  {},
	ErrExternalOrderIDTooLong:          {},
	ErrPositionExceedsRiskLimit:        {},
	ErrPriceLessThanLiqPrice:           {},
	ErrPriceMoreThanLiqPrice:           {},
	ErrBatchQueryQuantityExceeded:      {},
	ErrUnsupportedMarketPriceTier:      {},
	ErrOrdersMoreThanLimit:             {},
	ErrFrequentTransactions:            {},
	ErrMaxAllowablePositionExceeded:    {},
	ErrTriggerPriceTypeError:           {},
	ErrTriggerTypeError:                {},
	ErrExecutiveCycleError:             {},
	ErrTriggerPriceError:               {},
	ErrSLTPPriceCannotBeNone:           {},
	ErrSLTPOrderDoesNotExist:           {},
	ErrSLTPPriceSettingWrong:           {},
	ErrSLTPVolumeMoreThanPosition:      {},
	ErrTimeRangeError:                  {},
}

// IsCommonError returns true if the error is a common error.
func (e ErrorCode) IsCommonError() bool {
	_, ok := commonErrors[e]
	return ok
}

// IsAuthError returns true if the error is an authentication error.
func (e ErrorCode) IsAuthError() bool {
	_, ok := authErrors[e]
	return ok
}

// IsUserError returns true if the error is a user error.
func (e ErrorCode) IsUserError() bool {
	_, ok := userErrors[e]
	return ok
}

// IsMarketError returns true if the error is a market error.
func (e ErrorCode) IsMarketError() bool {
	_, ok := marketErrors[e]
	return ok
}

// IsBalanceError returns true if the error is a balance error.
func (e ErrorCode) IsBalanceError() bool {
	_, ok := balanceErrors[e]
	return ok
}

// IsParameterError returns true if the error is a parameter error.
func (e ErrorCode) IsParameterError() bool {
	_, ok := parameterErrors[e]
	return ok
}

// IsTradingError returns true if the error is a trading error.
func (e ErrorCode) IsTradingError() bool {
	_, ok := tradingErrors[e]
	return ok
}

func (e ErrorCode) Error() string {
	if msg, ok := errorCodes[int(e)]; ok {
		return msg
	}
	return fmt.Sprintf("unknown error code: %d", e)
}

