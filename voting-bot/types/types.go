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

	// ProposalsLegacy struct holds result of array of legacy proposals
	ProposalsLegacy struct {
		Proposals  []LegacyProposal `json:"proposals"`
		Pagination struct {
			NextKey string `json:"next_key"`
			Total   string `json:"total"`
		} `json:"pagination"`
	}

	// Proposals struct holds result of array of proposals
	V1Proposals struct {
		Proposals []Proposal `json:"proposals"`
	}

	Proposal struct {
		ID            string      `json:"id"`
		Messages      []Message   `json:"messages"`
		Status        string      `json:"status"`
		Metadata      interface{} `json:"metadata"`
		VotingEndTime string      `json:"voting_end_time"`
	}

	Message struct {
		Type    string  `json:"@type"`
		Title   string  `json:"title"`
		Content Content `json:"content"`
	}

	Content struct {
		Type        string `json:"@type"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	Metadata struct {
		Title             string   `json:"title"`
		Authors           []string `json:"authors"`
		Summary           string   `json:"summary"`
		Details           string   `json:"details"`
		ProposalForumURL  string   `json:"proposal_forum_url"`
		VoteOptionContext string   `json:"vote_option_context"`
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
	LegacyProposal struct {
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
