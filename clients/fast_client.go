package clients

import (
	"encoding/json"
	"regexp"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"loadtester/types"
)

// FastClient sends Ethereum json-rpc using fasthttp.
type FastClient struct {
	cli                *fasthttp.Client
	wg                 *sync.WaitGroup
	jsonRPCAddr        string
	insufficientFundRe *regexp.Regexp
	invalidNonceRe     *regexp.Regexp
}

// NewFastClient creates a new FastClient.
func NewFastClient(jsonRPCAddr string) *FastClient {
	// TODO: configure timeouts
	readTimeout, _ := time.ParseDuration("3m")
	writeTimeout, _ := time.ParseDuration("3m")
	maxIdleConnDuration, _ := time.ParseDuration("30m")

	fastClient := &fasthttp.Client{
		ReadTimeout:         readTimeout,
		WriteTimeout:        writeTimeout,
		MaxIdleConnDuration: maxIdleConnDuration,
		RetryIf: func(request *fasthttp.Request) bool { // no retries
			return false
		},
		NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
		DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this
		DisablePathNormalizing:        true,
		MaxConnsPerHost:               100000,
		// increase DNS cache time to an hour instead of default minute
		Dial: (&fasthttp.TCPDialer{
			Concurrency:      8192,
			DNSCacheDuration: time.Hour,
		}).Dial,
	}

	return &FastClient{
		cli:                fastClient,
		wg:                 &sync.WaitGroup{},
		jsonRPCAddr:        jsonRPCAddr,
		insufficientFundRe: regexp.MustCompile(`sender balance < tx cost \(\d+ < \d+\): insufficient fund`),
		invalidNonceRe:     regexp.MustCompile(`expected (\d+)`),
	}
}

func (fc *FastClient) EthSendRawTransaction(rawTx []byte) error {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(fc.jsonRPCAddr)
	req.SetBodyRaw(rawTx)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.Set("Content-Type", "application/json")

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := fc.cli.Do(req, resp)

	insufficientFund := fc.insufficientFundRe.Match(resp.Body())
	invalidNonceMatches := fc.invalidNonceRe.FindStringSubmatch(string(resp.Body()))
	if err != nil {
		return err
	}
	var errResp types.ErrResponse
	if insufficientFund || invalidNonceMatches != nil {
		if err := json.Unmarshal(resp.Body(), &errResp); err != nil {
			panic(err)
		}
	}
	if insufficientFund {
		return errors.Wrap(types.ErrorInsufficientFund, errResp.Error.Message)
	}
	if invalidNonceMatches != nil && len(invalidNonceMatches) > 1 {
		nonce, _ := strconv.Atoi(invalidNonceMatches[1])
		return &types.NonceError{
			Message: errResp.Error.Message,
			Nonce:   uint64(nonce),
		}
	}
	return nil
}

func (fc *FastClient) EthSendRawTransactionNoWaiting(rawTx []byte) error {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(fc.jsonRPCAddr)
	req.SetBodyRaw(rawTx)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.Set("Content-Type", "application/json")

	err := fc.cli.Do(req, nil)
	if err != nil {
		return err
	}
	return nil
}

func (fc *FastClient) EthPendingNonce(addr common.Address) (uint64, error) {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getTransactionCount",
		"params":  []interface{}{addr, "latest"},
		"id":      1,
	})
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(fc.jsonRPCAddr)
	req.SetBodyRaw(reqBody)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.Set("Content-Type", "application/json")

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err := fc.cli.Do(req, resp)

	if err != nil {
		return 0, err
	}
	var nonceResp types.Response
	var errResp types.ErrResponse
	if err := json.Unmarshal(resp.Body(), &nonceResp); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(resp.Body(), &errResp); err != nil {
		panic(err)
	}

	if errResp.Error.Message != "" {
		return 0, errors.Wrap(types.ErrorFailedToFetchNonce, errResp.Error.Message)
	}
	// convert hex string to uint64
	hexNonce := nonceResp.Result.(string)
	nonce, _ := hexutil.DecodeUint64(hexNonce)
	return nonce, nil
}

func (fc *FastClient) EthSendMultipleRawTransactions(rawTxs [][]byte, cb func(*sync.Mutex, int)) (failed int64) {
	mu := sync.Mutex{}
	for i, rawTx := range rawTxs {
		fc.wg.Add(1)
		go func(w *sync.WaitGroup, mut *sync.Mutex, idx int, data []byte) {
			defer w.Done()

			err := fc.EthSendRawTransaction(data)
			// TODO: Make NoWaiting version
			// err := ctx.fastClient.EthSendRawTransactionNoWaiting(ctx, data)
			if err != nil {
				atomic.AddInt64(&failed, 1)
				log.Err(err).Msg("failed to send transaction")
				return
			}
			cb(&mu, idx)
		}(fc.wg, &mu, i, rawTx)
	}
	fc.wg.Wait()
	return failed
}
