package alerting

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/likhita-809/lens-bot/config"
	"github.com/likhita-809/lens-bot/database"
	"github.com/likhita-809/lens-bot/keyshandler"
	"github.com/likhita-809/lens-bot/voting"
	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
	registry "github.com/strangelove-ventures/lens/client/chain_registry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Slackbot struct {
	bot  *slacker.Slacker
	db   *database.Sqlitedb
	cfg  *config.Config
	key  *keyshandler.Keys
	vote *voting.Vote
}

// Creates a new bot client
func NewBotClient(config *config.Config, db *database.Sqlitedb, key *keyshandler.Keys, vote *voting.Vote) *Slackbot {
	bot := slacker.NewClient(config.Slack.BotToken, config.Slack.AppToken)
	return &Slackbot{
		bot:  bot,
		db:   db,
		cfg:  config,
		key:  key,
		vote: vote,
	}
}

// Creates and initialises commands
func (a *Slackbot) Initializecommands() error {
	// Command to register validator address with chain name
	a.bot.Command("register-validator <chainName> <validatorAddress>", &slacker.CommandDefinition{
		Description: "registers a new validator",
		Examples:    []string{"register-validator cosmoshub cosmos1a..."},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			chainName := request.Param("chainName")
			validatorAddress := request.Param("validatorAddress")

			cr := registry.DefaultChainRegistry(zap.New(zapcore.NewNopCore()))
			chainInfo, err := cr.GetChain(context.Background(), chainName)
			if err != nil {
				response.Reply(fmt.Sprintf("failed to get chain information from registry: %s", err.Error()))
				panic(err)
			}

			config := sdk.GetConfig()
			config.SetBech32PrefixForAccount(chainInfo.Bech32Prefix, chainInfo.Bech32Prefix+"pub")
			config.SetBech32PrefixForValidator(chainInfo.Bech32Prefix+"valoper", chainInfo.Bech32Prefix+"valoperpub")
			config.Seal()

			_, err = sdk.ValAddressFromBech32(validatorAddress)
			if err != nil {
				response.Reply("Invalid validator address")
			} else {
				isExists := a.db.HasValidator(validatorAddress)
				if isExists {
					response.Reply("Validator is already registered")
				} else {
					a.db.AddValidator(chainName, validatorAddress)
					r := fmt.Sprintf("Your validator %s is successfully registered", validatorAddress)
					response.Reply(r)
				}
			}
		},
	})

	// Command to remove validator address from db
	a.bot.Command("remove-validator <validatorAddress>", &slacker.CommandDefinition{
		Description: "remove an existing validator",
		Examples:    []string{"remove-validator cosmos1a..."},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			validatorAddress := request.Param("validatorAddress")
			if !a.db.HasValidator(validatorAddress) {
				response.Reply("Cannot delete a validator which is not in the registered validators")
			} else {
				a.db.RemoveValidator(validatorAddress)
				r := fmt.Sprintf("Your validator %s is successfully removed", validatorAddress)
				response.Reply(r)
			}
		},
	})

	// Creates keys which are used for voting
	a.bot.Command("create-key <chainName> <keyNameOptional>", &slacker.CommandDefinition{
		Description: "create a new account with key name",
		Examples:    []string{"create-key myKey"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			keyName := request.StringParam("keyNameOptional", "default")
			chainName := request.Param("chainName")
			err := a.key.CreateKeys(chainName, keyName)
			if err != nil {
				response.Reply(err.Error())
			} else {
				NewSlackAlerter().Send(fmt.Sprintf("Successfully created your key with name %s", keyName), a.cfg.Slack.BotToken, a.cfg.Slack.ChannelID)
			}
		},
	})

	// Vote command is used to vote on the proposals based on proposal Id, validator address with vote option using keys stored from db.
	a.bot.Command(
		"vote <chainId> <proposalId> <granterAddress> <voteOption> <granteeKeyname> <gasPrices> <memoOptional> <metadataOptional>",
		&slacker.CommandDefinition{
			Description: "votes on the proposal",
			Examples:    []string{"vote cosmoshub-4 12 YES 0.25uatom example_memo example_metadata"},
			Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
				chainID := request.Param("chainId")
				pID := request.Param("proposalId")
				granter := request.Param("granterAddress")
				voteOption := request.Param("voteOption")
				fromKey := request.Param("granteeKeyname")
				metadata := request.StringParam("metadataOptional", "")
				memo := request.StringParam("memoOptional", "")
				gasPrices := request.StringParam("gasPrices", "")
				if len(memo) > 1 {
					memo = strings.Replace(memo, "_", " ", -1)
				}

				if len(metadata) > 1 {
					metadata = strings.Replace(metadata, "_", " ", -1)
				}

				result, err := a.vote.ExecVote(chainID, pID, granter, voteOption, fromKey, metadata, memo, gasPrices)
				if err != nil {
					log.Printf("error on executing vote: %v", err)
				}

				response.Reply(result)
			},
		},
	)

	// Lists all keys stored in the database
	a.bot.Command("list-keys", &slacker.CommandDefinition{
		Description: "lists all keys",
		Examples:    []string{"list-keys"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			keys, err := a.db.GetKeys()
			if err != nil {
				response.ReportError(err)
			} else {

				apiClient := botCtx.APIClient()
				event := botCtx.Event()

				var blocks []slack.Block
				for _, key := range keys {
					blocks = append(blocks, slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*%s* ---- *%s* ---- *%s*", key.ChainName, key.KeyName, key.KeyAddress), false, false),
						nil, nil))
				}

				attachment := []slack.Block{
					slack.NewHeaderBlock(slack.NewTextBlockObject("plain_text", "Network ---- Key name---- Key address", false, false)),
				}
				attachment = append(attachment, blocks...)

				if event.ChannelID != "" {
					_, _, err := apiClient.PostMessage(event.ChannelID, slack.MsgOptionBlocks(attachment...))
					if err != nil {
						response.ReportError(err)
					}
				}
			}
		},
	})

	// Command to list all registered validators
	a.bot.Command("list-validators", &slacker.CommandDefinition{
		Description: "lists all validators addresses with associated chains",
		Examples:    []string{"list-validators"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			validators, err := a.db.GetValidators()
			if err != nil {
				response.ReportError(err)
			}

			apiClient := botCtx.APIClient()
			event := botCtx.Event()

			var blocks []slack.Block
			for _, val := range validators {
				blocks = append(blocks, slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*%s* ---- *%s*", val.ChainName, val.Address), false, false),
					nil, nil))
			}

			attachment := []slack.Block{
				slack.NewHeaderBlock(slack.NewTextBlockObject("plain_text", "Network ---- Validator address", false, false)),
			}
			attachment = append(attachment, blocks...)
			if event.ChannelID != "" {
				_, _, err := apiClient.PostMessage(event.ChannelID, slack.MsgOptionBlocks(attachment...))
				if err != nil {
					response.ReportError(err)
				}
			}
		},
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := a.bot.Listen(ctx)
	if err != nil {
		return fmt.Errorf("%s", err)
	}
	return err
}

// Send allows bot to send a slack alert to the configured channelID
func (s slackAlert) Send(msgText, botToken string, channelID string) error {
	// Create a new client to slack by giving token
	// Set debug to true while developing
	client := slack.New(botToken, slack.OptionDebug(true))

	// Create the Slack attachment that we will send to the channel
	attachment := slack.Attachment{
		Pretext: "Lens Bot Message",
		Title:   msgText,
	}

	// PostMessage will send the message away.
	// First parameter is just the channelID, makes no sense to accept it
	_, timestamp, err := client.PostMessage(
		channelID,
		slack.MsgOptionAttachments(attachment),
	)
	if err != nil {
		return err
	}
	log.Printf("Message sent at %s", timestamp)
	return nil
}
