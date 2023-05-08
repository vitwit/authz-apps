package keyshandler

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/likhita-809/lens-bot/database"
	lensclient "github.com/strangelove-ventures/lens/client"
	registry "github.com/strangelove-ventures/lens/client/chain_registry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Keys struct {
	Db *database.Sqlitedb
}

// Creates keys using chain name and chain registry
func (k Keys) CreateKeys(chainName, keyName string) error {
	// Fetch chain info from chain registry
	cr := registry.DefaultChainRegistry(zap.New(zapcore.NewNopCore()))

	chainInfo, err := cr.GetChain(context.Background(), chainName)
	if err != nil {
		return fmt.Errorf("failed to get chain info. Err: %v", err)
	}

	//	Use Chain info to select random endpoint
	rpc, err := chainInfo.GetRandomRPCEndpoint(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get random RPC endpoint on chain %s. Err: %v", chainInfo.ChainID, err)
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
		return fmt.Errorf("error while getting current directory: %v", err)
	}

	// Create client object to pull chain info
	chainClient, err := lensclient.NewChainClient(zap.L(), &chainConfig, curDir, os.Stdin, os.Stdout)
	if err != nil {
		return fmt.Errorf("failed to build new chain client for %s. Err: %v", chainInfo.ChainID, err)
	}

	// If a mnemonic seed phrase exists, use the same seed phrase for all accounts.
	var address string
	if k.hasMnemonicSeed() {
		seed, err := k.readSeedFile()
		if err != nil {
			return err
		}

		address, err = chainClient.RestoreKey(keyName, seed, uint32(chainInfo.Slip44))
		if err != nil {
			return fmt.Errorf("account already exist for this network")
		}
	} else {
		res, err := chainClient.AddKey(keyName, uint32(chainInfo.Slip44))
		if err != nil {
			return fmt.Errorf("account already exist for this network")
		}

		// store Mnemonic seed
		if err := k.storeMnemonicSeed(res.Mnemonic); err != nil {
			return err
		}

		address = res.Address
	}

	chainConfig.Key = keyName
	if err := k.Db.AddKey(chainName, keyName, address); err != nil {
		return err
	}

	return nil
}

var SEED_FILE = filepath.Join("keys", "seed.txt")

func (k Keys) readSeedFile() (string, error) {
	stream, err := os.ReadFile(SEED_FILE)
	if err != nil {
		return "", err
	}

	return string(stream), err
}

// hasMnemonicSeed returns true if SEED_FILE exists
func (k Keys) hasMnemonicSeed() bool {
	if _, err := os.Stat(SEED_FILE); errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

// storeMnemonicSeed stores mnemonic seed string the SEED_FILE.
func (k Keys) storeMnemonicSeed(seed string) error {
	if _, err := os.Stat("keys"); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir("keys", os.ModePerm)
		if err != nil {
			return err
		}
	}

	f, err := os.Create(SEED_FILE)

	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString(seed)
	return err
}
