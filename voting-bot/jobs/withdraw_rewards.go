package jobs

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/slack-go/slack"
	lensclient "github.com/strangelove-ventures/lens/client"
	registry "github.com/strangelove-ventures/lens/client/chain_registry"
	"github.com/vitwit/authz-apps/voting-bot/endpoints"
	"github.com/vitwit/authz-apps/voting-bot/types"
	"github.com/vitwit/authz-apps/voting-bot/utils"
	"github.com/vitwit/authz-apps/voting-bot/voting"
	"go.uber.org/zap"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	distribution "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

const WITHDRAW_REWARDS_TYPEURL = "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward"
const WITHDRAW_COMMISSION = "/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission"

// Withdraw retrieves validator rewards and commissions using authz if
// available, and then stores them in the database.
func Withdraw(ctx types.Context) error {
	keys, err := ctx.Database().GetKeys()
	if err != nil {
		return err
	}

	validators, err := ctx.Database().GetValidators()
	if err != nil {
		return err
	}

	for _, key := range keys {
		if key.Type != "rewards" {
			continue // TODO: fetch only rewards records
		}

		chainInfo, chainClient, err := createChainClient(ctx, key.ChainName, key.KeyName)
		if err != nil {
			log.Printf("Error in creating chain client for %s chain", key.ChainName)
			return err
		}

		for _, val := range validators {
			if val.ChainName == key.ChainName {
				validEndpoint, err := endpoints.GetValidEndpointForChain(key.ChainName)
				if err != nil {
					log.Printf("Error in getting valid LCD endpoints for %s chain", key.ChainName)
					return err
				}

				var msgs []*cdctypes.Any
				granter, err := ConvertValAddrToAccAddr(ctx, val.Address, key.ChainName)
				if err != nil {
					return err
				}

				hasAuthz, err := utils.HasAuthzGrant(validEndpoint, granter, key.GranteeAddress, WITHDRAW_REWARDS_TYPEURL)
				if err != nil {
					return err
				}

				if hasAuthz {
					msg, err := withdrawRewardsMsg(granter, val.Address)
					if err != nil {
						log.Printf("Error in creating withdraw rewards message for %s", val.Address)
						return err
					}

					msgs = append(msgs, msg)
				}

				hasAuthz, err = utils.HasAuthzGrant(validEndpoint, granter, key.GranteeAddress, WITHDRAW_COMMISSION)
				if err != nil {
					return err
				}

				if hasAuthz {
					msg, err := withdrawCommissionMsg(val.Address)
					if err != nil {
						log.Printf("Error in creating withdraw commission message for %s", val.Address)
						return err
					}

					msgs = append(msgs, msg)
				}

				if len(msgs) == 0 {
					return nil
				}

				res, err := executeMsgs(chainClient, msgs, key.GranteeAddress)
				if err != nil {
					log.Printf("Error in creating withdraw commission message for %s", val.Address)
					return err
				}

				mintscanName := val.ChainName
				if newName, ok := utils.RegisrtyNameToMintscanName[val.ChainName]; ok {
					mintscanName = newName
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

				url := fmt.Sprintf("https://mintscan.io/%s/txs/%s", mintscanName, res.TxHash)
				SendMsgExecAlert(ctx, val.Address, url)

				denom, err := voting.GetChainDenom(chainInfo)
				if err != nil {
					log.Printf("Error in getting denom for chain %s", val.ChainName)
					return err
				}

				if err := ctx.Database().AddRewards(chainInfo.ChainID, denom, val.Address, rewards.String(), commission.String()); err != nil {
					log.Printf("Failed to store reward and commission for %s on %s", val.Address, val.ChainName)
					return err
				}
			}
		}
	}
	return nil
}

func createChainClient(ctx types.Context, chainName, keyName string) (registry.ChainInfo, lensclient.ChainClient, error) {
	// Fetch chain info from chain registry
	chainInfo, err := voting.GetChainInfo(ctx, chainName)
	if err != nil {
		return registry.ChainInfo{}, lensclient.ChainClient{}, fmt.Errorf("chain info not found: %v", err)
	}

	//	Use Chain info to select random endpoint
	rpc, err := chainInfo.GetRandomRPCEndpoint(ctx.Context())
	if err != nil {
		return registry.ChainInfo{}, lensclient.ChainClient{}, fmt.Errorf("failed to get random RPC endpoint on chain %s. Err: %v", chainInfo.ChainID, err)
	}

	denom, err := voting.GetChainDenom(chainInfo)
	if err != nil {
		return registry.ChainInfo{}, lensclient.ChainClient{}, fmt.Errorf("failed to get denom from chain %s: %v", chainInfo.ChainID, err)
	}

	gasPrices := "0.55" + denom

	chainConfig := utils.GetChainConfig(keyName, chainInfo, gasPrices, rpc)

	curDir, err := os.Getwd()
	if err != nil {
		return registry.ChainInfo{}, lensclient.ChainClient{}, fmt.Errorf("error while getting current directory: %v", err)
	}

	// Create client object to pull chain info
	chainClient, err := lensclient.NewChainClient(zap.L(), &chainConfig, filepath.Join(curDir, "reward-keys"), os.Stdin, os.Stdout)
	if err != nil {
		return registry.ChainInfo{}, lensclient.ChainClient{}, fmt.Errorf("failed to build new chain client for %s. Err: %v", chainInfo.ChainID, err)
	}
	return chainInfo, *chainClient, nil
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
			return nil, fmt.Errorf("failed to vote on proposal: code(%d) msg(%s)", res.Code, res.Logs)
		}
		return nil, fmt.Errorf("failed to vote.Err: %v", err)
	}

	return res, nil
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

// SendMsgExecAlert which sends alerts on withdraw rewards and commission txs
func SendMsgExecAlert(ctx types.Context, valAddr, URL string) error {
	api := ctx.Slacker().APIClient()

	attachment := []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", fmt.Sprintf("Withdraw rewards and commission executed for %s\nTransaction broadcasted: %s", valAddr, URL), false, false),
		),
	}

	_, _, err := api.PostMessage(
		ctx.Config().Slack.ChannelID,
		slack.MsgOptionBlocks(attachment...),
	)
	if err != nil {
		return err
	}

	return nil
}

func getRewardAmount(res *sdk.TxResponse, eventType string) (sdk.Coins, error) {
	var totalRewards sdk.Coins

	for _, event := range res.Events {
		if event.Type == eventType {
			for _, attr := range event.Attributes {
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
