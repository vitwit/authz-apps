package main

import (
	"log"
	"sync"
	"time"

	"lens-bot/lens-bot-1/config"
	"lens-bot/lens-bot-1/targets"
)

func main() {
	cfg, err := config.ReadConfigFromFile()
	if err != nil {
		log.Fatal(err)
	}

	m := targets.InitTargets()
	runner := targets.NewRunner()

	log.Printf("targets initialized")
	var wg sync.WaitGroup
	for _, tg := range m.List {
		wg.Add(1)
		go func(target targets.Target) {
			scrapeRate, err := time.ParseDuration(target.ScraperRate)
			if err != nil {
				log.Fatal(err)
			}
			for {
				runner.Run(target.Func, cfg)
				time.Sleep(scrapeRate)
			}
		}(tg)
	}
	wg.Wait()
}
