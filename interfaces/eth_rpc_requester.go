package interfaces

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

type EthRpcRequester interface {
	EthSendRawTransaction(rawTx []byte) error
	EthSendRawTransactionNoWaiting(rawTx []byte) error
	EthPendingNonce(addr common.Address) (uint64, error)
	EthSendMultipleRawTransactions(rawTxs [][]byte, cb func(*sync.Mutex, int)) (failed int64)
}
