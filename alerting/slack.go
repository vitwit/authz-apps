package alerting

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/likhita-809/lens-bot/config"
	keyshandler "github.com/likhita-809/lens-bot/keysHandler"
	"github.com/likhita-809/lens-bot/sqldata"
	"github.com/likhita-809/lens-bot/voting"

	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
)

func printCommandEvents(analyticsChannel <-chan *slacker.CommandEvent) {
	for event := range analyticsChannel {
		fmt.Println("Command Events")
		fmt.Println(event.Timestamp)
		fmt.Println(event.Command)
		fmt.Println(event.Parameters)
		fmt.Println(event.Event)
	}
}

// Send allows bot to send a slack alert to the configured channelID
func RegisterSlack(config *config.Config) {
	// Create a new client to slack by giving token
	// Set debug to true while developing

	bot := slacker.NewClient(config.Slack.BotToken, config.Slack.AppToken)

	// show logs of command events
	go printCommandEvents(bot.CommandEvents())

	bot.Command("register <chain_id> <validator_address>", &slacker.CommandDefinition{
		Description: "register",
		Examples:    []string{"/register cosmoshub cosmos1a..."},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			chain_id := request.Param("chain_id")
			validator_address := request.Param("validator_address")
			sqldata.ChainDataInsert(chain_id, validator_address)
			r := fmt.Sprintf("your respose has been recorded %s", validator_address)
			response.Reply(r)
		},
	})
	bot.Command("create-key <chain_name> <key_name_optional>", &slacker.CommandDefinition{
		Description: "create a new account with key name",
		Examples:    []string{"create-key my_key"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			key_name := request.StringParam("key_name_optional", "default")
			chain_name := request.Param("chain_name")
			err := keyshandler.CreateKeys(chain_name, key_name)
			if err != nil {
				response.Reply(err.Error())
			} else {
				NewSlackAlerter().Send(fmt.Sprintf("Successfully created your key with name %s", key_name), config.Slack.BotToken, config.Slack.ChannelID)
			}
		},
	})
	bot.Command(
		"vote <chain_id> <proposal_id> <validator_address> <vote_option> <from_key> <metadata_optional> <memo_optional> <gas_units_optional> <fees_optional>",
		&slacker.CommandDefinition{
			Description: "vote",
			Examples:    []string{"/vote cosmoshub 123 YES memodata 300000 0.25uatom "},
			Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
				chainID := request.Param("chain_id")
				pID := request.Param("proposal_id")
				valAddr := request.Param("validator_address")
				voteOption := request.Param("vote_option")
				fromKey := request.Param("from_key")
				metadata := request.StringParam("metadata_optional", "")
				memo := request.StringParam("memo_optional", "")
				gas := request.StringParam("gas_units_optional", "")
				fees := request.StringParam("fees_optional", "")
				err := voting.ExecVote(chainID, pID, valAddr, voteOption, fromKey, metadata, memo, gas, fees)
				if err != nil {
					fmt.Printf("error on executing vote: %v", err)
				}
				a := fmt.Sprintf("%v", err.Error())
				response.Reply(a)
			},
		},
	)
	bot.Command("list-keys", &slacker.CommandDefinition{
		Description: "lists all keys",
		Examples:    []string{"list-all"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			r, err := sqldata.ListKeys()
			if err != nil {
				response.ReportError(err)
			} else {
				data := fmt.Sprintf("%v", r)

				apiClient := botCtx.APIClient()
				event := botCtx.Event()

				attachment := slack.Attachment{
					Title: "List of all keys",
					Text:  data,
				}
				if event.ChannelID != "" {
					_, _, err := apiClient.PostMessage(event.ChannelID, slack.MsgOptionAttachments(attachment))
					if err != nil {
						response.ReportError(err)
					}
				}
			}
		},
	})
	bot.Command("list-all", &slacker.CommandDefinition{
		Description: "lists all chains with associated validator addresses",
		Examples:    []string{"list-all"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			r, err := sqldata.ChainDataList()
			if err != nil {
				response.ReportError(err)
			} else {
				data := fmt.Sprintf("%v", r)

				apiClient := botCtx.APIClient()
				event := botCtx.Event()

				attachment := slack.Attachment{
					Title: "List of all chains and their validators",
					Text:  data,
				}
				if event.ChannelID != "" {
					_, _, err := apiClient.PostMessage(event.ChannelID, slack.MsgOptionAttachments(attachment))
					if err != nil {
						response.ReportError(err)
					}
				}
			}
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := bot.Listen(ctx)
	if err != nil {
		log.Fatal(err)
	}
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
		// Fields are Optional extra data!
		Fields: []slack.AttachmentField{
			{
				Title: "Date",
				Value: time.Now().String(),
			},
		},
	}

	// PostMessage will send the message away.
	// First parameter is just the channelID, makes no sense to accept it
	_, timestamp, err := client.PostMessage(
		channelID,
		// uncomment the item below to add a extra Header to the message, try it out :)
		// slack.MsgOptionText("New message from bot", false),
		// slack.MsgOptionText(msgText, false),
		slack.MsgOptionAttachments(attachment),
	)
	if err != nil {
		return err
	}
	fmt.Printf("Message sent at %s", timestamp)
	return nil
}
