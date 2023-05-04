package targets

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/slack-go/slack"

	"github.com/likhita-809/lens-bot/config"
	"github.com/likhita-809/lens-bot/database"
)

type Data struct {
	db  *database.Sqlitedb
	cfg *config.Config
	//bot *alerting.Slackbot
}

// Gets Proposals,valid LCD endpoints and alerts on proposals.

func (a *Data) GetProposals(db *database.Sqlitedb) {

	var networksMap map[string]bool
	var networks []string
	vals, err := db.GetValidators()
	if err != nil {
		panic(err)
	}

	for _, val := range vals {
		if !networksMap[val.ChainName] {
			networks = append(networks, val.ChainName)
		}
	}
	a.AlertOnProposals(networks)

}

type MissedProposal struct {
	accAdd     string
	pID        string
	votEndTime string
}

// Alerts on Active Proposals
func (a *Data) AlertOnProposals(networks []string) error {
	validators, err := a.db.GetValidators()
	if err != nil {
		return err
	}
	for _, val := range validators {
		validEndpoints, err := GetValidLCDEndpoints(val.ChainName)
		if err != nil {
			log.Printf("Error in getting%s validator's proposals endpoint : %v\n", val.Address, err)
		}
		for _, endpoint := range validEndpoints {
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
			log.Println("alert1")

			var missedProposals []MissedProposal

			for _, proposal := range p.Proposals {
				log.Println("alert2")

				validatorVote := a.GetValidatorVote(endpoint, proposal.ProposalID, val.Address)
				if validatorVote == "" {
					log.Println("alert3")
					missedProposals = append(missedProposals, MissedProposal{
						accAdd:     val.Address,
						pID:        proposal.ProposalID,
						votEndTime: proposal.VotingEndTime,
					})
				}

			}

			// err = a.SendVotingPeriodProposalAlerts(val.Address, proposal.ProposalID, proposal.VotingEndTime)
			err = a.SendVotingPeriodProposalAlerts(missedProposals)
			if err != nil {
				log.Printf("error on sending voting period proposals alert: %v", err)
			}
		}
	}
	return nil
}

// GetValidatorVote to check validator voted for the proposal or not.

func (a *Data) GetValidatorVote(endpoint, proposalID, valAddr string) string {

	ValAddr, _ := sdk.ValAddressFromBech32(valAddr)
	accAddr, _ := sdk.AccAddressFromHexUnsafe(hex.EncodeToString(ValAddr.Bytes()))
	ops := HTTPOptions{
		Endpoint: endpoint + "/cosmos/gov/v1beta1/proposals/" + proposalID + "/votes/" + accAddr.String(),
		Method:   http.MethodGet,
	}
	log.Println(ops.Endpoint)
	resp, err := HitHTTPTarget(ops)
	if err != nil {
		log.Printf("Error while getting http response: %v", err)
	}
	var v Vote
	err = json.Unmarshal(resp.Body, &v)
	if err != nil {
		log.Printf("Error while unmarshalling the proposal votes: %v", err)
	}
	log.Println(v)

	validatorVoted := ""
	for _, value := range v.Vote.Options {
		validatorVoted = value.Option
	}

	return validatorVoted
}

// SendVotingPeriodProposalAlerts which send alerts of voting period proposals
func (a *Data) SendVotingPeriodProposalAlerts(proposals []MissedProposal) error {
	//err := alerting.NewSlackAlerter().Send(fmt.Sprintf("you have not voted on proposal %s with address %s", proposalID, accountAddress), a.cfg.Slack.BotToken, a.cfg.Slack.ChannelID)
	// var botCtx slacker.BotContext
	api := slack.New(a.cfg.Slack.BotToken)
	var blocks []slack.Block

	for _, p := range proposals {
		now := time.Now().UTC()
		endTime, _ := time.Parse(time.RFC3339, p.votEndTime)
		timeDiff := now.Sub(endTime)
		log.Println("timeDiff...", timeDiff.Hours())

		if timeDiff.Hours() <= 24 {
			blocks = append(blocks, slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*%s* ---*%s*--- *%s*", p.accAdd, p.pID, p.votEndTime), false, false),
				nil, nil))
		} else {
			log.Println("Sent alert of voting period proposals")
		}
	}
	attachment := []slack.Block{
		slack.NewHeaderBlock(slack.NewTextBlockObject("plain_text", "Account Address ---Proposal ID--- Voting End Time", false, false)),
	}
	attachment = append(attachment, blocks...)

	_, _, err := api.PostMessage(
		a.cfg.Slack.ChannelID,
		slack.MsgOptionBlocks(attachment...),
	)
	if err != nil {
		return err
	}

	return nil
}
