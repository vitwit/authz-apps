package targets

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/likhita-809/lens-bot/alerting"
	"github.com/likhita-809/lens-bot/config"
	"github.com/likhita-809/lens-bot/database"
	"github.com/robfig/cron"
)

// Cron wraps all required parameters to create cron jobs
type Cron struct {
	db  *database.Sqlitedb
	cfg *config.Config
	bot *alerting.Slackbot
}

// NewCron sets necessary config and clients to begin jobs
func NewCron(db *database.Sqlitedb, config *config.Config, bot *alerting.Slackbot) *Cron {
	return &Cron{
		db:  db,
		cfg: config,
		bot: bot,
	}
}

// Start starts to create cron jobs which fetches chosen asset list information and
// store them in database every hour and every 24 hours.
func (c *Cron) Start() error {
	log.Println("Starting cron jobs...")

	cron := cron.New()

	d := Data{
		db:  c.db,
		cfg: c.cfg,
	}

	// Every 15 minute
	cron.AddFunc("1 * * * *", func() {
		d.GetProposals(c.db)
		// log.Println("successfully saved price information list")
	})

	go cron.Start()

	return nil
}

// // NewRunner returns targetRunner
// func NewRunner() *targetRunner {
// 	return &targetRunner{}
// }

// // Run to run the request
// func (m targetRunner) Run(function func(cfg *config.Config)) {
// 	function(m.cfg)
// }

// func (m targetRunner) InitTargets() *Targets {
// 	return &Targets{List: []Target{
// 		{
// 			ExecutionType: "http",
// 			Name:          "Proposals",
// 			HTTPOptions: HTTPOptions{
// 				Method: http.MethodGet,
// 			},
// 			Func:        m.data.GetProposals,
// 			ScraperRate: "2h",
// 		},
// 		{
// 			Name: "Slack cmds",
// 			HTTPOptions: HTTPOptions{
// 				Method: http.MethodGet,
// 			},
// 			Func:        m.bot.Initializecommands,
// 			ScraperRate: "3s",
// 		},
// 	}}
// }

func addQueryParameters(req *http.Request, queryParams QueryParams) {
	q := req.URL.Query()
	for key, value := range queryParams {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()
}

// newHTTPRequest to make a new http request
func newHTTPRequest(ops HTTPOptions) (*http.Request, error) {
	// make new request
	req, err := http.NewRequest(ops.Method, ops.Endpoint, bytes.NewBuffer(ops.Body))
	if err != nil {
		return nil, err
	}

	// Add any query parameters to the URL.
	if len(ops.QueryParams) != 0 {
		addQueryParameters(req, ops.QueryParams)
	}

	return req, nil
}

func makeResponse(res *http.Response) (*PingResp, error) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return &PingResp{}, err
	}

	response := &PingResp{
		StatusCode: res.StatusCode,
		Body:       body,
	}
	_ = res.Body.Close()
	return response, nil
}

// HitHTTPTarget to hit the target and get response
func HitHTTPTarget(ops HTTPOptions) (*PingResp, error) {
	req, err := newHTTPRequest(ops)
	if err != nil {
		return nil, err
	}

	httpcli := http.Client{Timeout: time.Duration(30 * time.Second)}
	resp, err := httpcli.Do(req)
	if err != nil {
		return nil, err
	}

	res, err := makeResponse(resp)
	if err != nil {
		return nil, err
	}

	return res, nil
}
