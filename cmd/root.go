package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"loadtester/clients"
	"loadtester/cmd/evmtx"
	"loadtester/cmd/offchain_feeding"
)

var rootCmd = &cobra.Command{
	Use:   "loadtester",
	Short: "loadtester",
	Long:  `loadtester`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Error().Err(err).Msg("failed to execute command")
	}
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	fmt.Println(` _     ___    _    ____ _____ _____ ____ _____ _____ ____  
| |   / _ \  / \  |  _ \_   _| ____/ ___|_   _| ____|  _ \ 
| |  | | | |/ _ \ | | | || | |  _| \___ \ | | |  _| | |_) |
| |__| |_| / ___ \| |_| || | | |___ ___) || | | |___|  _ < 
|_____\___/_/   \_\____/ |_| |_____|____/ |_| |_____|_| \_\`)

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05.000"}) // set pretty logging
	cfg := MustRead(DefaultConfigPath)
	ethRpc := clients.NewFastClient(cfg.CommonConfig.EthJsonRpcAddr)
	rootCmd.AddCommand(evmtx.NewEvmTxCmd(cfg.EvmTxConfig, ethRpc))
	rootCmd.AddCommand(offchain_feeding.NewOffchainFeedingCmd(cfg.OffchainFeedingConfig))
}
