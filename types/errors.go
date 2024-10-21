package types

import (
	"errors"
)

type NonceError struct {
	Message string
	Nonce   uint64
}

func (e *NonceError) Error() string {
	return e.Message
}

var (
	ErrorInsufficientFund   = errors.New("insufficient fund")
	ErrorFailedToFetchNonce = errors.New("failed to fetch nonce")
)
