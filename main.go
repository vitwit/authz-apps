package main

import (
	"log"
	"sync"

	"github.com/likhita-809/lens-bot/alerting"
	"github.com/likhita-809/lens-bot/config"
	"github.com/likhita-809/lens-bot/database"
	"github.com/likhita-809/lens-bot/keyshandler"
	"github.com/likhita-809/lens-bot/targets"
	"github.com/likhita-809/lens-bot/voting"
)

func main() {
	db, err := database.NewDatabase()
	if err != nil {
		panic(err)
	}
	db.InitializeTables()

	cfg, err := config.ReadConfigFromFile()
	if err != nil {
		log.Printf("%s", err)
	}
	alerter := alerting.NewBotClient(cfg, db, &keyshandler.Keys{}, &voting.Vote{})
	cron := targets.NewCron(db, cfg, alerter)
	cron.Start()
	alerter.Initializecommands()

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
