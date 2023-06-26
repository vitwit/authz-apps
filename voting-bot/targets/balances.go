package targets

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"cosmossdk.io/math"
	"github.com/slack-go/slack"

	"github.com/vitwit/authz-apps/voting-bot/endpoints"
	"github.com/vitwit/authz-apps/voting-bot/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Gets accounts with low balances (i.e., 1ATOM) and alerts on them
func GetLowBalAccs(ctx types.Context) error {
	db := ctx.Database()
	keys, err := db.GetKeys()
	if err != nil {
		return fmt.Errorf("error while getting keys from db: %v", err)
	}

	var denom string
	for _, key := range keys {
		chainInfo, err := ctx.ChainRegistry().GetChain(context.Background(), key.ChainName)
		if err != nil {
			return fmt.Errorf("chain info not found: %v: %v", key.ChainName, err)
		}

		assetList, err := chainInfo.GetAssetList(context.Background())
		if err != nil {
			log.Printf("Error while getting asset list for %s chain", key.ChainName)
			return err
		}

		if len(assetList.Assets) > 0 {
			denom = assetList.Assets[0].Base
		} else {
			return fmt.Errorf("base denom unit not found for %s chain", key.ChainName)
		}

		endpoint, err := endpoints.GetValidEndpointForChain(key.ChainName)
		if err != nil {
			log.Printf("Error in getting valid LCD endpoints for %s chain", key.ChainName)
			return err
		}
		addr := key.KeyAddress
		err = AlertOnLowBalance(ctx, endpoint, addr, denom)
		if err != nil {
			log.Printf("error on sending low balance alert: %v", err)
			return err
		}
	}
	return nil
}

// Gets balance of an account and alerts if the balance is low
func AlertOnLowBalance(ctx types.Context, endpoint, addr, denom string) error {
	ops := types.HTTPOptions{
		Endpoint:    endpoint + "/cosmos/bank/v1beta1/balances/" + addr + "/by_denom",
		Method:      http.MethodGet,
		QueryParams: types.QueryParams{"denom": denom},
	}

	resp, err := endpoints.HitHTTPTarget(ops)
	if err != nil {
		log.Printf("Error in external rpc: %v", err)
		log.Printf("⛔⛔ Unreachable to EXTERNAL RPC :: %s and the ERROR is : %v\n\n", ops.Endpoint, err.Error())
		return err
	}

	var balance types.Balance
	err = json.Unmarshal(resp.Body, &balance)
	if err != nil {
		log.Printf("Error while unmarshalling the balances: %v", err)
		return err
	}

	amount, ok := sdk.NewIntFromString(balance.Balance.Amount)
	if !ok {
		return fmt.Errorf("unable to convert amount string to int")
	}
	coin := sdk.NewCoin(balance.Balance.Denom, amount)
	if !coin.Amount.GT(math.NewInt(1)) {
		err := SendLowBalanceAlerts(ctx, addr, balance.Balance.Amount, balance.Balance.Denom)
		if err != nil {
			log.Printf("error while sending low balance alert: %v", err)
			return err
		}
	}

	return nil
}

// SendLowBalanceAlerts which sends alerts on low balance grantee accounts
func SendLowBalanceAlerts(ctx types.Context, addr, amount, denom string) error {
	api := ctx.Slacker().APIClient()

	attachment := []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", fmt.Sprintf("%s is low on balance\nAvailable balance: %s%s", addr, amount, denom), false, false),
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
