package targets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/likhita-809/lens-bot/alerting"
	"github.com/likhita-809/lens-bot/config"
	"github.com/likhita-809/lens-bot/sqldata"
)

func GetProposals(cfg *config.Config) {
	validEndpoints, err := GetValidLCDEndpoints()
	if err != nil {
		fmt.Printf("Error in getting proposals endpoint : %v\n", err)
	}
	for _, proposalsEndpoint := range validEndpoints {
		err = AlertOnProposals(proposalsEndpoint, cfg)
		if err != nil {
			fmt.Printf("Error in sending proposals alert : %v\n", err)
		}
	}
}

func AlertOnProposals(endpoint string, cfg *config.Config) error {
	ops := HTTPOptions{
		Endpoint:    endpoint + "/cosmos/gov/v1beta1/proposals",
		Method:      http.MethodGet,
		QueryParams: QueryParams{"proposal_status": "2"},
	}
	resp, err := HitHTTPTarget(ops)
	if err != nil {
		fmt.Printf("Error while getting http response: %v", err)
		return err
	}

	var p Proposals
	err = json.Unmarshal(resp.Body, &p)
	if err != nil {
		fmt.Printf("Error while unmarshalling the proposals: %v", err)
		return err
	}

	for _, proposal := range p.Proposals {
		valAddrs, err := sqldata.GetAllValAddrs()
		if err != nil {
			return err
		}
		for _, valAddr := range valAddrs {
			validatorVote := GetValidatorVote(endpoint, proposal.ProposalID, valAddr)
			if validatorVote == "" {
				err := SendVotingPeriodProposalAlerts(valAddr, proposal.ProposalID, proposal.VotingEndTime, cfg)
				if err != nil {
					return fmt.Errorf("error on sending voting period proposals alert: %v", err)
				}
			} else {
				StoreValidatorVote(endpoint, proposal.ProposalID, valAddr)
			}

		}
	}

	return nil
}

// StoreValidatorVote to store the validator vote information.
func StoreValidatorVote(endpoint, proposalID, valAddr string) {
	ops := HTTPOptions{
		Endpoint: endpoint + "/cosmos/gov/v1beta1/proposals/" + proposalID + "/votes/" + valAddr,
		Method:   http.MethodGet,
	}
	resp, err := HitHTTPTarget(ops)
	if err != nil {
		fmt.Printf("Error while getting http response: %v", err)
	}

	var v Vote
	err = json.Unmarshal(resp.Body, &v)
	if err != nil {
		fmt.Printf("Error while unmarshalling the proposal votes: %v", err)
	}
	var voteOption string
	for _, v := range v.Vote.Options {
		voteOption = v.Option
	}
	sqldata.VotesDataInsert(v.Vote.ProposalID, v.Vote.Voter, voteOption)
}

// GetValidatorVote to check validator voted for the proposal or not.
func GetValidatorVote(endpoint, proposalID, valAddr string) string {
	ops := HTTPOptions{
		Endpoint: endpoint + "/cosmos/gov/v1beta1/proposals/" + proposalID + "/votes/" + valAddr,
		Method:   http.MethodGet,
	}
	resp, err := HitHTTPTarget(ops)
	if err != nil {
		fmt.Printf("Error while getting http response: %v", err)
	}
	var v Vote
	err = json.Unmarshal(resp.Body, &v)
	if err != nil {
		fmt.Printf("Error while unmarshalling the proposal votes: %v", err)
	}

	validatorVoted := ""
	for _, value := range v.Vote.Options {
		validatorVoted = value.Option
	}

	return validatorVoted
}

// SendVotingPeriodProposalAlerts which send alerts of voting period proposals
func SendVotingPeriodProposalAlerts(accountAddress, proposalID, votingEndTime string, cfg *config.Config) error {
	now := time.Now().UTC()
	endTime, _ := time.Parse(time.RFC3339, votingEndTime)
	timeDiff := now.Sub(endTime)
	fmt.Println("timeDiff...", timeDiff.Hours())

	err := alerting.NewSlackAlerter().Send(fmt.Sprintf("you have not voted on proposal %s with address %s", proposalID, accountAddress), cfg.Slack.BotToken, cfg.Slack.ChannelID)
	if err != nil {
		return err
	} else {
		fmt.Println("Sent alert of voting period proposals")
	}
	return nil
}
