package errs

import "fmt"

// ErrorCode represents an error code.
type ErrorCode int

// --- Common & HTTP-level errors (11)
const (
	ErrUnknownOrderSent    = ErrorCode(-2011)
	ErrOperationNotAllowed = ErrorCode(26)
	ErrAPIKeyRequired      = ErrorCode(400)
	ErrNoAuthority         = ErrorCode(401)
	ErrAccessDenied        = ErrorCode(403)
	ErrTooManyRequests     = ErrorCode(429)
	ErrInternalError       = ErrorCode(500)
	ErrServiceUnavailable  = ErrorCode(503)
	ErrGatewayTimeout      = ErrorCode(504)
	ErrInternalSystemError = ErrorCode(20002)
	ErrRequestFailed       = ErrorCode(730000)
)

// --- Auth & Signature errors (9)
const (
	ErrSignatureVerificationFailed = ErrorCode(602)
	ErrInvalidAccessKey            = ErrorCode(10072)
	ErrInvalidRequestTime          = ErrorCode(10073)
	ErrAPIKeyFormatInvalid         = ErrorCode(700001)
	ErrSignatureInvalid            = ErrorCode(700002)
	ErrTimestampOutsideRecvWindow  = ErrorCode(700003)
	ErrRecvWindowTooLarge          = ErrorCode(700005)
	ErrNoPermissionToEndpoint      = ErrorCode(700007)
	ErrAccessKeyNull               = ErrorCode(730707)
)

// --- User & account errors (16)
const (
	ErrUserDoesNotExist            = ErrorCode(10001)
	ErrUserIDCannotBeNull          = ErrorCode(10015)
	ErrUserSubAccountNotOpen       = ErrorCode(10099)
	ErrRecordDoesNotExist          = ErrorCode(22222)
	ErrAccountAbnormal             = ErrorCode(60005)
	ErrPairUserBanTradeAPIKey      = ErrorCode(70011)
	ErrSubAccountNotExist          = ErrorCode(140001)
	ErrSubAccountForbidden         = ErrorCode(140002)
	ErrUserInfoError               = ErrorCode(730001)
	ErrUserStatusUnusual           = ErrorCode(730100)
	ErrUsernameAlreadyExists       = ErrorCode(730101)
	ErrSubAccountNameNull          = ErrorCode(730600)
	ErrSubAccountNameInvalidFormat = ErrorCode(730601)
	ErrSubAccountRemarksNull       = ErrorCode(730602)
	ErrAPIKeyRemarksNull           = ErrorCode(730700)
	ErrAPIKeyInfoNotExist          = ErrorCode(730706)
)

// --- Symbol, pair & market errors (18)
const (
	ErrBadSymbol                      = ErrorCode(10007)
	ErrCurrencyCannotBeNull           = ErrorCode(10222)
	ErrCurrencyDoesNotExist           = ErrorCode(10232)
	ErrTransactionSuspended           = ErrorCode(30000)
	ErrTransactionDirectionNotAllowed = ErrorCode(30001)
	ErrNoValidTradePrice              = ErrorCode(30010)
	ErrInvalidSymbol                  = ErrorCode(30014)
	ErrTradingDisabled                = ErrorCode(30016)
	ErrMarketOrderDisabled            = ErrorCode(30018)
	ErrAPIMarketOrderDisabled         = ErrorCode(30019)
	ErrNoPermissionForSymbol          = ErrorCode(30020)
	ErrInvalidSymbolAlt               = ErrorCode(30021)
	ErrOpponentOrderNotExist          = ErrorCode(30025)
	ErrInvalidOrderIDs                = ErrorCode(30026)
	ErrMaxPositionLimitReached        = ErrorCode(30027)
	ErrRiskControlSellingSuspended    = ErrorCode(30028)
	ErrMaxOrderLimitExceeded          = ErrorCode(30029)
	ErrOrderTypeNotAllowed            = ErrorCode(30041)
)

// --- Balance & amount related errors (10)
const (
	ErrAmountCannotBeNull           = ErrorCode(10095)
	ErrAmountDecimalPlacesTooLong   = ErrorCode(10096)
	ErrAmountIsError                = ErrorCode(10097)
	ErrInsufficientBalance          = ErrorCode(10101)
	ErrAmountZeroOrNegative         = ErrorCode(10102)
	ErrMinTransactionVolumeTooSmall = ErrorCode(30002)
	ErrMaxTransactionVolumeTooLarge = ErrorCode(30003)
	ErrInsufficientPosition         = ErrorCode(30004)
	ErrOversold                     = ErrorCode(30005)
	ErrMaxPositionExceeded          = ErrorCode(30032)
)

// --- Transfer & withdrawal errors (13)
const (
	ErrRiskControlAbnormal              = ErrorCode(10098)
	ErrCurrencyTransferNotSupported     = ErrorCode(10100)
	ErrAccountTransferNotSupported      = ErrorCode(10103)
	ErrTransferOperationProcessing      = ErrorCode(10200)
	ErrTransferInFailed                 = ErrorCode(10201)
	ErrTransferOutFailed                = ErrorCode(10202)
	ErrTransferDisabled                 = ErrorCode(10206)
	ErrTransferForbidden                = ErrorCode(10211)
	ErrWithdrawalAddressInvalid         = ErrorCode(10212)
	ErrNoAddressAvailable               = ErrorCode(10216)
	ErrAssetFlowWritingFailed           = ErrorCode(10219)
	ErrIntermediateAccountNotInRedis    = ErrorCode(10259)
	ErrWithdrawalUnavailableRiskControl = ErrorCode(10265)
)

// --- Parameter & format errors (14)
const (
	ErrRemarkTooLong                = ErrorCode(10268)
	ErrSubsystemNotSupported        = ErrorCode(20001)
	ErrParamIsError                 = ErrorCode(33333)
	ErrParamCannotBeNull            = ErrorCode(44444)
	ErrOrderIDOrClientIDRequired    = ErrorCode(700004)
	ErrIPNotWhitelisted             = ErrorCode(700006)
	ErrIllegalCharactersInParameter = ErrorCode(700008)
	ErrInvalidContentType           = ErrorCode(700013) // undocumented
	ErrParameterError               = ErrorCode(730002)
	ErrUnsupportedOperation         = ErrorCode(730003)
	ErrAPIKeyPermissionNull         = ErrorCode(730701)
	ErrAPIKeyPermissionNotExist     = ErrorCode(730702)
	ErrIPInfoIncorrect              = ErrorCode(730703)
	ErrIPFormatIncorrect            = ErrorCode(730704)
	ErrAPIKeyGroupLimitReached      = ErrorCode(730705)
)

var errorCodes = map[int]string{
	-2011:  "Unknown order sent",
	26:     "operation not allowed",
	400:    "api key required",
	401:    "No authority",
	403:    "Access Denied",
	429:    "Too Many Requests",
	500:    "Internal error",
	503:    "service not available, please try again",
	504:    "Gateway Time-out",
	602:    "Signature verification failed",
	10001:  "user does not exist",
	10007:  "bad symbol",
	10015:  "user id cannot be null",
	10072:  "invalid access key",
	10073:  "invalid Request-Time",
	10095:  "amount cannot be null",
	10096:  "amount decimal places is too long",
	10097:  "amount is error",
	10098:  "risk control system detected abnormal",
	10099:  "user sub account does not open",
	10100:  "this currency transfer is not supported",
	10101:  "Insufficient balance",
	10102:  "amount cannot be zero or negative",
	10103:  "this account transfer is not supported",
	10200:  "transfer operation processing",
	10201:  "transfer in failed",
	10202:  "transfer out failed",
	10206:  "transfer is disabled",
	10211:  "transfer is forbidden",
	10212:  "This withdrawal address is not on the commonly used address list or has been invalidated",
	10216:  "no address available. Please try again later",
	10219:  "asset flow writing failed please try again",
	10222:  "currency cannot be null",
	10232:  "currency does not exist",
	10259:  "Intermediate account does not configured in redisredis",
	10265:  "Due to risk control, withdrawal is unavailable, please try again later",
	10268:  "remark length is too long",
	20001:  "subsystem is not supported",
	20002:  "Internal system error please contact support",
	22222:  "record does not exist",
	30000:  "suspended transaction for the symbol",
	30001:  "The current transaction direction is not allowed to place an order",
	30002:  "The minimum transaction volume cannot be less than :",
	30003:  "The maximum transaction volume cannot be greater than :",
	30004:  "Insufficient position",
	30005:  "Oversold",
	30010:  "no valid trade price",
	30014:  "invalid symbol",
	30016:  "trading disabled",
	30018:  "market order is disabled",
	30019:  "api market order is disabled",
	30020:  "no permission for the symbol",
	30021:  "invalid symbol",
	30025:  "no exist opponent order",
	30026:  "invalid order ids",
	30027:  "The currency has reached the maximum position limit, the buying is suspended",
	30028:  "The currency triggered the platform risk control, the selling is suspended",
	30029:  "Cannot exceed the maximum order limit",
	30032:  "Cannot exceed the maximum position",
	30041:  "current order type can not place order",
	33333:  "param is error",
	44444:  "param cannot be null",
	60005:  "your account is abnormal",
	70011:  "Pair user ban trade apikey",
	140001: "sub account does not exist",
	140002: "sub account is forbidden",
	700001: "API-key format invalid",
	700002: "Signature for this request is not valid",
	700003: "Timestamp for this request is outside of the recvWindow",
	700004: "Param 'origClientOrderId' or 'orderId' must be sent, but both were empty/null",
	700005: "recvWindow must less than 60000",
	700006: "IP non white list",
	700007: "No permission to access the endpoint",
	700008: "Illegal characters found in parameter",
	700013: "Invalid content Type.",
	730000: "Request failed, please contact the customer service",
	730001: "User information error",
	730002: "Parameter error",
	730003: "Unsupported operation, please contact the customer service",
	730100: "Unusual user status",
	730101: "The user Name already exists",
	730600: "Sub-account Name cannot be null",
	730601: "Sub-account Name must be a combination of 8-32 letters and numbers",
	730602: "Sub-account remarks cannot be null",
	730700: "API KEY remarks cannot be null",
	730701: "API KEY permission cannot be null",
	730702: "API KEY permission does not exist",
	730703: "The IP information is incorrect, and a maximum of 10 IPs are allowed to be bound only",
	730704: "The bound IP format is incorrect, please refill",
	730705: "At most 30 groups of Api Keys are allowed to be created only",
	730706: "API KEY information does not exist",
	730707: "accessKey cannot be null",
}

var commonErrors = map[ErrorCode]struct{}{
	ErrUnknownOrderSent:    {},
	ErrOperationNotAllowed: {},
	ErrAPIKeyRequired:      {},
	ErrNoAuthority:         {},
	ErrAccessDenied:        {},
	ErrTooManyRequests:     {},
	ErrInternalError:       {},
	ErrServiceUnavailable:  {},
	ErrGatewayTimeout:      {},
	ErrInternalSystemError: {},
	ErrRequestFailed:       {},
}

var authErrors = map[ErrorCode]struct{}{
	ErrSignatureVerificationFailed: {},
	ErrInvalidAccessKey:            {},
	ErrInvalidRequestTime:          {},
	ErrAPIKeyFormatInvalid:         {},
	ErrSignatureInvalid:            {},
	ErrTimestampOutsideRecvWindow:  {},
	ErrRecvWindowTooLarge:          {},
	ErrNoPermissionToEndpoint:      {},
	ErrAccessKeyNull:               {},
}

var userErrors = map[ErrorCode]struct{}{
	ErrUserDoesNotExist:            {},
	ErrUserIDCannotBeNull:          {},
	ErrUserSubAccountNotOpen:       {},
	ErrRecordDoesNotExist:          {},
	ErrAccountAbnormal:             {},
	ErrPairUserBanTradeAPIKey:      {},
	ErrSubAccountNotExist:          {},
	ErrSubAccountForbidden:         {},
	ErrUserInfoError:               {},
	ErrUserStatusUnusual:           {},
	ErrUsernameAlreadyExists:       {},
	ErrSubAccountNameNull:          {},
	ErrSubAccountNameInvalidFormat: {},
	ErrSubAccountRemarksNull:       {},
	ErrAPIKeyRemarksNull:           {},
	ErrAPIKeyInfoNotExist:          {},
}

var marketErrors = map[ErrorCode]struct{}{
	ErrBadSymbol:                      {},
	ErrCurrencyCannotBeNull:           {},
	ErrCurrencyDoesNotExist:           {},
	ErrTransactionSuspended:           {},
	ErrTransactionDirectionNotAllowed: {},
	ErrNoValidTradePrice:              {},
	ErrInvalidSymbol:                  {},
	ErrTradingDisabled:                {},
	ErrMarketOrderDisabled:            {},
	ErrAPIMarketOrderDisabled:         {},
	ErrNoPermissionForSymbol:          {},
	ErrInvalidSymbolAlt:               {},
	ErrOpponentOrderNotExist:          {},
	ErrInvalidOrderIDs:                {},
	ErrMaxPositionLimitReached:        {},
	ErrRiskControlSellingSuspended:    {},
	ErrMaxOrderLimitExceeded:          {},
	ErrOrderTypeNotAllowed:            {},
}

var balanceErrors = map[ErrorCode]struct{}{
	ErrAmountCannotBeNull:           {},
	ErrAmountDecimalPlacesTooLong:   {},
	ErrAmountIsError:                {},
	ErrInsufficientBalance:          {},
	ErrAmountZeroOrNegative:         {},
	ErrMinTransactionVolumeTooSmall: {},
	ErrMaxTransactionVolumeTooLarge: {},
	ErrInsufficientPosition:         {},
	ErrOversold:                     {},
	ErrMaxPositionExceeded:          {},
}

var transferErrors = map[ErrorCode]struct{}{
	ErrRiskControlAbnormal:              {},
	ErrCurrencyTransferNotSupported:     {},
	ErrAccountTransferNotSupported:      {},
	ErrTransferOperationProcessing:      {},
	ErrTransferInFailed:                 {},
	ErrTransferOutFailed:                {},
	ErrTransferDisabled:                 {},
	ErrTransferForbidden:                {},
	ErrWithdrawalAddressInvalid:         {},
	ErrNoAddressAvailable:               {},
	ErrAssetFlowWritingFailed:           {},
	ErrIntermediateAccountNotInRedis:    {},
	ErrWithdrawalUnavailableRiskControl: {},
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

// IsTransferError returns true if the error is a transfer error.
func (e ErrorCode) IsTransferError() bool {
	_, ok := transferErrors[e]
	return ok
}

func (e ErrorCode) Error() string {
	if msg, ok := errorCodes[int(e)]; ok {
		return msg
	}
	return fmt.Sprintf("unknown error code: %d", e)
}
