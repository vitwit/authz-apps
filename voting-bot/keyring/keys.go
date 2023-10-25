package keyring

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	lensclient "github.com/strangelove-ventures/lens/client"
	"github.com/vitwit/authz-apps/voting-bot/types"
	"github.com/vitwit/authz-apps/voting-bot/utils"
	"go.uber.org/zap"
)

// Creates keys using chain name and chain registry
func CreateKeys(ctx types.Context, chainName, keyName, keyType string) error {
	// Fetch chain info from chain registry
	cr := ctx.ChainRegistry()
	chainInfo, err := cr.GetChain(context.Background(), chainName)
	if err != nil {
		return fmt.Errorf("failed to get chain info. Err: %v", err)
	}

	//	Use Chain info to select random endpoint
	rpc, err := chainInfo.GetRandomRPCEndpoint(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get random RPC endpoint on chain %s. Err: %v", chainInfo.ChainID, err)
	}

	chainConfig := utils.GetChainConfig("", chainInfo, "", rpc)

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
	if hasMnemonicSeed(keyType) {
		seed, err := readSeedFile(keyType)
		if err != nil {
			return err
		}

		address, err = chainClient.RestoreKey(keyName, seed, uint32(chainInfo.Slip44))
		if err != nil {
			return fmt.Errorf("error while restoring account: %v", err)
		}
	} else {
		res, err := chainClient.AddKey(keyName, uint32(chainInfo.Slip44))
		if err != nil {
			return fmt.Errorf("error while adding key: %v", err)
		}

		// store Mnemonic seed
		if err := storeMnemonicSeed(keyType, res.Mnemonic); err != nil {
			return err
		}

		address = res.Address
	}

	if err := ctx.Database().AddAuthzKey(chainName, keyName, address, keyType); err != nil {
		return err
	}

	return nil
}

func readSeedFile(keyType string) (string, error) {
	seedFile := getSeedfilePath(keyType)
	stream, err := os.ReadFile(seedFile)
	if err != nil {
		return "", err
	}

	return string(stream), err
}

func getSeedfilePath(keyType string) string {
	var seedFile string
	switch keyType {
	case "voting":
		seedFile = filepath.Join("voting-keys", "keys", "seed.txt")
	case "rewards":
		seedFile = filepath.Join("reward-keys", "keys", "seed.txt")
	default:
		panic("invalid option. Key-type must be either \"voting\" or \"rewards\"")
	}

	return seedFile
}

// hasMnemonicSeed returns true if SEED_FILE exists
func hasMnemonicSeed(keyType string) bool {
	seedFile := getSeedfilePath(keyType)
	if _, err := os.Stat(seedFile); errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

// storeMnemonicSeed stores mnemonic seed string the SEED_FILE.
func storeMnemonicSeed(keyType, seed string) error {
	var keyTypeName string
	if keyType == "voting" {
		if _, err := os.Stat(filepath.Join("voting-keys", "keys")); os.IsNotExist(err) {
			err := os.MkdirAll(filepath.Join("voting-keys", "keys"), os.ModePerm)
			if err != nil {
				return err
			}
		}
		keyTypeName = "voting-keys"
	} else if keyType == "rewards" {
		if _, err := os.Stat(filepath.Join("reward-keys", "keys")); os.IsNotExist(err) {
			err := os.MkdirAll(filepath.Join("reward-keys", "keys"), os.ModePerm)
			if err != nil {
				return err
			}
		}
		keyTypeName = "reward-keys"
	} else {
		return errors.New("invalid option. Key-type must be either \"voting\" or \"rewards\"")
	}

	f, err := os.Create(filepath.Join(keyTypeName, "keys", "seed.txt"))
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString(seed)
	return err
}
