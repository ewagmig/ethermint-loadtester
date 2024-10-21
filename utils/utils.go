package utils

import (
	"encoding/json"
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog/log"
	"loadtester/types"
)

func CreateRandomAddress() *common.Address {
	key, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(key.PublicKey)
	return &addr
}

func CreateRandomAcc() *types.Account {
	key, _ := crypto.GenerateKey()
	return types.NewAccount(key)
}

func WritePrivateKeysToFile(accs []*types.Account) {
	usr, err := user.Current()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get current user")
	}
	filePath := filepath.Join(usr.HomeDir, "test_accounts", "pks.json")
	var pks []string
	for _, acc := range accs {
		pks = append(pks, hexutil.Encode(crypto.FromECDSA(acc.GetEthPrivKey())))
	}
	bz, err := json.Marshal(pks)
	if err != nil {
		log.Err(err).Msg("failed to marshal private keys")
		return
	}
	log.Debug().Msgf("writing private keys to %s", filePath)
	err = os.WriteFile(filePath, bz, 0644)
	if err != nil {
		log.Err(err).Msg("failed to write account file")
		return
	}
	log.Debug().Msg("done writing private keys")
}

func AccSanityCheck(accounts []*types.Account, accMap map[string]bool) error {
	for _, acc := range accounts {
		// sanity check
		if _, ok := accMap[acc.GetEthAddr().Hex()]; ok {
			log.Info().Msg("duplicated account, stop testing")
			return errors.New("duplicated account")
		}
		accMap[acc.GetEthAddr().Hex()] = true
	}
	return nil
}

func TxSanityCheck(txHashes []string, txHashMap map[string]bool) error {
	for _, txHash := range txHashes {
		// sanity check
		if _, ok := txHashMap[txHash]; ok {
			log.Info().Msg("duplicated txHash, stop testing")
			return errors.New("duplicated txHash")
		}
		txHashMap[txHash] = true
	}
	return nil
}

func MustPareDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestEnded(end time.Time) bool {
	if time.Now().After(end) {
		log.Info().Msg("test ended")
		return true
	}
	return false
}
