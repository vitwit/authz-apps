package jobs

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/slack-go/slack"
	"github.com/vitwit/authz-apps/voting-bot/database"
	"github.com/vitwit/authz-apps/voting-bot/endpoints"
	"github.com/vitwit/authz-apps/voting-bot/types"
	"github.com/vitwit/authz-apps/voting-bot/utils"
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
	vals, err := ctx.Database().GetValidators()
	if err != nil {
		log.Fatalf("Error while getting validators: %v", err)
	}

	networksMap := make(map[string]bool)
	var networks []string
	for _, val := range vals {
		if !networksMap[val.ChainName] {
			networks = append(networks, val.ChainName)
			networksMap[val.ChainName] = true
		}
	}

	alertOnProposals(ctx, networks, vals)
}

// Alerts on Active Proposals
func alertOnProposals(ctx types.Context, networks []string, validators []database.Validator) {
	for _, val := range validators {
		endpoint, err := endpoints.GetValidEndpointForChain(val.ChainName)
		if err != nil {
			log.Printf("no active REST endpoint for %s", val.ChainName)
			sendPlainAlert(ctx, fmt.Sprintf("No active %s endpoint available for %s", "REST", val.ChainName))
			continue
		}

		var missedProposals []MissedProposal
		if utils.GovV1Support[val.ChainName]["govv1_enabled"] {

			proposals, err := GetActiveProposals(ctx, true, endpoint)
			if err != nil {
				log.Printf("failed to get active proposal for %s", val.ChainName)
				sendPlainAlert(ctx, fmt.Sprintf("failed to get active proposals for chain %s: %v", val.ChainName, err))
				continue
			}

			for _, proposal := range proposals {
				if err := ctx.Database().AddLog(val.ChainName, proposal.Title, proposal.ProposalID, ""); err != nil {
					fmt.Printf("failed to store vote logs: %v", err)
				}

				vote, err := GetValidatorVoteOption(ctx, true, val.ChainName, endpoint, proposal.ProposalID, val.Address)
				if err != nil {
					log.Printf("failed to get validator vote for %s", val.ChainName)
					sendPlainAlert(ctx, fmt.Sprintf("failed to get validator vote fo %s: %v", val.ChainName, err))
					continue
				}

				if vote == "" {
					missedProposals = append(missedProposals, MissedProposal{
						accAddr:       val.Address,
						pTitle:        proposal.Title,
						pID:           proposal.ProposalID,
						votingEndTime: proposal.VotingEndTime,
					})
				} else {
					if err := ctx.Database().UpdateVoteLog(val.ChainName, proposal.ProposalID, vote); err != nil {
						fmt.Printf("failed to update vote log: %v", err)
					}
				}
			}

		} else {
			proposals, err := GetActiveProposals(ctx, false, endpoint)
			if err != nil {
				log.Printf("failed to get active proposal for %s", val.ChainName)
				sendPlainAlert(ctx, fmt.Sprintf("failed to get active proposals for chain %s: %v", val.ChainName, err))
				continue
			}

			for _, proposal := range proposals {
				if err := ctx.Database().AddLog(val.ChainName, proposal.Title, proposal.ProposalID, ""); err != nil {
					fmt.Printf("failed to store vote logs: %v", err)
				}

				vote, err := GetValidatorVoteOption(ctx, false, val.ChainName, endpoint, proposal.ProposalID, val.Address)
				if err != nil {
					log.Printf("failed to get validator vote for %s", val.ChainName)
					sendPlainAlert(ctx, fmt.Sprintf("failed to get validator vote fo %s: %v", val.ChainName, err))
					continue
				}

				if vote == "" {
					missedProposals = append(missedProposals, MissedProposal{
						accAddr:       val.Address,
						pTitle:        proposal.Title,
						pID:           proposal.ProposalID,
						votingEndTime: proposal.VotingEndTime,
					})
				} else {
					if err := ctx.Database().UpdateVoteLog(val.ChainName, proposal.ProposalID, vote); err != nil {
						fmt.Printf("failed to update vote log: %v", err)
					}
				}
			}

		}

		log.Println("Network name = ", val.ChainName)
		log.Println("Missed proposals = ", len(missedProposals))
		if len(missedProposals) > 0 {
			err = sendVotingPeriodProposalAlerts(ctx, val.ChainName, missedProposals)
			if err != nil {
				log.Printf("error on sending voting period proposals alert: %v", err)
			}
		}
	}
}

func sendPlainAlert(ctx types.Context, msg string) error {
	api := ctx.Slacker().APIClient()

	attachment := []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", msg, false, false),
		),
	}

	_, _, err := api.PostMessage(
		ctx.Config().Slack.ChannelID,
		slack.MsgOptionBlocks(attachment...),
	)
	if err != nil {
		return err
	}

	return nil
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
		if newName, ok := utils.RegisrtyNameToMintscanName[chainName]; ok {
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

type ActiveProposalResult struct {
	ProposalID    string
	Title         string
	VotingEndTime string
}

func GetActiveProposals(ctx types.Context, isV1 bool, restEndpoint string) ([]ActiveProposalResult, error) {
	if isV1 {
		resp, err := endpoints.HitHTTPTarget(types.HTTPOptions{
			Endpoint:    restEndpoint + "/cosmos/gov/v1/proposals",
			Method:      http.MethodGet,
			QueryParams: types.QueryParams{"proposal_status": "2"},
		})
		if err != nil {
			return nil, err
		}

		var proposals types.Proposals
		if err := json.Unmarshal(resp.Body, &proposals); err != nil {
			return nil, err
		}

		var result []ActiveProposalResult
		for _, proposal := range proposals.Proposals {
			title, err := getTitleFromProposal(proposal)
			if err != nil {
				title = "Unknown title"
			}

			result = append(result, ActiveProposalResult{
				ProposalID:    proposal.ID,
				Title:         title,
				VotingEndTime: proposal.VotingEndTime,
			})
		}

		return result, nil
	} else {
		resp, err := endpoints.HitHTTPTarget(types.HTTPOptions{
			Endpoint:    restEndpoint + "/cosmos/gov/v1beta1/proposals",
			Method:      http.MethodGet,
			QueryParams: types.QueryParams{"proposal_status": "2"},
		})
		if err != nil {
			return nil, err
		}

		var proposals types.LegacyProposals
		if err := json.Unmarshal(resp.Body, &proposals); err != nil {
			return nil, err
		}

		var result []ActiveProposalResult
		for _, proposal := range proposals.Proposals {
			result = append(result, ActiveProposalResult{
				ProposalID:    proposal.ProposalID,
				Title:         proposal.Content.Title,
				VotingEndTime: proposal.VotingEndTime,
			})
		}

		return result, nil
	}
}

func GetValidatorVoteOption(ctx types.Context, isV1 bool, chainName, restEndpoint, proposalID, validatorAddress string) (string, error) {
	accAddrString, err := ConvertValAddrToAccAddr(ctx, validatorAddress, chainName)
	if err != nil {
		return "", err
	}

	if isV1 {
		resp, err := endpoints.HitHTTPTarget(types.HTTPOptions{
			Endpoint: restEndpoint + "/cosmos/gov/v1/proposals/" + proposalID + "/votes/" + accAddrString,
			Method:   http.MethodGet,
		})
		if err != nil {
			return "", err
		}

		var vote types.VoteResponse
		if err := json.Unmarshal(resp.Body, &vote); err != nil {
			return "", err
		}

		if len(vote.Vote.Options) == 0 {
			return "", nil
		}

		if len(vote.Vote.Options) == 1 {
			return vote.Vote.Options[0].Option, nil
		}

		var voteResult string
		for _, option := range vote.Vote.Options {
			voteResult += fmt.Sprintf("%s-%s", option.Option, option.Weight)
		}

		return voteResult, nil

	} else {
		resp, err := endpoints.HitHTTPTarget(types.HTTPOptions{
			Endpoint: restEndpoint + "/cosmos/gov/v1beta1/proposals/" + proposalID + "/votes/" + accAddrString,
			Method:   http.MethodGet,
		})
		if err != nil {
			return "", err
		}
		var vote types.LegacyVoteResponse
		if err := json.Unmarshal(resp.Body, &vote); err != nil {
			return "", err
		}

		if len(vote.Vote.Options) == 0 {
			return "", nil
		}

		if len(vote.Vote.Options) == 1 {
			return vote.Vote.Options[0].Option, nil
		}

		var voteResult string
		for _, option := range vote.Vote.Options {
			voteResult += fmt.Sprintf("%s-%s", option.Option, option.Weight)
		}

		return voteResult, nil
	}
}

func getTitleFromProposal(proposal types.Proposal) (string, error) {
	if len(proposal.Messages) == 0 {
		return getMetadataTitle(proposal.Metadata)
	}

	message := proposal.Messages[0]

	if message.Type == "/cosmos.gov.v1.MsgExecLegacyContent" {
		return message.Content.Title, nil
	}

	if message.Title != "" {
		return message.Title, nil
	}

	return getMetadataTitle(proposal.Metadata)
}

func getMetadataTitle(metadata interface{}) (string, error) {
	switch metadata.(type) {
	case string:
		return getTitleFromString(metadata.(string))
	case types.Metadata:
		return metadata.(types.Metadata).Title, nil
	}

	return "", nil
}

func getTitleFromString(metadataStr string) (string, error) {
	if metadataStr == "" {
		return "", nil
	}

	if strings.Contains(metadataStr, "ipfs://") {
		ipfsTitle, err := fetchMetadataFromIPFS(metadataStr[7:])
		return ipfsTitle, err
	}

	// Otherwise, unmarshal and get the title
	var meta types.Metadata
	err := json.Unmarshal([]byte(metadataStr), &meta)
	if err != nil {
		return "", err
	}

	return meta.Title, nil
}

type GovMetadata struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

func fetchMetadataFromIPFS(ipfsLink string) (string, error) {
	resp, err := http.Get("https://ipfs.io/ipfs/" + ipfsLink)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var meta types.Metadata
	err = json.Unmarshal(body, &meta)
	if err != nil {
		return "", err
	}

	return meta.Title, nil
}

func ConvertValAddrToAccAddr(ctx types.Context, valAddr, chainName string) (string, error) {
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
	return accAddrString, nil
}
