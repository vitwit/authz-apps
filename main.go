package main

import (
	"log"
	"sync"
	"time"

	"github.com/likhita-809/lens-bot/config"
	"github.com/likhita-809/lens-bot/targets"
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
