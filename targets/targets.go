package targets

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/likhita-809/lens-bot/types"
	"github.com/robfig/cron"
	"github.com/rs/zerolog"
)

// Cron wraps all required parameters to create cron jobs
type Cron struct {
	ctx    types.Context
	logger *zerolog.Logger
}

// NewCron sets necessary config and clients to begin cron jobs
func NewCron(ctx types.Context) *Cron {
	return &Cron{
		ctx:    ctx,
		logger: ctx.Logger(),
	}
}

// Start starts to create cron jobs which sends alerts on proposal alerts which have not been voted
func (c *Cron) Start() error {
	c.logger.Info().Msg("Starting cron jobs...")

	cron := cron.New()

	// Everday at 8AM and 8PM
	err := cron.AddFunc("0 0 8,20 * * *", func() {
		GetProposals(c.ctx)
		GetLowBalAccs(c.ctx)
	})
	if err != nil {
		log.Println("Error while adding Proposals and Low balance accounts alerting cron jobs:", err)
		return err
	}
	err = cron.AddFunc("@every 1h", func() {
		SyncAuthzStatus(c.ctx)
	})
	if err != nil {
		log.Println("Error while adding Key Authorization syncing cron job:", err)
		return err
	}
	go cron.Start()

	return nil
}

// Adds the Query parameters
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

// Creates response
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
