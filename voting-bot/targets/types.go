package targets

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

	// V1Proposals struct holds result of array of gov v1 proposals
	V1Proposals struct {
		Proposals []struct {
			ID       string `json:"id"`
			Messages []struct {
				Type    string `json:"@type"`
				Content struct {
					Type        string `json:"@type"`
					Title       string `json:"title"`
					Description string `json:"description"`
				} `json:"content,omitempty"`
			}
			Status        string `json:"status"`
			VotingEndTime string `json:"voting_end_time"`
			Metadata      string `json:"metadata"`
		} `json:"proposals"`
	}

	// Proposals struct holds result of array of gov v1beta1 proposals
	Proposals struct {
		Proposals []struct {
			ProposalID string `json:"proposal_id"`
			Content    struct {
				Type        string `json:"@type"`
				Title       string `json:"title"`
				Description string `json:"description"`
			} `json:"content,omitempty"`
			Status        string `json:"status"`
			VotingEndTime string `json:"voting_end_time"`
		} `json:"proposals"`
	}

	// Balance struct holds the parameters of balance of grantee
	Balance struct {
		Balance struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		}
	}

	// Vote struct holds the parameters of vote of voter.
	Vote struct {
		Vote struct {
			ProposalID string `json:"proposal_id"`
			Voter      string `json:"voter"`
			Option     string `json:"option"`
			Options    []struct {
				Option string `json:"option"`
				Weight string `json:"weight"`
			}
		}
	}
)
