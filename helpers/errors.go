package helpers

import "github.com/pkg/errors"

// General errors.
var (
	ErrInvalidTxID       = errors.New("invalid TxID")
	ErrTxNil             = errors.New("tx is nil")
	ErrTxTooShort        = errors.New("too short to be a tx - even an empty tx has 10 bytes")
	ErrNLockTimeLength   = errors.New("nLockTime length must be 4 bytes long")
	ErrEmptyValues       = errors.New("empty value or values passed, all arguments are required and cannot be empty")
	ErrUnsupportedScript = errors.New("non-P2PKH input used in the tx - unsupported")
	ErrInvalidScriptType = errors.New("invalid script type")
	ErrNoUnlocker        = errors.New("unlocker not supplied")
)

// Sentinal errors reported by inputs.
var (
	ErrSendMessage = errors.New("Failed to send message:")
)
