package transaction

type (
	// TxnStatus represented zcncore.TransactionCallback operations statuses.
	TxnStatus int
)

const (
	// StatusSuccess represent zcncore.StatusSuccess.
	StatusSuccess TxnStatus = iota
	// StatusNetworkError represent zcncore.StatusNetworkError.
	StatusNetworkError
	// StatusError represent zcncore.StatusError.
	StatusError
	// StatusRejectedByUser represent zcncore.StatusRejectedByUser.
	StatusRejectedByUser
	// StatusInvalidSignature represent zcncore.StatusInvalidSignature.
	StatusInvalidSignature
	// StatusAuthError represent zcncore.StatusAuthError.
	StatusAuthError
	// StatusAuthVerifyFailed represent zcncore.StatusAuthVerifyFailed.
	StatusAuthVerifyFailed
	// StatusAuthTimeout represent zcncore.StatusAuthTimeout.
	StatusAuthTimeout
	// StatusUnknown represent zcncore.StatusUnknown.
	StatusUnknown = -1
)

// String returns represented in string format TxnStatus.
func (ts TxnStatus) String() string {
	switch ts {
	case StatusSuccess:
		return "success"

	case StatusNetworkError:
		return "network error"

	case StatusError:
		return "error"

	case StatusRejectedByUser:
		return "rejected byt user"

	case StatusInvalidSignature:
		return "invalid signature"

	case StatusAuthError:
		return "auth error"

	case StatusAuthVerifyFailed:
		return "auth verify error"

	case StatusAuthTimeout:
		return "auth timeout error"

	case StatusUnknown:
		return "unknown"

	default:
		return ""
	}
}
