package targets

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"github.com/slack-go/slack"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/vitwit/authz-apps/voting-bot/types"
	"github.com/vitwit/authz-apps/voting-bot/utils"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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

		chainInfo, err := ctx.ChainRegistry().GetChain(ctx.Context(), key.ChainName)
		if err != nil {
			return err
		}

		grpcEndpoint, err := chainInfo.GetActiveGRPCEndpoint(ctx.Context())
		if err != nil {
			log.Printf("Error in getting valid GRPC endpoint for %s chain", key.ChainName)
			return err
		}

		addr := key.KeyAddress
		err = AlertOnLowBalance(ctx, grpcEndpoint, addr, baseDenom, coinDecimals)
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
	conn, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Printf("Failed to connect to %s: %v", endpoint, err)
		return err
	}
	defer conn.Close()

	ctx1, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	client := banktypes.NewQueryClient(conn)

	balance, err := client.Balance(ctx1, &banktypes.QueryBalanceRequest{
		Address: addr,
		Denom:   denom,
	})
	if err != nil {
		log.Printf("Failed to getbalance %s: %v", addr, err)
		return err
	}

	if balance.Balance.IsGTE(sdk.NewCoin(denom, sdk.NewInt(1).Mul(sdk.NewInt(coinDecimals)))) {
		err := SendLowBalanceAlerts(ctx, addr, balance.Balance.Amount.String(), balance.Balance.Denom)
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
