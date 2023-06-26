package main

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/shomali11/slacker"
	"github.com/vitwit/authz-apps/voting-bot/client"
	"github.com/vitwit/authz-apps/voting-bot/config"
	"github.com/vitwit/authz-apps/voting-bot/database"
	"github.com/vitwit/authz-apps/voting-bot/targets"
	"github.com/vitwit/authz-apps/voting-bot/types"
)

func main() {
	db, err := database.NewDatabase()
	if err != nil {
		panic(err)
	}
	db.InitializeTables()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := log.Logger
	defer func() {
		if r := recover(); r != nil {
			logger.Debug().Msgf("recovered from panic:: %v", r)
		}
	}()

	cfg, err := config.ReadConfigFromFile()
	if err != nil {
		logger.Error().Err(err)
	}

	bot := slacker.NewClient(cfg.Slack.BotToken, cfg.Slack.AppToken)
	logger.Info().Msg("bot connected")
	ctx := types.NewContext(logger, db, cfg, bot)

	fmt.Println(bot.BotCommands())

	cron := targets.NewCron(ctx)
	cron.Start()

	client.InitializeBotcommands(ctx)

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
