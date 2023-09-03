package types

import (
	"github.com/vitwit/authz-apps/voting-bot/config"
)

type (
	// QueryParams to map the query params of an url
	QueryParams map[string]string

	// HTTPOptions of a target
	HTTPOptions struct {
		Endpoint    string
		QueryParams QueryParams
		Body        []byte
		Method      string
	}

	// PingResp struct
	PingResp struct {
		StatusCode int
		Body       []byte
	}

	// Target is a structure which holds all the parameters of a target
	// this could be used to write endpoints for each functionality
	Target struct {
		ExecutionType string
		HTTPOptions   HTTPOptions
		Name          string
		Func          func(cfg *config.Config)
		ScraperRate   string
	}

	// Targets list of all the targets
	Targets struct {
		List []Target
	}

	// Balance struct holds the parameters of balance of grantee
	Balance struct {
		Balance struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		}
	}
)
