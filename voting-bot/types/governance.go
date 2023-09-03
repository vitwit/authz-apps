package types

type LegacyProposals struct {
	Proposals []LegacyProposal `json:"proposals"`
}

type LegacyProposal struct {
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

type Proposals struct {
	Proposals []Proposal `json:"proposals"`
}

type Proposal struct {
	ID            string      `json:"id"`
	Messages      []Message   `json:"messages"`
	Status        string      `json:"status"`
	Metadata      interface{} `json:"metadata"`
	VotingEndTime string      `json:"voting_end_time"`
}

type Message struct {
	Type    string  `json:"@type"`
	Title   string  `json:"title"`
	Content Content `json:"content"`
}

type Content struct {
	Type        string `json:"@type"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type Metadata struct {
	Title             string   `json:"title"`
	Authors           []string `json:"authors"`
	Summary           string   `json:"summary"`
	Details           string   `json:"details"`
	ProposalForumURL  string   `json:"proposal_forum_url"`
	VoteOptionContext string   `json:"vote_option_context"`
}

type VoteResponse struct {
	Vote Vote `json:"vote"`
}

type LegacyVoteResponse struct {
	Vote LegacyVote `json:"vote"`
}

type Vote struct {
	ProposalID string       `json:"proposal_id"`
	Voter      string       `json:"voter"`
	Options    []VoteOption `json:"options"`
	Metadata   string       `json:"metadata"`
}

type LegacyVote struct {
	ProposalID string       `json:"proposal_id"`
	Voter      string       `json:"voter"`
	Option     string       `json:"option"`
	Options    []VoteOption `json:"options"`
}

type VoteOption struct {
	Option string `json:"option"`
	Weight string `json:"weight"`
}
