package main

import (
	"log"
	"sync"

	"github.com/likhita-809/lens-bot/alerting"
	"github.com/likhita-809/lens-bot/config"
	"github.com/likhita-809/lens-bot/database"
	"github.com/likhita-809/lens-bot/targets"
)

func main() {
	db, err := database.NewDatabase()
	if err != nil {
		panic(err)
	}
	db.InitializeTables()

	cfg, err := config.ReadConfigFromFile()
	if err != nil {
		log.Fatal(err)
	}
	alerter := alerting.NewBotClient(cfg, db)
	cron := targets.NewCron(db, cfg, alerter)
	cron.Start()
	alerter.Initializecommands()

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
