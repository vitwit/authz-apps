package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/likhita-809/lens-bot/config"
	"github.com/likhita-809/lens-bot/targets"
)

func main() {
	cfg, err := config.ReadConfigFromFile()
	if err != nil {
		fmt.Errorf("%s", err)
	}

	m := targets.InitTargets()
	runner := targets.NewRunner()

	fmt.Printf("targets initialized")
	var wg sync.WaitGroup
	for _, tg := range m.List {
		wg.Add(1)
		go func(target targets.Target) {
			scrapeRate, err := time.ParseDuration(target.ScraperRate)
			if err != nil {
				fmt.Errorf("%s", err)
			}
			for {
				runner.Run(target.Func, cfg)
				time.Sleep(scrapeRate)
			}
		}(tg)
	}
	wg.Wait()
}
