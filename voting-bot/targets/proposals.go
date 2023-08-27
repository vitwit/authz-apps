package targets

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/slack-go/slack"
	"github.com/vitwit/authz-apps/voting-bot/database"
	"github.com/vitwit/authz-apps/voting-bot/endpoints"
	"github.com/vitwit/authz-apps/voting-bot/types"
	"github.com/vitwit/authz-apps/voting-bot/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
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

	err = alertOnProposals(ctx, networks, vals)
	if err != nil {
		log.Printf("Error while alerting on proposals: %s", err)
	}
}

// Alerts on Active Proposals
func alertOnProposals(ctx types.Context, networks []string, validators []database.Validator) error {
	for _, val := range validators {

		chainInfo, err := ctx.ChainRegistry().GetChain(ctx.Context(), val.ChainName)
		if err != nil {
			return err
		}

		endpoint, err := endpoints.GetValidEndpointForChain(val.ChainName)
		if err != nil {
			return err
		}

		grpcEndpoint, err := chainInfo.GetActiveGRPCEndpoint(ctx.Context())
		if err != nil {
			log.Printf("Error while getting grpc endpoint : %v", err)
			return err
		}

		creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: false})
		conn, err := grpc.Dial(grpcEndpoint, grpc.WithTransportCredentials(creds))
		if err != nil {
			log.Printf("Failed to connect to %s: %v", grpcEndpoint, err)
			return err
		}
		defer conn.Close()

		var missedProposals []MissedProposal
		if utils.GovV1Support[val.ChainName]["govv1_enabled"] {
			ops := types.HTTPOptions{
				Endpoint:    endpoint + "/cosmos/gov/v1/proposals",
				Method:      http.MethodGet,
				QueryParams: types.QueryParams{"proposal_status": "2"},
			}
			resp, err := endpoints.HitHTTPTarget(ops)
			if err != nil {
				log.Printf("Error while getting http response: %v", err)
				return err
			}

			var proposals types.V1Proposals
			err = json.Unmarshal(resp.Body, &proposals)
			if err != nil {
				log.Printf("Error while unmarshalling proposals: %v", err)
				return err
			}

			for _, proposal := range proposals.Proposals {
				client := govv1types.NewQueryClient(conn)

				title, err := getTitleFromProposal(proposal)
				if err != nil {
					log.Printf("failed to get proposal title: %v", err)
					return err
				}
				if err := ctx.Database().AddLog(val.ChainName, title, fmt.Sprint(proposal.ID), ""); err != nil {
					fmt.Printf("failed to store vote logs: %v", err)
				}
				validatorVote, err := getValidatorVoteV1(ctx, client, proposal.ID, val.Address, val.ChainName)
				if err != nil {
					return err
				}

				if validatorVote == "" {
					missedProposals = append(missedProposals, MissedProposal{
						accAddr:       val.Address,
						pTitle:        title,
						pID:           proposal.ID,
						votingEndTime: proposal.VotingEndTime,
					})
				} else {
					if err := ctx.Database().UpdateVoteLog(val.ChainName, proposal.ID, validatorVote); err != nil {
						fmt.Printf("failed to update vote log: %v", err)
					}
				}
			}

		} else {
			ops := types.HTTPOptions{
				Endpoint:    endpoint + "/cosmos/gov/v1beta1/proposals",
				Method:      http.MethodGet,
				QueryParams: types.QueryParams{"proposal_status": "2"},
			}
			resp, err := endpoints.HitHTTPTarget(ops)
			if err != nil {
				log.Printf("Error while getting http response: %v", err)
				return err
			}

			var proposals types.ProposalsLegacy
			err = json.Unmarshal(resp.Body, &proposals)
			if err != nil {
				log.Printf("Error while unmarshalling proposals: %v", err)
				return err
			}

			client := govv1beta1types.NewQueryClient(conn)
			for _, proposal := range proposals.Proposals {
				if err := ctx.Database().AddLog(val.ChainName, proposal.Content.Title, proposal.ProposalID, ""); err != nil {
					fmt.Printf("failed to store vote logs: %v", err)
				}
				validatorVote, err := getValidatorVoteV1beta1(ctx, client, proposal.ProposalID, val.Address, val.ChainName)
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
				} else {
					if err := ctx.Database().UpdateVoteLog(val.ChainName, proposal.ProposalID, validatorVote); err != nil {
						fmt.Printf("failed to update vote log: %v", err)
					}
				}
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

func convertValAddrToAccAddr(ctx types.Context, valAddr, chainName string) (string, error) {
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

// getValidatorVoteV1 to check validator voted for the proposal or not.
func getValidatorVoteV1(ctx types.Context, client govv1types.QueryClient, proposalID string, valAddr, chainName string) (string, error) {
	accAddrString, err := convertValAddrToAccAddr(ctx, valAddr, chainName)
	if err != nil {
		return "", err
	}

	pID, err := strconv.ParseUint(proposalID, 10, 64)
	if err != nil {
		return "", err
	}

	resp, err := client.Vote(ctx.Context(), &govv1types.QueryVoteRequest{
		ProposalId: pID,
		Voter:      accAddrString,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found for proposal") {
			return "", nil
		}
		log.Printf("Error while getting  v1 vote response: %v", err)
		return "", err
	}

	validatorVoted := ""
	for _, value := range resp.Vote.Options {
		validatorVoted = value.Option.String()
	}

	return validatorVoted, nil
}

// getValidatorVoteV1beta1 to check validator voted for the proposal or not.
func getValidatorVoteV1beta1(ctx types.Context, client govv1beta1types.QueryClient, proposalID string, valAddr, chainName string) (string, error) {
	accAddrString, err := convertValAddrToAccAddr(ctx, valAddr, chainName)
	if err != nil {
		return "", err
	}

	pID, err := strconv.ParseUint(proposalID, 10, 64)
	if err != nil {
		return "", err
	}

	resp, err := client.Vote(ctx.Context(), &govv1beta1types.QueryVoteRequest{
		ProposalId: pID,
		Voter:      accAddrString,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found for proposal") {
			return "", nil
		}
		log.Printf("Error while getting v1beta1 vote response: %v", err)
		return "", err
	}

	validatorVoted := ""
	for _, value := range resp.Vote.Options {
		validatorVoted = value.Option.String()
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

	var metadata map[string]interface{}
	err := json.Unmarshal([]byte(metadataStr), &metadata)
	if err != nil {
		return "", err
	}

	// Check if the metadata is an IPFS link
	if reflect.TypeOf(metadata["title"]).String() == "string" {
		ipfsLink := metadata["title"].(string)
		if len(ipfsLink) > 0 && ipfsLink[:8] == "ipfs://" {
			ipfsTitle, err := fetchMetadataFromIPFS(ipfsLink)
			return ipfsTitle, err
		}
	}

	// Otherwise, unmarshal and get the title
	var meta types.Metadata
	err = json.Unmarshal([]byte(metadataStr), &meta)
	if err != nil {
		return "", err
	}

	return meta.Title, nil
}

func fetchMetadataFromIPFS(ipfsLink string) (string, error) {
	resp, err := http.Get(ipfsLink)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
