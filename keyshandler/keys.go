package keyshandler

import (
	"context"
	"log"
	"os"

	"github.com/likhita-809/lens-bot/database"
	lensclient "github.com/strangelove-ventures/lens/client"
	registry "github.com/strangelove-ventures/lens/client/chain_registry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keys struct {
	db *database.Sqlitedb
}

// Creates keys using chain name and chain registry
func (k Keys) CreateKeys(chainName, keyName string) error {
	// Fetch chain info from chain registry
	cr := registry.DefaultChainRegistry(zap.New(zapcore.NewNopCore()))

	chainInfo, err := cr.GetChain(context.Background(), chainName)
	if err != nil {
		log.Printf("failed to get chain info. Err: %v", err)
	}

	//	Use Chain info to select random endpoint
	rpc, err := chainInfo.GetRandomRPCEndpoint(context.Background())
	if err != nil {
		log.Printf("failed to get random RPC endpoint on chain %s. Err: %v", chainInfo.ChainID, err)
	}

	chainConfig := lensclient.ChainClientConfig{
		ChainID:        chainInfo.ChainID,
		RPCAddr:        rpc,
		AccountPrefix:  chainInfo.Bech32Prefix,
		KeyringBackend: "test",
		Debug:          true,
		Timeout:        "20s",
		OutputFormat:   "json",
		SignModeStr:    "direct",
		Modules:        lensclient.ModuleBasics,
	}

	curDir, err := os.Getwd()
	if err != nil {
		log.Printf("error while getting current directory: %v", err)
	}

	// Create client object to pull chain info
	chainClient, err := lensclient.NewChainClient(zap.L(), &chainConfig, curDir, os.Stdin, os.Stdout)
	if err != nil {
		log.Printf("failed to build new chain client for %s. Err: %v", chainInfo.ChainID, err)
	}

	res, err := chainClient.AddKey(keyName, sdk.CoinType)
	if err != nil {
		log.Printf("error while adding key: %v", err)
	}

	chainConfig.Key = keyName

	k.db.AddKey(chainName, keyName, res.Address)
	return nil
}
