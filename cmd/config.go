package cmd

import (
	"os"

	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog/log"

	"loadtester/cmd/evmtx"
	"loadtester/cmd/offchain_feeding"
)

var (
	DefaultConfigPath = "./config.toml"
)

type CommonConfig struct {
	EthJsonRpcAddr string `toml:"eth_jsonrpc_addr"`
}

// Config defines all necessary configuration parameters.
type Config struct {
	CommonConfig          CommonConfig            `toml:"common"`
	EvmTxConfig           evmtx.Config            `toml:"evmtx"`
	OffchainFeedingConfig offchain_feeding.Config `toml:"offchain_feeding"`
}

func DefaultConfig() Config {
	return Config{
		CommonConfig: CommonConfig{
			EthJsonRpcAddr: "http://localhost:8545",
		},
		EvmTxConfig:           evmtx.DefaultConfig(),
		OffchainFeedingConfig: offchain_feeding.DefaultConfig(),
	}
}

func MustRead(configPath string) *Config {
	if configPath == "" {
		panic("empty configuration path")
	}
	log.Debug().Msg("read config file")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal().Msgf("failed to read config: %v", err)
	}
	log.Debug().Msg("done reading config file")

	return MustParseString(configData)
}

// ParseString attempts to read and parse  config from the given string bytes.
// An error reading or parsing the config results in a panic.
func MustParseString(configData []byte) *Config {
	cfg := DefaultConfig()

	log.Debug().Msg("parsing config data")
	err := toml.Unmarshal(configData, &cfg)
	if err != nil {
		log.Fatal().Msgf("failed to decode config: %s", err)
	}
	log.Debug().Msg("done parsing config data")

	return &cfg
}
