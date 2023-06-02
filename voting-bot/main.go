package main

import (
	"log"
	"sync"

	"github.com/vitwit/authz-apps/voting-bot/client"
	"github.com/vitwit/authz-apps/voting-bot/config"
	"github.com/vitwit/authz-apps/voting-bot/database"
	"github.com/vitwit/authz-apps/voting-bot/keyring"
	"github.com/vitwit/authz-apps/voting-bot/targets"
	"github.com/vitwit/authz-apps/voting-bot/voting"
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

	keys := keyring.Keys{
		Db: db,
	}
	votes := voting.Vote{
		Db: db,
	}
	alerter := client.NewBotClient(cfg, db, &keys, &votes)

	cron := targets.NewCron(db, cfg, alerter)
	cron.Start()

	alerter.Initializecommands()

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
