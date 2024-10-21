package evmtx

const (
	DefaultGasLimit     = 200000
	DefaultGasPrice     = 201417240
	DefaultSendingAmt   = 1
	DefaultChainId      = 1124124
	DefaultDuration     = "10m"
	DefaultTps          = 10
	DefaultAccNum       = 100
	DefaultValidatorNum = 1
	DefaultScenario     = ScenarioEthTransferToRandom
)

const (
	// eth transfer to random recipient
	ScenarioEthTransferToRandom = "eth_transfer_to_random"
	// eth transfer to known recipient which is one of the test accounts
	ScenarioEthTransferToKnown = "eth_transfer_to_known"
	// eth transfer to self
	ScenarioEthTransferToSelf = "eth_transfer_to_self"
	// TODO: erc20 transfer
	//ScenarioErc20Transfer = "erc20_transfer"
)

type Config struct {
	GasLimit               int64  `toml:"gas_limit"`
	GasPrice               int64  `toml:"gas_price"`
	SendingAmt             int64  `toml:"sending_amt"`
	ChainID                int64  `toml:"chain_id"`
	Duration               string `toml:"duration"`
	TransactionPerTimeUnit int    `toml:"tpu"`
	TimeUnit               string `toml:"time_unit"`
	AccNum                 int    `toml:"acc_num"`
	Scenario               string `toml:"scenario"`
}

func DefaultConfig() Config {
	return Config{
		GasLimit:               DefaultGasLimit,
		GasPrice:               DefaultGasPrice,
		SendingAmt:             DefaultSendingAmt,
		ChainID:                DefaultChainId,
		Duration:               DefaultDuration,
		TransactionPerTimeUnit: DefaultTps,
		AccNum:                 DefaultAccNum,
		Scenario:               DefaultScenario,
	}
}
