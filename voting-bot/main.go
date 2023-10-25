package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/shomali11/slacker"
	"github.com/vitwit/authz-apps/voting-bot/client"
	"github.com/vitwit/authz-apps/voting-bot/config"
	"github.com/vitwit/authz-apps/voting-bot/database"
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

	// // Define REST API endpoints
	router.HandleFunc("/rewards", getRewardsHandler(db)).Methods("GET")

	// // Start the server

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

// TODO: seperate file
func getRewardsHandler(db *database.Sqlitedb) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		params := r.URL.Query()
		chainId := params.Get("id")
		date := params.Get("date")

		rewards, err := db.GetRewards(chainId, date)
		if err != nil {
			http.Error(w, fmt.Errorf("error while getting rewards: %w", err).Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(rewards)
		if err != nil {
			http.Error(w, fmt.Errorf("error while encoding rewards: %w", err).Error(), http.StatusInternalServerError)
			return
		}
	}
}
