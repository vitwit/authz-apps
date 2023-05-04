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

	// m := targets.InitTargets()
	// runner := targets.NewRunner()

	// log.Printf("targets initialized")
	// var wg sync.WaitGroup
	// for _, tg := range m.List {
	// 	wg.Add(1)
	// 	go func(target targets.Target) {
	// 		scrapeRate, err := time.ParseDuration(target.ScraperRate)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		for {
	// 			runner.Run(target.Func, cfg)
	// 			time.Sleep(scrapeRate)
	// 		}
	// 	}(tg)
	// }
	// wg.Wait()

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
