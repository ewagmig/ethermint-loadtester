package utils

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"loadtester/cmd"
	types2 "loadtester/types"
)

func TestSelectAccountsToUse(t *testing.T) {
	accounts := createRandomAccounts(1000)

	tcs := []struct {
		startIdx int
		tps      int
	}{
		{0, 100},
		{990, 10},
		{700, 500},
	}

	// Test the function
	for _, tc := range tcs {
		expectedAccountsToUse := append(accounts[tc.startIdx:tc.startIdx+tc.tps], accounts[:tc.tps]...)
		accountsToUse := SelectAccountsToUse(tc.tps, accounts, tc.startIdx, "senders")
		require.Equal(t, tc.tps, len(accountsToUse), "Expected %d accounts to use, got %d", tc.tps, len(accountsToUse))
		for i := 0; i < tc.tps; i++ {
			require.Equal(t, expectedAccountsToUse[i], accountsToUse[i], "Expected account %v, got %v", expectedAccountsToUse[i], accountsToUse[i])
		}
	}
}

func TestSelectAccInfosForNode(t *testing.T) {
	cfg := cmd.DefaultConfig()
	accInfos := createDummyAccInfos(cfg.EvmTxConfig.AccNum)
	require.Len(t, accInfos, 100)

	tcs := []struct {
		nodeNum  int
		nodeIdx  int
		expected []types2.AccountInfo
	}{
		{1, 0, accInfos},
		{2, 0, accInfos[:50]},
		{3, 1, accInfos[33:66]},
	}

	// Test the function
	for _, tc := range tcs {
		selected := SelectAccInfosForNode(tc.nodeNum, tc.nodeIdx, accInfos)
		require.Equal(t, tc.expected, selected, "Expected %v, got %v", tc.expected, selected)
	}
}

func createRandomAccounts(n int) []*types2.Account {
	accounts := make([]*types2.Account, n)
	for i := 0; i < n; i++ {
		key, _ := crypto.GenerateKey()
		accounts = append(accounts, types2.NewAccount(key))
	}
	return accounts
}

func createDummyAccInfos(n int) []types2.AccountInfo {
	var accInfos []types2.AccountInfo
	for i := 0; i < n; i++ {
		accInfos = append(accInfos, types2.AccountInfo{
			Name:     "",
			Address:  fmt.Sprintf("%d", i), // dummy address for identification
			Mnemonic: "",
		})
	}
	return accInfos
}
