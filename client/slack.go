package client

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/likhita-809/lens-bot/config"
	"github.com/likhita-809/lens-bot/database"
	"github.com/likhita-809/lens-bot/keyring"
	"github.com/likhita-809/lens-bot/targets"
	"github.com/likhita-809/lens-bot/utils"
	"github.com/likhita-809/lens-bot/voting"
	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
	registry "github.com/strangelove-ventures/lens/client/chain_registry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Slackbot struct {
	bot  *slacker.Slacker
	db   *database.Sqlitedb
	data *targets.Data
	cfg  *config.Config
	key  *keyring.Keys
	vote *voting.Vote
}

// Creates a new bot client
func NewBotClient(config *config.Config, db *database.Sqlitedb, data *targets.Data, key *keyring.Keys, vote *voting.Vote) *Slackbot {
	bot := slacker.NewClient(config.Slack.BotToken, config.Slack.AppToken)
	return &Slackbot{
		bot:  bot,
		db:   db,
		data: data,
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

			done := utils.SetBech32Prefixes(chainInfo)
			_, err = utils.ValAddressFromBech32(validatorAddress)
			done()

			if err != nil {
				response.Reply(fmt.Sprintf("invalid validator address: %v", err))
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
				response.ReportError(fmt.Errorf("cannot delete a validator which is not in the registered validators"))
			} else {
				a.db.RemoveValidator(validatorAddress)
				r := fmt.Sprintf("Your validator %s is successfully removed", validatorAddress)
				response.Reply(r)
			}
		},
	})

	// Creates keys which are used for voting
	a.bot.Command("create-key <chainName> <keyNameOptional>", &slacker.CommandDefinition{
		Description: "create a new account with key name.\nKeys need to be funded manually and given authorization to vote in order to use them while voting.\nThe granter must give the vote authorization to the grantee key before the voting can proceed.\nThe authorization to a grantee can be given by using the following command:\nFor Cosmos chain:\nUsage: simd tx authz grant <grantee> <authorization_type> --msg-type <msg_type> --from <granter> [flags]\nExample: simd tx authz grant cosmos1... --msg-type /cosmos.gov.v1beta1.MsgVote --from granter\nThe authorized keys can then be funded to have the ability to vote on behalf of the granter.\nThe following command can be used to fund the key:\nsimd tx bank send [from_key_or_address] [to_address] [amount] [flags]",
		Examples:    []string{"create-key cosmoshub myKey"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			keyName := request.StringParam("keyNameOptional", "")
			chainName := request.Param("chainName")
			if keyName == "" {
				keyName = chainName
			}

			err := a.key.CreateKeys(chainName, keyName)
			if err != nil {
				response.Reply(err.Error())
			} else {
				response.Reply(fmt.Sprintf("Successfully created your key with name %s.\n *NOTE*\n *This key cannot be used in voting until it has the vote authorization from granter and got funded. The vote authorization can be given using the following command:*\n ```simd tx authz grant <grantee> <authorization_type=generic> --msg-type /cosmos.gov.v1beta1.MsgVote  --from <granter> [flags]```\n\n *The authorized keys can be funded using the following command:* \n ```simd tx bank send [from_key_or_address] [to_address] [amount] [flags]```\n", keyName))
			}
		},
	})

	// Command to list all the commands present
	a.bot.Command("list-commands", &slacker.CommandDefinition{
		Description: "Lists all commands",
		Examples:    []string{"list-commands"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			r := " *SLACK BOT COMMANDS* \n\n *• register-validator*: registers the validator using chain name and validator address\n```Command : register-validator <chainName> <validatorAddress>```\n *• remove-validator* : removes an existing validator data using validator address\n```Command:remove-validator <validatorAddress>```\n *• list-keys* : Lists all keys\n```Command:list-keys```\n *• list-proposals* : Lists all Active unvoted proposals \n```Command:list-proposals```\n *• list-validators* : List of all registered validators addresses with associated chains\n```Command:list-validators```\n* • vote* : votes on a proposal\n```Command:vote <chainName> <proposalId> <voteOption> <gasPrices> <memoOptional> <metadataOptional>\n```\n* • votes-history* : Lists history of all votes for a given chain\n```Command:votes-history <chainName> <startDate> <endDateOptional>\n```\n *• create-key* : Create a new account with key name. This key name is used while voting\n```Command:create-key <chainName> <keyNameOptional>```\n"
			response.Reply(r)
		},
	})

	// Vote command is used to vote on the proposals based on proposal Id with vote option using key stored from db.
	a.bot.Command(
		"vote <chainName> <proposalId> <voteOption> <gasPrices> <memoOptional> <metadataOptional>",
		&slacker.CommandDefinition{
			Description: "votes on the proposal",
			Examples:    []string{"vote cosmoshub 12 YES 0.25uatom example_memo example_metadata"},
			Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
				chainName := request.Param("chainName")
				pID := request.Param("proposalId")

				address, err := a.db.GetChainValidator(chainName)
				if err != nil {
					response.ReportError(fmt.Errorf("failed to get validator address from the database: %v", err))
					return
				}

				cr := registry.DefaultChainRegistry(zap.New(zapcore.NewNopCore()))
				chainInfo, err := a.vote.GetChainInfo(chainName, cr)
				if err != nil {
					response.ReportError(fmt.Errorf("failed to get chain-info: %v", err))
					return
				}

				done := utils.SetBech32Prefixes(chainInfo)
				hexAddr, err := utils.ValAddressFromBech32(address)
				if err != nil {
					done()
					response.ReportError(fmt.Errorf("error while getting validator address of chain %s", chainName))
					return
				}

				granter, err := utils.AccAddressFromHexUnsafe(hex.EncodeToString(hexAddr.Bytes()))
				done()

				if err != nil {
					response.ReportError(fmt.Errorf("error while decoding validator address %s", chainName))
					return
				}

				voteOption := request.Param("voteOption")
				fromKey, err := a.db.GetChainKey(chainName)
				if err != nil {
					response.ReportError(fmt.Errorf("error while getting key address of chain %s", chainName))
					return
				}

				metadata := request.StringParam("metadataOptional", "")
				memo := request.StringParam("memoOptional", "")
				gasPrices := request.StringParam("gasPrices", "")
				if len(memo) > 1 {
					memo = strings.Replace(memo, "_", " ", -1)
				}

				if len(metadata) > 1 {
					metadata = strings.Replace(metadata, "_", " ", -1)
				}

				result, err := a.vote.ExecVote(chainName, pID, granter.String(), voteOption, fromKey, metadata, memo, gasPrices, response)
				if err != nil {
					log.Printf("error on executing vote: %v", err)
					response.ReportError(fmt.Errorf("error on executing vote: %v", err))
					return
				}

				response.Reply(result)
			},
		},
	)

	// Lists all votes stored in the database
	a.bot.Command("votes-history <chainName> <startDate> <endDateOptional>", &slacker.CommandDefinition{
		Description: "lists history of all votes for a given chain",
		Examples:    []string{"list-votes cosmoshub-4 2023-01-26  2023-02-30"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			chainName := request.Param("chainName")
			startDate := request.Param("startDate")
			if len(startDate) < 1 {
				response.ReportError(fmt.Errorf("StartDate cannot be empty"))
				return
			}

			endDate := request.StringParam("endDateOptional", "")
			votes, err := a.db.GetVoteLogs(chainName, startDate, endDate)
			if err != nil {
				response.ReportError(err)
			} else {

				apiClient := botCtx.APIClient()
				event := botCtx.Event()

				var blocks []slack.Block
				for _, vote := range votes {
					t := time.Unix(vote.Date, 0)
					date := t.Format("2006-01-02")
					blocks = append(
						blocks,
						slack.NewSectionBlock(
							slack.NewTextBlockObject(
								"mrkdwn",
								fmt.Sprintf("*%s* ---- *%s* ---- Proposal *%s* ---- *%s*", date, vote.ChainID, vote.ProposalID, vote.VoteOption), false, false),
							nil, nil,
						),
					)
				}

				attachment := []slack.Block{
					slack.NewHeaderBlock(slack.NewTextBlockObject("plain_text", "Date ---- Address ---- ProposalID ---- Vote", false, false)),
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
					blocks = append(
						blocks,
						slack.NewSectionBlock(
							slack.NewTextBlockObject(
								"mrkdwn",
								fmt.Sprintf("*%s* ---- *%s* ---- *%s*", key.ChainName, key.KeyName, key.KeyAddress), false, false),
							nil, nil,
						),
					)
				}

				attachment := []slack.Block{
					slack.NewHeaderBlock(slack.NewTextBlockObject("plain_text", "Network ---- Key name---- Address", false, false)),
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

	a.bot.Command("list-proposals", &slacker.CommandDefinition{
		Description: "lists all proposals",
		Examples:    []string{"list-proposals"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			a.data.GetProposals(a.db)
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
				return
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
