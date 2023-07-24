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

	// Proposals struct holds result of array of proposals
	Proposals struct {
		Proposals  []Proposal `json:"proposals"`
		Pagination struct {
			NextKey string `json:"next_key"`
			Total   string `json:"total"`
		} `json:"pagination"`
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
	Proposal struct {
		ProposalID string `json:"proposal_id"`
		Content    struct {
			Type        string `json:"@type"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Changes     []struct {
				Subspace string `json:"subspace"`
				Key      string `json:"key"`
				Value    string `json:"value"`
			} `json:"changes"`
		} `json:"content,omitempty"`
		Status           string `json:"status"`
		FinalTallyResult struct {
			Yes        string `json:"yes"`
			Abstain    string `json:"abstain"`
			No         string `json:"no"`
			NoWithVeto string `json:"no_with_veto"`
		} `json:"final_tally_result"`
		SubmitTime     string `json:"submit_time"`
		DepositEndTime string `json:"deposit_end_time"`
		TotalDeposit   []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"total_deposit"`
		VotingStartTime string `json:"voting_start_time"`
		VotingEndTime   string `json:"voting_end_time"`
	}
)
