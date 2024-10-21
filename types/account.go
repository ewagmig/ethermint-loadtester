package types

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Account struct {
	EthPrivKey *ecdsa.PrivateKey `json:"EthPrivKey"`
	EthAddr    common.Address    `json:"EthAddr"`
	Nonce      uint64            `json:"Nonce"`
}

func NewAccount(ethPrivKey *ecdsa.PrivateKey) (account *Account) {
	account = &Account{
		EthPrivKey: ethPrivKey,
		EthAddr:    crypto.PubkeyToAddress(ethPrivKey.PublicKey),
	}
	return
}

func (a *Account) SetNonce(nonce uint64) {
	a.Nonce = nonce
}

func (a *Account) IncreaseNonce() {
	a.Nonce++
}

func (a Account) GetEthPrivKey() *ecdsa.PrivateKey {
	return a.EthPrivKey
}

func (a Account) GetEthAddr() *common.Address {
	return &a.EthAddr
}

func (a Account) GetNonce() uint64 {
	return a.Nonce
}

// create copy method
func (a Account) Copy() Account {
	return Account{
		EthPrivKey: a.EthPrivKey,
		EthAddr:    a.EthAddr,
		Nonce:      a.Nonce,
	}
}
