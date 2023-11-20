package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shomali11/slacker"

	"github.com/vitwit/authz-apps/voting-bot/client"
	"github.com/vitwit/authz-apps/voting-bot/config"
	"github.com/vitwit/authz-apps/voting-bot/database"
	"github.com/vitwit/authz-apps/voting-bot/handler"
	"github.com/vitwit/authz-apps/voting-bot/jobs"
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

	// Initialize the router
	router := mux.NewRouter()

	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	// Wrap the router with the CORS middleware

	// Define REST API endpoints
	router.HandleFunc("/rewards", handler.GetRewardsHandler(db)).Methods("OPTIONS", "GET")
	router.HandleFunc("/votes/{chainName}", handler.RetrieveProposalsHandler(db)).Methods("OPTIONS", "GET")
	router.HandleFunc("/votes", handler.RetrieveProposalsForAllNetworksHandler(db)).Methods("OPTIONS", "GET")

	// CORS middleware
	http.Handle("/", corsMiddleware(router))

	// Start the server
	go func() {
		logger.Info().Msg("REST server started on 8080 port")
		log.Error().Err(http.ListenAndServe(":8080", router))
	}()

	cfg, err := config.ReadConfigFromFile()
	if err != nil {
		logger.Error().Err(err)
	}

	bot := slacker.NewClient(cfg.Slack.BotToken, cfg.Slack.AppToken)
	logger.Info().Msg("bot connected")
	ctx := types.NewContext(logger, db, cfg, bot)

	fmt.Println(bot.BotCommands())

	cron := jobs.NewCron(ctx)
	cron.Start()

	client.InitializeBotcommands(ctx)

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
