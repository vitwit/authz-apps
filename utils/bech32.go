package utils

import (
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"

	registry "github.com/strangelove-ventures/lens/client/chain_registry"
)

var prefixMutex sync.Mutex

func SetBech32Prefixes(chainInfo registry.ChainInfo) func() {
	prefixMutex.Lock()

	// Set the Bech32 prefixes
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(chainInfo.Bech32Prefix, chainInfo.Bech32Prefix+"pub")
	config.SetBech32PrefixForValidator(chainInfo.Bech32Prefix+"valoper", chainInfo.Bech32Prefix+"valoperpub")

	return prefixMutex.Unlock
}
