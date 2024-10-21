package offchain_feeding

const (
	DefaultAccNum       = 100
	DefaultValidatorNum = 1
)

type Config struct {
	AccNum     int    `toml:"acc_num"`
	GenesisLoc string `toml:"genesis_loc"`
	BechPrefix string `toml:"bech_prefix"`
	Denom      string `toml:"denom"`
}

func DefaultConfig() Config {
	return Config{
		AccNum:     DefaultAccNum,
		BechPrefix: "canto",
		Denom:      "acanto",
	}
}
