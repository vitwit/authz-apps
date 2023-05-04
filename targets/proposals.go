package targets

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/likhita-809/lens-bot/alerting"
	"github.com/likhita-809/lens-bot/config"
	"github.com/likhita-809/lens-bot/database"
)

type Data struct {
	db *database.Sqlitedb
}

// Gets Proposals,valid LCD endpoints and alerts on proposals.

func (a *Data) GetProposals(cfg *config.Config) {

	validEndpoints, err := GetValidLCDEndpoints()
	if err != nil {
		log.Printf("Error in getting proposals endpoint : %v\n", err)
	}
	for _, proposalsEndpoint := range validEndpoints {
		err = a.AlertOnProposals(proposalsEndpoint, cfg)
		if err != nil {
			log.Printf("Error in sending proposals alert : %v\n", err)
		}
	}
}

// Alerts on Active Proposals
func (a *Data) AlertOnProposals(endpoint string, cfg *config.Config) error {
	ops := HTTPOptions{
		Endpoint:    endpoint + "/cosmos/gov/v1beta1/proposals",
		Method:      http.MethodGet,
		QueryParams: QueryParams{"proposal_status": "2"},
	}
	resp, err := HitHTTPTarget(ops)
	if err != nil {
		log.Printf("Error while getting http response: %v", err)
		return err
	}

	var p Proposals
	err = json.Unmarshal(resp.Body, &p)
	if err != nil {
		log.Printf("Error while unmarshalling the proposals: %v", err)
		return err
	}

	for _, proposal := range p.Proposals {
		valAddrs, err := a.db.GetValidatorAddress()
		if err != nil {
			return err
		}
		for _, valAddr := range valAddrs {
			validatorVote := GetValidatorVote(endpoint, proposal.ProposalID, valAddr)
			if validatorVote == "" {
				err := SendVotingPeriodProposalAlerts(valAddr, proposal.ProposalID, proposal.VotingEndTime, cfg)
				if err != nil {
					log.Printf("error on sending voting period proposals alert: %v", err)
				}
			} // else {
			//StoreValidatorVote(endpoint, proposal.ProposalID, valAddr)
			//}

		}
	}

	return nil
}

// // StoreValidatorVote to store the validator vote information.
// func (a *Data) StoreValidatorVote(endpoint, proposalID, valAddr string) {
// 	ops := HTTPOptions{
// 		Endpoint: endpoint + "/cosmos/gov/v1beta1/proposals/" + proposalID + "/votes/" + valAddr,
// 		Method:   http.MethodGet,
// 	}
// 	resp, err := HitHTTPTarget(ops)
// 	if err != nil {
// 		log.Printf("Error while getting http response: %v", err)
// 	}

// 	var v Vote
// 	err = json.Unmarshal(resp.Body, &v)
// 	if err != nil {
// 		log.Printf("Error while unmarshalling the proposal votes: %v", err)
// 	}
// 	var voteOption string
// 	for _, v := range v.Vote.Options {
// 		voteOption = v.Option
// 	}
// 	a.db.VotesDataInsert(v.Vote.ProposalID, v.Vote.Voter, voteOption)
// }

// GetValidatorVote to check validator voted for the proposal or not.

// Checks whether validator has voted or not
func GetValidatorVote(endpoint, proposalID, valAddr string) string {
	ops := HTTPOptions{
		Endpoint: endpoint + "/cosmos/gov/v1beta1/proposals/" + proposalID + "/votes/" + valAddr,
		Method:   http.MethodGet,
	}
	resp, err := HitHTTPTarget(ops)
	if err != nil {
		log.Printf("Error while getting http response: %v", err)
	}
	var v Vote
	err = json.Unmarshal(resp.Body, &v)
	if err != nil {
		log.Printf("Error while unmarshalling the proposal votes: %v", err)
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
	log.Println("timeDiff...", timeDiff.Hours())

	err := alerting.NewSlackAlerter().Send(fmt.Sprintf("you have not voted on proposal %s with address %s", proposalID, accountAddress), cfg.Slack.BotToken, cfg.Slack.ChannelID)
	if err != nil {
		return err
	} else {
		log.Println("Sent alert of voting period proposals")
	}
	return nil
}
