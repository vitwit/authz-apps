package utils

import (
	lensclient "github.com/strangelove-ventures/lens/client"
	"github.com/strangelove-ventures/lens/client/chain_registry"
)

func GetChainConfig(from string, chainInfo chain_registry.ChainInfo, gasPrices string, rpc string, gasAdjustment float64) lensclient.ChainClientConfig {
	return lensclient.ChainClientConfig{
		Key:            from,
		ChainID:        chainInfo.ChainID,
		RPCAddr:        rpc,
		AccountPrefix:  chainInfo.Bech32Prefix,
		KeyringBackend: "test",
		GasPrices:      gasPrices,
		Debug:          true,
		Timeout:        "30s",
		GasAdjustment:  gasAdjustment,
		OutputFormat:   "json",
		SignModeStr:    "direct",
		Modules:        lensclient.ModuleBasics,
	}
}
