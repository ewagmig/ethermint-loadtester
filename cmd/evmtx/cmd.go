package evmtx

import (
	"encoding/json"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"loadtester/interfaces"
	"loadtester/types"
	"loadtester/utils"
)

func NewEvmTxCmd(cfg Config, ethRpc interfaces.EthRpcRequester) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evmtx",
		Short: "Send multiple evm tx through JSON-RPC",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info().Msgf(`
start sending multiple evm tx thorugh JSON-RPC 
scenario: %s
transaction_per_timeunit: %d
timeunit: %s
duration: %s`, cfg.Scenario, cfg.TransactionPerTimeUnit, cfg.TimeUnit, cfg.Duration)
			var testAccs []*types.Account
			var err error
			// load accs from file
			// nonce of those accounts must be zero
			testAccs, err = utils.LoadAccsFromFile()
			if err != nil {
				return err
			}

			// testAccs with 1 eth are ready
			log.Info().Msgf("3 seconds rest before starting load testing")
			time.Sleep(3 * time.Second)

			log.Info().Msgf("start load testing: scenario=%s, unit=%s, tpu=%d, duration=%s", cfg.Scenario, cfg.TimeUnit, cfg.TransactionPerTimeUnit, cfg.Duration)
			RunScenario(&cfg, ethRpc, testAccs)
			return nil
		},
	}
	return cmd
}

// RunScenario handles the transaction execution for a given scenario configuration.
func RunScenario(cfg *Config, ethRpc interfaces.EthRpcRequester, testAccs []*types.Account) {
	senders, receivers, err := PrepareAccountsForScenario(cfg, testAccs)
	i := 0
	start := time.Now()
	end := start.Add(utils.MustPareDuration(cfg.Duration))
	timeSpentTotal := time.Duration(0)
	txHashMap := make(map[string]bool)
	accMap := make(map[string]bool)

	if err != nil {
		panic(err)
	}
	for {
		startIdx := (i * cfg.TransactionPerTimeUnit) % len(senders)
		sendersTouse := utils.SelectAccountsToUse(cfg.TransactionPerTimeUnit, senders, startIdx, "senders")
		if err := utils.AccSanityCheck(sendersTouse, accMap); err != nil {
			break
		}
		receiversToUse := utils.SelectAccountsToUse(cfg.TransactionPerTimeUnit, receivers, startIdx, "receivers")

		sentEthTxHashes, _, timeSpent := ExecuteEthTransactions(&TransactionContext{
			Config:    cfg,
			EthRpc:    ethRpc,
			Senders:   sendersTouse,
			Receivers: receiversToUse,
		})
		if err := utils.TxSanityCheck(sentEthTxHashes, txHashMap); err != nil {
			break
		}
		UpdateMetrics(&timeSpentTotal, timeSpent)
		if utils.TestEnded(end) {
			break
		}
		i++
	}
	LogResults(
		utils.MustPareDuration(cfg.TimeUnit), timeSpentTotal,
		cfg.TransactionPerTimeUnit, len(txHashMap))
}

// Prepares senders and receivers based on the test scenario.
func PrepareAccountsForScenario(cfg *Config, testAccs []*types.Account) (senders, receivers []*types.Account, err error) {
	switch cfg.Scenario {
	case ScenarioEthTransferToKnown:
		half := len(testAccs) / 2
		return testAccs[:half], testAccs[half:], nil
	case ScenarioEthTransferToSelf:
		return testAccs, testAccs, nil
	case ScenarioEthTransferToRandom:
		receivers = utils.CreateRandomAccounts(len(testAccs))
		return testAccs, receivers, nil
	default:
		return nil, nil, errors.New("invalid scenario")
	}
}

// ExecuteEthTransactions executes the transactions for the given context.
func ExecuteEthTransactions(ctx *TransactionContext) ([]string, int64, time.Duration) {
	signingStart := time.Now()
	log.Debug().Msgf("signing %d transactions", len(ctx.Senders))
	wg := sync.WaitGroup{}
	reqBodies, txHashes := CreateEthSendRawTransactionReqBodies(ctx, &wg)
	log.Debug().Msgf("done signing %d. took %s", len(reqBodies), time.Since(signingStart).String())

	sendingStart := time.Now()
	log.Debug().Msgf("sending %d transactions", len(reqBodies))

	var sentEthTxHashes []string
	failed := ctx.EthRpc.EthSendMultipleRawTransactions(reqBodies, func(mu *sync.Mutex, idx int) {
		ctx.Senders[idx].IncreaseNonce() // off-chain nonce increment for faster processing
		mu.Lock()
		sentEthTxHashes = append(sentEthTxHashes, txHashes[idx])
		mu.Unlock()
	})

	timeSpentForSending := time.Since(sendingStart)
	succeeded := int64(len(reqBodies)) - failed
	log.Debug().Msgf("done sending. succeeded: %d, failed: %d, took %s", succeeded, failed, timeSpentForSending.String())

	timeUnit := utils.MustPareDuration(ctx.Config.TimeUnit)
	if timeSpentForSending < timeUnit {
		remaining := timeUnit - timeSpentForSending
		log.Debug().Msgf("sleeping for %v to keep transaction per %v", remaining, timeUnit)
		time.Sleep(remaining)
		timeSpentForSending = timeUnit
	}

	return sentEthTxHashes, failed, timeSpentForSending
}

// CreateEthSendRawTransactionReqBodies creates eth_sendRawTransaction request bodies with go routines
func CreateEthSendRawTransactionReqBodies(
	ctx *TransactionContext, wg *sync.WaitGroup,
) (reqBodies [][]byte, txHashes []string) {
	gasLimit := uint64(ctx.Config.GasLimit)
	gasPrice := big.NewInt(ctx.Config.GasPrice)
	val := new(big.Int).SetInt64(ctx.Config.SendingAmt)

	reqBodies = make([][]byte, len(ctx.Senders))
	txHashes = make([]string, len(ctx.Senders))

	for i := 0; i < len(ctx.Senders); i++ {
		wg.Add(1)
		go func(w *sync.WaitGroup, idx int) {
			defer w.Done()
			// prepare legacy tx
			unsignedTx := gethtypes.NewTx(&gethtypes.LegacyTx{
				To:       ctx.Receivers[idx].GetEthAddr(),
				Nonce:    ctx.Senders[idx].GetNonce(),
				Value:    val,
				Gas:      gasLimit,
				GasPrice: gasPrice,
			})
			signer := gethtypes.NewEIP155Signer(big.NewInt(ctx.Config.ChainID))
			signedTx, _ := gethtypes.SignTx(unsignedTx, signer, ctx.Senders[idx].GetEthPrivKey())
			marshaled, _ := signedTx.MarshalBinary()
			reqBody, err := json.Marshal(map[string]interface{}{
				"jsonrpc": "2.0",
				"method":  "eth_sendRawTransaction",
				"params":  []string{hexutil.Encode(marshaled)},
				"id":      1,
			})
			if err != nil {
				log.Err(err).Msg("Failed to marshal request body")
				return
			}
			reqBodies[idx] = reqBody
			txHashes[idx] = signedTx.Hash().Hex()
		}(wg, i)
	}
	wg.Wait()
	return
}

func LogResults(timeUnit, timeSpentTotal time.Duration, targetTpu, succeeded int) {
	totalSent := float64(succeeded)
	var tpu float64
	switch timeUnit {
	case time.Millisecond:
		tpu = totalSent / float64(timeSpentTotal.Milliseconds())
	case time.Second:
		tpu = totalSent / timeSpentTotal.Seconds()
	}
	log.Info().Msgf(
		"evmtx load testing finished, numTotalSent:%v, timeSpent:%v, timeUnit:%s, targetTpu:%d, realTpu:%.2f",
		totalSent, timeSpentTotal, timeUnit, targetTpu, tpu)
}

func UpdateMetrics(timeSpentTotal *time.Duration, timeSpent time.Duration) {
	*timeSpentTotal += timeSpent
}
