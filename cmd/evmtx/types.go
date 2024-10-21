package evmtx

import (
	"loadtester/interfaces"
	"loadtester/types"
)

// TransactionContext holds the configuration and RPC interfaces needed for transactions.
type TransactionContext struct {
	Config    *Config
	EthRpc    interfaces.EthRpcRequester
	Senders   []*types.Account
	Receivers []*types.Account
}
