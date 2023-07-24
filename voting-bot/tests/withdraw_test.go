package targets_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	distribution "github.com/cosmos/cosmos-sdk/x/distribution/types"
	lensclient "github.com/strangelove-ventures/lens/client"
	registry "github.com/strangelove-ventures/lens/client/chain_registry"
	"github.com/vitwit/authz-apps/voting-bot/database"
	"github.com/vitwit/authz-apps/voting-bot/types"
	"github.com/vitwit/authz-apps/voting-bot/utils"
)

func TestWithdraw(t *testing.T) {
	// Load .env file
	require.NoError(t, godotenv.Load(".env"))
	ctx, chainClient := setup(t)
	defer ctx.Database().Close()
	err := withdraw(ctx, chainClient)
	require.NoError(t, err)
	require.Equal(t, 0, 1)
}

func withdraw(ctx types.Context, chainClient lensclient.ChainClient) error {
	keys, err := ctx.Database().GetKeys()
	if err != nil {
		return err
	}
	validators, err := ctx.Database().GetValidators()
	if err != nil {
		return err
	}
	fmt.Println("KEYS: ", keys)
	fmt.Println("VALIDATORS: ", validators)

	for _, key := range keys {
		validEndpoint := "http://localhost:1317"

		for _, val := range validators {
			if val.ChainName == key.ChainName {
				var msgs []*cdctypes.Any

				accAddr, err := convertValAddrToAccAddr(ctx, val.Address, key.ChainName)
				if err != nil {
					return err
				}

				ops := types.HTTPOptions{
					Endpoint: validEndpoint + "/cosmos/authz/v1beta1/grants?granter=" + accAddr + "&grantee=" + key.KeyAddress + "&msg_url_type=/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward",
					Method:   http.MethodGet,
				}
				fmt.Println("ENDPOINT: ", ops.Endpoint)
				g1, err := getAuthzGrants(ops.Endpoint)
				if err != nil {
					return err
				}
				if len(g1) > 0 {
					msg, err := withdrawRewardsMsg(accAddr, val.Address)
					if err != nil {
						log.Printf("Error in creating withdraw rewards message for %s", val.Address)
						return err
					}

					msgs = append(msgs, msg)
				}

				ops = types.HTTPOptions{
					Endpoint: validEndpoint + "/cosmos/authz/v1beta1/grants?granter=" + accAddr + "&grantee=" + key.KeyAddress + "&msg_url_type=/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission",
					Method:   http.MethodGet,
				}
				fmt.Println("ENDPOINT: ", ops.Endpoint)
				g2, err := getAuthzGrants(ops.Endpoint)
				if err != nil {
					return err
				}

				if len(g2) > 0 {
					msg, err := withdrawCommissionMsg(val.Address)
					if err != nil {
						log.Printf("Error in creating withdraw commission message for %s", val.Address)
						return err
					}

					msgs = append(msgs, msg)
				}

				res, err := executeMsgs(chainClient, msgs, key.KeyAddress)
				if err != nil {
					log.Printf("Error in executing messages message for %s", val.Address)
					return err
				}

				rewards, err := getRewardAmount(res, "withdraw_rewards")
				if err != nil {
					log.Printf("Error in getting rewards from tx resp for chain %s. txhash: %s", key.ChainName, res.TxHash)
					return err
				}

				commission, err := getRewardAmount(res, "withdraw_commission")
				if err != nil {
					log.Printf("Error in getting rewards from tx resp for chain %s. txhash: %s", key.ChainName, res.TxHash)
					return err
				}

				fmt.Println("Rewards: ", rewards)
				fmt.Println("Commission: ", commission)

				if err := ctx.Database().AddRewards(os.Getenv("CHAINID"), "stake", val.Address, rewards.String(), commission.String()); err != nil {
					log.Printf("Failed to store reward and commission for %s on %s", val.Address, val.ChainName)
					return err
				}
			}
		}
	}
	return nil
}

func createChainClient(keyName, chainID string) (lensclient.ChainClient, error) {
	gasPrices := "2500stake"
	chainConfig := lensclient.ChainClientConfig{
		Key:            keyName,
		ChainID:        chainID,
		RPCAddr:        "tcp://localhost:26657",
		AccountPrefix:  "cosmos",
		KeyringBackend: "test",
		GasPrices:      gasPrices,
		Debug:          true,
		Timeout:        "20s",
		GasAdjustment:  1.4,
		OutputFormat:   "json",
		SignModeStr:    "direct",
		Modules:        lensclient.ModuleBasics,
	}

	curDir, err := os.Getwd()
	if err != nil {
		return lensclient.ChainClient{}, fmt.Errorf("error while getting current directory: %v", err)
	}

	// Create client object to pull chain info
	chainClient, err := lensclient.NewChainClient(zap.L(), &chainConfig, curDir, os.Stdin, os.Stdout)
	if err != nil {
		return lensclient.ChainClient{}, fmt.Errorf("failed to build new chain client for %s. Err: %v", chainID, err)
	}

	return *chainClient, nil
}

func setup(t *testing.T) (types.Context, lensclient.ChainClient) {
	db, err := database.NewDatabase()
	require.NoError(t, err)
	db.InitializeTables()

	logger := log.Logger
	ctx := types.NewContext(logger, db, nil, nil)

	chainClient, err := createChainClient(os.Getenv("GRANTEE"), os.Getenv("CHAINID"))
	require.NoError(t, err)

	chainName := os.Getenv("CHAINID")
	keyName := os.Getenv("GRANTEE")
	keyAddress := os.Getenv("grantee_addr")
	require.NoError(t, db.AddKey(chainName, keyName, keyAddress))

	valAddr := os.Getenv("val_addr")
	require.NoError(t, db.AddValidator(chainName, valAddr))

	return ctx, chainClient
}

func convertValAddrToAccAddr(ctx types.Context, valAddr, chainName string) (string, error) {
	chainInfo := registry.ChainInfo{
		Schema:       "../chain.schema.json",
		ChainName:    os.Getenv("CHAINID"),
		ChainID:      os.Getenv("CHAINID"),
		Bech32Prefix: "cosmos",
		Slip44:       118,
	}

	done := utils.SetBech32Prefixes(chainInfo)
	addr, err := utils.ValAddressFromBech32(valAddr)
	if err != nil {
		done()
		return "", err
	}

	accAddr, err := utils.AccAddressFromHexUnsafe(hex.EncodeToString(addr.Bytes()))
	if err != nil {
		return "", err
	}

	accAddrString := accAddr.String()
	done()
	return accAddrString, nil
}

func getAuthzGrants(endpoint string) ([]interface{}, error) {
	response, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Failed to read response body:", err)
		return nil, err
	}
	var jsonData map[string]interface{}
	err = json.Unmarshal(body, &jsonData)
	if err != nil {
		fmt.Println("Failed to parse JSON data:", err)
		return nil, err
	}

	grantsInterface, ok := jsonData["grants"]
	if !ok || grantsInterface == nil {
		// handle the case when "grants" is not present or nil
		return []interface{}(nil), nil
	}

	grants, ok := grantsInterface.([]interface{})
	if !ok {
		// handle the case when "grants" is not a slice of interfaces
		return []interface{}(nil), nil
	}
	return grants, nil
}

// Generate withdraw rewards message
func withdrawRewardsMsg(granter, valAddr string) (*cdctypes.Any, error) {
	var msgAny *cdctypes.Any

	msgWithdrawRewards := distribution.MsgWithdrawDelegatorReward{
		DelegatorAddress: granter,
		ValidatorAddress: valAddr,
	}

	msgAny, err := cdctypes.NewAnyWithValue(&msgWithdrawRewards)
	if err != nil {
		return nil, fmt.Errorf("error on converting msg to Any: %v", err)
	}

	return msgAny, nil
}

// Generate withdraw commission message
func withdrawCommissionMsg(valAddr string) (*cdctypes.Any, error) {
	var msgAny *cdctypes.Any

	msgWithdrawCommission := distribution.MsgWithdrawValidatorCommission{
		ValidatorAddress: valAddr,
	}

	msgAny, err := cdctypes.NewAnyWithValue(&msgWithdrawCommission)
	if err != nil {
		return nil, fmt.Errorf("error on converting msg to Any: %v", err)
	}

	return msgAny, nil
}

func executeMsgs(chainClient lensclient.ChainClient, msgs []*cdctypes.Any, keyAddr string) (*sdk.TxResponse, error) {
	req := &authz.MsgExec{
		Grantee: keyAddr,
		Msgs:    msgs,
	}

	// Send msg and get response
	res, err := chainClient.SendMsg(context.Background(), req, "")
	if err != nil {
		if res != nil {
			return nil, fmt.Errorf("failed to exec messages: code(%d) msg(%s)", res.Code, res.Logs)
		}
		return nil, fmt.Errorf("failed to exec messages.Err: %v", err)
	}

	return res, nil
}

func getRewardAmount(res *sdk.TxResponse, eventType string) (sdk.Coins, error) {
	var totalRewards sdk.Coins

	for _, event := range res.Events {
		if event.Type == eventType {
			for _, attr := range event.Attributes {
				fmt.Println("KEY :", string(attr.Key))
				fmt.Println("VALUE :", string(attr.Value))

				if string(attr.Key) == "amount" {
					rewards, err := sdk.ParseCoinsNormalized(string(attr.Value))
					if err != nil {
						return nil, fmt.Errorf("error parsing reward coins: %v", err)
					}
					for _, reward := range rewards {
						totalRewards = totalRewards.Add(reward)
					}
				}
			}
		}
	}

	return totalRewards, nil
}
