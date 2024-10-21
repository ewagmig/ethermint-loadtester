package offchain_feeding

import (
	"encoding/json"
	"math/big"
	"os"
	"strconv"
	"sync"

	"github.com/btcsuite/btcutil/bech32"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"loadtester/types"
	"loadtester/utils"
)

const GENESIS = "genesis.json"
const ONE_ETH = 1_000_000_000_000_000_000
const EMPTY_CODEHASH = "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"

// TODO: Refactor this function to use for other chains, it is only for Ethermint based chain currently.
func NewOffchainFeedingCmd(cfg Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "offchain_feeding",
		Short: "Create multiple 1 eth acconts on the genesis",
		RunE: func(cmd *cobra.Command, args []string) error {
			genesisBz := readGenesis(cfg)
			accs, bals := loadAccountsAndBalances(genesisBz)
			totalSupply := loadTotalSupply(genesisBz)
			pkAccs, newAccs, newBals := make([]*types.Account, cfg.AccNum), make([]EthAccount, cfg.AccNum), make([]Balance, cfg.AccNum)

			// add new accounts
			wg := sync.WaitGroup{}
			currentLatestAccNum, _ := strconv.Atoi(accs[len(accs)-1].BaseAccount.AccountNumber)
			accNum := currentLatestAccNum + 1
			log.Info().Int("currentLatestAccNum", currentLatestAccNum).Int("cfg.AccNum", cfg.AccNum).Msg("creating new accounts")
			for i := 0; i < cfg.AccNum; i++ {
				wg.Add(1)
				go func(idx, accountNumber int) {
					defer wg.Done()
					acc := utils.CreateRandomAcc()
					pkAccs[idx] = acc
					ethAcc, bal := createAccountAndBalance(cfg, acc, accountNumber)
					newAccs[idx], newBals[idx] = ethAcc, bal
				}(i, accNum)
				accNum++
			}
			wg.Wait()
			log.Info().Msg("done creating new accounts")

			// write private keys
			utils.WritePrivateKeysToFile(pkAccs)

			// append newAccs and newBals to accs and bals
			accs = append(accs, newAccs...)
			bals = append(bals, newBals...)

			// write accounts and balances
			genesisBz["app_state"].(map[string]interface{})["auth"].(map[string]interface{})["accounts"] = accs
			genesisBz["app_state"].(map[string]interface{})["bank"].(map[string]interface{})["balances"] = bals

			originalTotalSupplyAmt, ok := new(big.Int).SetString(totalSupply[0].Amount, 10)
			if !ok {
				log.Fatal().Msg("failed to set total supply amount")
			}
			totalSupply[0].Amount = new(big.Int).Add(originalTotalSupplyAmt, new(big.Int).Mul(big.NewInt(int64(cfg.AccNum)), big.NewInt(ONE_ETH))).String()
			genesisBz["app_state"].(map[string]interface{})["bank"].(map[string]interface{})["supply"] = totalSupply

			writeGenesis(cfg, genesisBz)
			return nil
		},
	}
	return cmd
}

func readGenesis(cfg Config) map[string]interface{} {
	bz, err := os.ReadFile(cfg.GenesisLoc)
	if err != nil {
		panic(err)
	}

	// parse genesis.json
	var genesis map[string]interface{}
	err = json.Unmarshal(bz, &genesis)
	if err != nil {
		panic(err)
	}
	return genesis
}

func writeGenesis(cfg Config, genesisBz map[string]interface{}) {
	log.Info().Msgf("writing %s", GENESIS)
	bz, err := json.Marshal(genesisBz)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to marshal genesisBz")
	}
	err = os.WriteFile(cfg.GenesisLoc, bz, 0644)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to write %s", GENESIS)
	}
	log.Info().Msgf("done writing %s", GENESIS)
}

func loadAccountsAndBalances(genesisBz map[string]interface{}) ([]EthAccount, []Balance) {
	log.Info().Msg("loading accounts and balances from genesis")
	var accounts []EthAccount
	accountsBz, err := json.Marshal(genesisBz["app_state"].(map[string]interface{})["auth"].(map[string]interface{})["accounts"])
	if err != nil {
		log.Fatal().Err(err).Msg("failed to marshal accounts")
	}
	err = json.Unmarshal(accountsBz, &accounts)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to unmarshal accounts")
	}
	var balances []Balance
	balancesBz, err := json.Marshal(genesisBz["app_state"].(map[string]interface{})["bank"].(map[string]interface{})["balances"])
	if err != nil {
		log.Fatal().Err(err).Msg("failed to marshal balances")
	}
	err = json.Unmarshal(balancesBz, &balances)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to unmarshal balances")
	}
	log.Info().Msg("done loading accounts and balances from genesis")
	return accounts, balances
}

func loadTotalSupply(genesisBz map[string]interface{}) []Coin {
	log.Info().Msg("loading total supply from genesis")
	totalSupplyBz, err := json.Marshal(genesisBz["app_state"].(map[string]interface{})["bank"].(map[string]interface{})["supply"])
	if err != nil {
		log.Fatal().Err(err).Msg("failed to marshal total supply")
	}
	var totalSupply []Coin
	err = json.Unmarshal(totalSupplyBz, &totalSupply)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to unmarshal total supply")
	}
	log.Info().Msg("done loading total supply from genesis")
	return totalSupply
}

// TODO: Refactor this function to use for other chains, it is only for Ethermint based chain currently.
func createAccountAndBalance(conf Config, acc *types.Account, accountNumber int) (EthAccount, Balance) {
	converted, err := bech32.ConvertBits(acc.EthAddr.Bytes(), 8, 5, true)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to convert bits")
	}
	bech32Address, err := bech32.Encode(conf.BechPrefix, converted)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to encode")
	}

	ethAcc := EthAccount{
		Type: "/ethermint.types.v1.EthAccount",
		BaseAccount: BaseAccount{
			Address:       bech32Address,
			PubKey:        nil,
			AccountNumber: strconv.Itoa(accountNumber),
			Sequence:      "0",
		},
		CodeHash: EMPTY_CODEHASH, // eoa has no code
	}
	balance := Balance{
		Address: bech32Address,
		Coins: []Coin{
			{
				Denom:  conf.Denom,
				Amount: strconv.Itoa(ONE_ETH),
			},
		},
	}
	return ethAcc, balance
}
