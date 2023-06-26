package targets

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/slack-go/slack"

	"github.com/likhita-809/lens-bot/database"
	"github.com/likhita-809/lens-bot/types"
	"github.com/likhita-809/lens-bot/utils"
	mint "github.com/likhita-809/lens-bot/voting"
)

type (
	MissedProposal struct {
		accAddr       string
		pTitle        string
		pID           string
		votingEndTime string
	}
)

// Gets proposals from the Registered chains and validators
func GetProposals(ctx types.Context) {
	var networksMap map[string]bool
	var networks []string
	vals, err := ctx.Database().GetValidators()
	if err != nil {
		log.Fatalf("Error while getting validators: %v", err)
	}

	for _, val := range vals {
		if !networksMap[val.ChainName] {
			networks = append(networks, val.ChainName)
		}
	}
	err = alertOnProposals(ctx, networks, vals)
	if err != nil {
		log.Printf("Error while alerting on proposals: %s", err)
	}
}

// Alerts on Active Proposals
func alertOnProposals(ctx types.Context, networks []string, validators []database.Validator) error {
	for _, val := range validators {
		endpoint, err := GetValidEndpointForChain(val.ChainName)
		if err != nil {
			log.Printf("Error in getting valid LCD endpoints for %s chain", val.ChainName)

			return err
		}
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

		var missedProposals []MissedProposal

		ctx.Logger().Info().Msgf("pending proposals = ", len(p.Proposals), "  chain-name = ", val.ChainName)
		for _, proposal := range p.Proposals {
			validatorVote, err := getValidatorVote(ctx, endpoint, proposal.ProposalID, val.Address, val.ChainName)
			if err != nil {
				return err
			}

			if validatorVote == "" {
				missedProposals = append(missedProposals, MissedProposal{
					accAddr:       val.Address,
					pTitle:        proposal.Content.Title,
					pID:           proposal.ProposalID,
					votingEndTime: proposal.VotingEndTime,
				})
			}

		}

		if len(missedProposals) > 0 {
			err = sendVotingPeriodProposalAlerts(ctx, val.ChainName, missedProposals)
			if err != nil {
				log.Printf("error on sending voting period proposals alert: %v", err)
				return err
			}
		}
	}
	return nil
}

// getValidatorVote to check validator voted for the proposal or not.
func getValidatorVote(ctx types.Context, endpoint, proposalID, valAddr, chainName string) (string, error) {
	chainInfo, err := ctx.ChainRegistry().GetChain(context.Background(), chainName)
	if err != nil {
		return "", err
	}

	done := utils.SetBech32Prefixes(chainInfo)
	addr, err := utils.ValAddressFromBech32(valAddr)
	if err != nil {
		done()
		return "", err
	}

	accAddr, err := utils.AccAddressFromHexUnsafe(hex.EncodeToString(addr.Bytes()))
	if err != nil {
		return "", err
	}

	accAddrString := accAddr.String()
	done()

	fmt.Println("chainID = ", chainName, "  Account Addr = ", accAddrString)
	ops := HTTPOptions{
		Endpoint: endpoint + "/cosmos/gov/v1beta1/proposals/" + proposalID + "/votes/" + accAddrString,
		Method:   http.MethodGet,
	}
	resp, err := HitHTTPTarget(ops)
	if err != nil {
		log.Printf("Error while getting http response: %v", err)
		return "", err
	}

	var v Vote
	err = json.Unmarshal(resp.Body, &v)
	if err != nil {
		log.Printf("Error while unmarshalling the proposal votes: %v", err)
		return "", err
	}

	validatorVoted := ""
	for _, value := range v.Vote.Options {
		validatorVoted = value.Option
	}

	return validatorVoted, nil
}

// sendVotingPeriodProposalAlerts which send alerts of voting period proposals
func sendVotingPeriodProposalAlerts(ctx types.Context, chainName string, proposals []MissedProposal) error {
	api := ctx.Slacker().APIClient()
	var blocks []slack.Block

	for _, p := range proposals {
		endTime, _ := time.Parse(time.RFC3339, p.votingEndTime)
		daysLeft := int(time.Until(endTime).Hours() / 24)
		if daysLeft == 0 {
			daysLeft = 1
		}
		blocks = append(blocks, slack.NewSectionBlock(
			slack.NewTextBlockObject(
				"mrkdwn",
				fmt.Sprintf("*%s*", p.pTitle),
				false, false,
			),
			nil, nil))
		var fields []*slack.TextBlockObject
		mintscanName := chainName
		if newName, ok := mint.RegisrtyNameToMintscanName[chainName]; ok {
			mintscanName = newName
		}
		fields = append(fields, slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Proposal Id*\n *<https://mintscan.io/%s/proposals/%s| %s >* ", mintscanName, p.pID, p.pID), false, false))
		fields = append(fields, slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Voting ends in* \n %d days ", daysLeft), false, false))
		fields = append(fields, slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Validator* \n%s", p.accAddr), false, false))
		blocks = append(blocks, slack.NewSectionBlock(nil, fields, nil, slack.SectionBlockOptionBlockID("")))
	}
	attachment := []slack.Block{
		slack.NewHeaderBlock(slack.NewTextBlockObject("plain_text", fmt.Sprintf(" %s ", chainName), false, false)),
	}
	attachment = append(attachment, blocks...)

	_, _, err := api.PostMessage(
		ctx.Config().Slack.ChannelID,
		slack.MsgOptionBlocks(attachment...),
	)
	if err != nil {
		return err
	}

	return nil
}
