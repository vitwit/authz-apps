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

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic:", r)
		}
	}()

	cfg, err := config.ReadConfigFromFile()
	if err != nil {
		log.Printf("%s", err)
	}

	keys := keyshandler.Keys{
		Db: db,
	}
	votes := voting.Vote{
		Db: db,
	}
	alerter := alerting.NewBotClient(cfg, db, &keys, &votes)

	cron := targets.NewCron(db, cfg, alerter)
	cron.Start()

	alerter.Initializecommands()

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
