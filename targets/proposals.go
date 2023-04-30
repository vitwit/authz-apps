package targets

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"lens-bot/lens-bot-1/alerting"
	"lens-bot/lens-bot-1/config"
	"lens-bot/lens-bot-1/sqldata"
)

func GetProposals(cfg *config.Config) {
	proposalsEndpoints, err := GetValidProposalsLCDEndpoints()
	// fmt.Printf("All valid proposalsEndpoints: %v\n", proposalsEndpoints)
	if err != nil {
		fmt.Printf("Error in getting proposals endpoint : %v\n", err)
	}
	for _, proposalsEndpoint := range proposalsEndpoints {
		err = AlertOnProposals(proposalsEndpoint, cfg)
		if err != nil {
			fmt.Printf("Error in sending proposals alert : %v\n", err)
		}
	}
}

func AlertOnProposals(endpoint string, cfg *config.Config) error {
	ops := HTTPOptions{
		Endpoint: endpoint + "/cosmos/gov/v1beta1/proposals",
		Method:   http.MethodGet,
	}

	resp, err := HitHTTPTarget(ops)
	if err != nil {
		log.Printf("Error in external rpc: %v", err)
		fmt.Printf("⛔⛔ Unreachable to EXTERNAL RPC :: %s and the ERROR is : %v\n\n", ops.Endpoint, err.Error())
		return err
	}

	var p Proposals
	err = json.Unmarshal(resp.Body, &p)
	if err != nil {
		log.Printf("Error: %v", err)
		return err
	}

	totalCount, _ := strconv.ParseFloat(p.Pagination.Total, 64)
	l := math.Ceil(totalCount / 100)
	nextKey := p.Pagination.NextKey

	for i := 1; i <= int(l); i++ {
		ops := HTTPOptions{
			Endpoint:    endpoint + "/cosmos/gov/v1beta1/proposals",
			Method:      http.MethodGet,
			QueryParams: QueryParams{"pagination.limit=": "100", "pagination.key": nextKey},
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

		fmt.Println(i, "==========START=============")
		fmt.Printf("endpoint: %s\n", endpoint)
		for _, proposal := range p.Proposals {
			if proposal.Status == "PROPOSAL_STATUS_VOTING_PERIOD" {
				valAddrs, err := sqldata.GetAllValAddrs()
				if err != nil {
					return err
				}
				for _, valAddr := range valAddrs {
					validatorVoted := GetValidatorVoted(endpoint, proposal.ProposalID, valAddr)
					SendVotingPeriodProposalAlerts(endpoint, valAddr, validatorVoted, proposal.ProposalID, proposal.VotingEndTime, cfg)

				}
			}
		}
		fmt.Println("i, ===========END=============")
	}
	return nil
}

// GetValidatorVoted to check validator voted for the proposal or not
func GetValidatorVoted(endpoint, proposalID, valAddr string) string {
	proposalURL := endpoint + "/cosmos/gov/v1beta1/proposals" + proposalID + "/votes"
	res, err := http.Get(proposalURL)
	if err != nil {
		log.Printf("Error in getting proposal votes: %v", err)
	}

	var voters ProposalVoters
	if res != nil {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Printf("Error while reading resp of proposal voters : %v ", err)
		}
		_ = json.Unmarshal(body, &voters)
	}

	validatorVoted := ""
	for _, value := range voters.Result {
		if value.Voter == valAddr {
			validatorVoted = value.Option
		}
	}
	return validatorVoted
}

// SendVotingPeriodProposalAlerts which send alerts of voting period proposals
func SendVotingPeriodProposalAlerts(endpoint, accountAddress, validatorVoted, proposalID, votingEndTime string, cfg *config.Config) {
	if (validatorVoted != "" && validatorVoted == "VOTE_OPTION_NO") || validatorVoted == "" {
		now := time.Now().UTC()
		votingEndTime, _ := time.Parse(time.RFC3339, votingEndTime)
		timeDiff := now.Sub(votingEndTime)
		log.Println("timeDiff...", timeDiff.Hours())

		if timeDiff.Hours() <= 24 {
			_ = alerting.NewSlackAlerter().Send(fmt.Sprintf("you have not voted on proposal = %s", proposalID), cfg.Slack.BotToken, cfg.Slack.ChannelID)
			log.Println("Sent alert of voting period proposals")
		}
	}
}
