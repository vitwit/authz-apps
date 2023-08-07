package targets

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"cosmossdk.io/math"
	"github.com/slack-go/slack"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/vitwit/authz-apps/voting-bot/endpoints"
	"github.com/vitwit/authz-apps/voting-bot/types"
	"github.com/vitwit/authz-apps/voting-bot/utils"
)

// Gets accounts with low balances (i.e., 1ATOM) and alerts on them
func GetLowBalAccs(ctx types.Context) error {
	db := ctx.Database()
	keys, err := db.GetKeys()
	if err != nil {
		return fmt.Errorf("error while getting keys from db: %v", err)
	}

	for _, key := range keys {
		var baseDenom string
		var coinDecimals int64 = 6
		if info, ok := utils.ChainNameToDenomInfo[key.ChainName]; ok {
			baseDenom = info.BaseDenom
			coinDecimals = info.DenomUnits
		} else {
			log.Printf("Error in getting valid LCD endpoints for %s chain", key.ChainName)
			return fmt.Errorf("chain %s is not supported", key.ChainName)
		}

		endpoint, err := endpoints.GetValidEndpointForChain(key.ChainName)
		if err != nil {
			log.Printf("Error in getting valid LCD endpoints for %s chain", key.ChainName)
			return err
		}
		addr := key.KeyAddress
		err = AlertOnLowBalance(ctx, endpoint, addr, baseDenom, coinDecimals)
		if err != nil {
			log.Printf("error on sending low balance alert: %v", err)
			return err
		}
	}
	return nil
}

// Gets balance of an account and alerts if the balance is low
func AlertOnLowBalance(ctx types.Context, endpoint, addr, denom string, coinDecimals int64) error {
	creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: false})
	ctx1, _ := context.WithTimeout(context.Background(), 5*time.Second)
	grpcConn, err := grpc.DialContext(ctx1, endpoint, grpc.WithTransportCredentials(creds))
	if err != nil {
		return err
	}
	defer grpcConn.Close()
	bankClient := banktypes.NewQueryClient(grpcConn)
	resp, err := bankClient.Balance(
		context.Background(),
		&banktypes.QueryBalanceRequest{Address: addr, Denom: "stake"},
	)

	amount, ok := sdk.NewIntFromString(resp.Balance.Amount.String())
	if !ok {
		return fmt.Errorf("unable to convert amount string to int")
	}
	coin := sdk.NewCoin(resp.Balance.Denom, amount)
	if !coin.Amount.GT(math.NewInt(1).Mul(sdk.NewInt(coinDecimals))) {
		err := SendLowBalanceAlerts(ctx, addr, resp.Balance.Amount.String(), resp.Balance.Denom)
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
