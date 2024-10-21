package utils

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog/log"

	"loadtester/types"
)

func LoadAccsFromFile() ([]*types.Account, error) {
	usr, err := user.Current()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get current user")
	}
	filePath := filepath.Join(usr.HomeDir, "test_accounts", "pks.json")
	log.Debug().Msgf("loading accounts from %s", filePath)
	bz, err := os.ReadFile(filePath)
	if err != nil {
		log.Err(err).Msg("failed to read account file")
		return nil, err
	}

	// load private keys
	var pks []string
	err = json.Unmarshal(bz, &pks)
	if err != nil {
		log.Err(err).Msg("failed to unmarshal private keys")
		return nil, err
	}

	var accs []*types.Account
	pkMap := make(map[string]bool)
	for _, pk := range pks {
		if pkMap[pk] {
			panic("duplicate private key")
		}
		pkMap[pk] = true
		// remove 0x prefix
		if len(pk) > 2 && pk[:2] == "0x" {
			pk = pk[2:]
		}
		privKey, err := crypto.HexToECDSA(pk)
		if err != nil {
			log.Err(err).Msg("failed to parse private key")
			return nil, err
		}
		accs = append(accs, types.NewAccount(privKey))
	}
	log.Debug().Msgf("done loading %d accounts", len(accs))
	return accs, nil
}

func SelectAccInfosForNode(nodeNum, nodeIdx int, accs []types.AccountInfo) []types.AccountInfo {
	if nodeNum == 1 {
		return accs
	}
	startIdx := nodeIdx * len(accs) / nodeNum
	log.Debug().Msgf("node %d will use accounts from %d to %d", nodeIdx, startIdx, (nodeIdx+1)*len(accs)/nodeNum)
	endIdx := (nodeIdx + 1) * len(accs) / nodeNum
	return accs[startIdx:endIdx]
}

func SelectAccountsToUse(tps int, accs []*types.Account, startIdx int, accType string) []*types.Account {
	endIndex := startIdx + tps
	if endIndex > len(accs) {
		// we need to wrap around
		endIndex = tps - (len(accs) - startIdx)
		log.Debug().Msgf("selecting %d %s to use, accs[%d:], accs[:%d]", tps, accType, startIdx, endIndex)
		return append(accs[startIdx:], accs[:endIndex]...)
	}
	log.Debug().Msgf("selecting %d %s to use accs[%d:%d]", tps, accType, startIdx, endIndex)
	return accs[startIdx:endIndex]
}

// CreateRandomAccounts creates random accounts which only holds address, not private key.
func CreateRandomAccounts(num int) []*types.Account {
	accs := make([]*types.Account, num)
	for i := 0; i < num; i++ {
		accs[i] = &types.Account{EthAddr: *CreateRandomAddress()}
	}
	return accs
}
