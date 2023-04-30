package targets

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"lens-bot/lens-bot-1/alerting"
	"lens-bot/lens-bot-1/config"
)

type targetRunner struct{}

// NewRunner returns targetRunner
func NewRunner() *targetRunner {
	return &targetRunner{}
}

// Run to run the request
func (m targetRunner) Run(function func(cfg *config.Config), cfg *config.Config) {
	function(cfg)
}

func InitTargets() *Targets {
	return &Targets{List: []Target{
		{
			ExecutionType: "http",
			Name:          "Proposals",
			HTTPOptions: HTTPOptions{
				Method: http.MethodGet,
			},
			Func:        GetProposals,
			ScraperRate: "2h",
		},
		{
			Name: "Slack cmds",
			HTTPOptions: HTTPOptions{
				Method: http.MethodGet,
			},
			Func:        alerting.RegisterSlack,
			ScraperRate: "3s",
		},
	}}
}

func addQueryParameters(req *http.Request, queryParams QueryParams) {
	params := url.Values{}
	for key, value := range queryParams {
		params.Add(key, value)
	}
	req.URL.RawQuery = params.Encode()
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
