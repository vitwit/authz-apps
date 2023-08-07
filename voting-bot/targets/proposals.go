package targets

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"time"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/slack-go/slack"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/vitwit/authz-apps/voting-bot/database"
	"github.com/vitwit/authz-apps/voting-bot/endpoints"
	"github.com/vitwit/authz-apps/voting-bot/types"
	"github.com/vitwit/authz-apps/voting-bot/utils"
	mint "github.com/vitwit/authz-apps/voting-bot/voting"
)

type (
	MissedProposal struct {
		accAddr string
		//pTitle        string
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
		endpoint, err := endpoints.GetValidEndpointForChain(val.ChainName)
		if err != nil {
			log.Printf("Error in getting valid LCD endpoints for %s chain", val.ChainName)

			return err
		}
		creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: false})
		ctx1, _ := context.WithTimeout(context.Background(), 5*time.Second)
		grpcConn, err := grpc.DialContext(ctx1, endpoint, grpc.WithTransportCredentials(creds))
		if err != nil {
			return err
		}
		defer grpcConn.Close()
		queryclient := govtypes.NewQueryClient(grpcConn)
		req := &govtypes.QueryProposalsRequest{
			ProposalStatus: 2,
		}
		resp, err := queryclient.Proposals(
			context.Background(),
			req,
		)
		if err != nil {
			return err
		}

		var missedProposals []MissedProposal

		ctx.Logger().Info().Msgf("pending proposals = ", len(resp.Proposals), "  chain-name = ", val.ChainName)
		for _, proposal := range resp.Proposals {

			validatorVote, err := getValidatorVote(ctx, endpoint, string(proposal.Content.Value), val.Address, val.ChainName)
			if err != nil {
				return err
			}

			if validatorVote == "" {
				missedProposals = append(missedProposals, MissedProposal{
					accAddr: val.Address,
					//pTitle:        proposal.Content.Title,
					pID:           strconv.FormatUint(uint64(proposal.ProposalId), 10),
					votingEndTime: time.Time.String(proposal.VotingEndTime),
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

// getValidatorVote to check validator voted for the proposal or not.
func getValidatorVote(ctx types.Context, endpoint, proposalID, valAddr, chainName string) (string, error) {
	accAddrString, err := convertValAddrToAccAddr(ctx, valAddr, chainName)
	if err != nil {
		return "", err
	}

	fmt.Println("chainID = ", chainName, "  Account Addr = ", accAddrString)
	creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: false})
	ctx1, _ := context.WithTimeout(context.Background(), 5*time.Second)
	grpcConn, err := grpc.DialContext(ctx1, endpoint, grpc.WithTransportCredentials(creds))
	if err != nil {
		return "", err
	}
	defer grpcConn.Close()
	queryclient := govtypes.NewQueryClient(grpcConn)
	req := &govtypes.QueryVoteRequest{
		ProposalId: 801,
		Voter:      accAddrString,
	}
	resp, err := queryclient.Vote(
		context.Background(),
		req,
	)
	if err != nil {
		return "", err
	}

	validatorVoted := ""
	for _, value := range resp.Vote.Options {
		validatorVoted = string(value.Option)
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
				fmt.Sprintf("*title*"),
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
