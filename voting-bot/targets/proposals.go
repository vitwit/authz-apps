package targets

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/vitwit/authz-apps/voting-bot/database"
	"github.com/vitwit/authz-apps/voting-bot/types"
	"github.com/vitwit/authz-apps/voting-bot/utils"

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
	cdc := ctx.Codec()
	for _, val := range validators {

		chainInfo, err := ctx.ChainRegistry().GetChain(ctx.Context(), val.ChainName)
		if err != nil {
			return err
		}

		grpcEndpoint, err := chainInfo.GetActiveGRPCEndpoint(ctx.Context())
		if err != nil {
			return err
		}

		creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: false})
		conn, err := grpc.Dial(grpcEndpoint, grpc.WithTransportCredentials(creds))
		if err != nil {
			log.Printf("Failed to connect to %s: %v", grpcEndpoint, err)
			return err
		}
		defer conn.Close()

		ctx1, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		var missedProposals []MissedProposal
		if utils.GovV1Support[val.ChainName]["govv1_enabled"] {
			client := govv1types.NewQueryClient(conn)
			resp, err := client.Proposals(ctx1, &govv1types.QueryProposalsRequest{
				ProposalStatus: govv1types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
			})
			if err != nil {
				log.Printf("Failed to get proposals %s: %v", grpcEndpoint, err)
				return err
			}

			for _, proposal := range resp.Proposals {
				if err := ctx.Database().AddLog(val.ChainName, proposal.Metadata, fmt.Sprint(proposal.Id), ""); err != nil {
					fmt.Printf("failed to store vote logs: %v", err)
				}
				validatorVote, err := getValidatorVoteV1(ctx, client, proposal.Id, val.Address, val.ChainName)
				if err != nil {
					return err
				}

				if validatorVote == "" {
					missedProposals = append(missedProposals, MissedProposal{
						accAddr:       val.Address,
						pTitle:        proposal.Metadata,
						pID:           fmt.Sprint(proposal.Id),
						votingEndTime: proposal.VotingEndTime.String(),
					})
				} else {
					if err := ctx.Database().UpdateVoteLog(val.ChainName, fmt.Sprint(proposal.Id), validatorVote); err != nil {
						fmt.Printf("failed to update vote log: %v", err)
					}
				}
			}

		} else {
			client := govv1beta1types.NewQueryClient(conn)
			resp, err := client.Proposals(ctx1, &govv1beta1types.QueryProposalsRequest{
				ProposalStatus: govv1beta1types.StatusVotingPeriod,
			})
			if err != nil {
				log.Printf("Failed to get proposals %s: %v", grpcEndpoint, err)
				return err
			}

			for _, proposal := range resp.Proposals {

				var content govv1beta1types.Content
				if err = cdc.UnpackAny(proposal.Content, &content); err != nil {
					log.Printf("Failed to unpack proposal any field %v", err)
					return err
				}

				if err := ctx.Database().AddLog(val.ChainName, content.GetTitle(), fmt.Sprint(proposal.ProposalId), ""); err != nil {
					fmt.Printf("failed to store vote logs: %v", err)
				}
				validatorVote, err := getValidatorVoteV1beta1(ctx, client, proposal.ProposalId, val.Address, val.ChainName)
				if err != nil {
					return err
				}

				if validatorVote == "" {
					missedProposals = append(missedProposals, MissedProposal{
						accAddr:       val.Address,
						pTitle:        content.GetTitle(),
						pID:           fmt.Sprint(proposal.ProposalId),
						votingEndTime: proposal.VotingEndTime.String(),
					})
				} else {
					if err := ctx.Database().UpdateVoteLog(val.ChainName, fmt.Sprint(proposal.ProposalId), validatorVote); err != nil {
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
func getValidatorVoteV1(ctx types.Context, client govv1types.QueryClient, proposalID uint64, valAddr, chainName string) (string, error) {
	accAddrString, err := convertValAddrToAccAddr(ctx, valAddr, chainName)
	if err != nil {
		return "", err
	}

	resp, err := client.Vote(ctx.Context(), &govv1types.QueryVoteRequest{
		ProposalId: proposalID,
		Voter:      accAddrString,
	})
	if err != nil {
		log.Printf("Error while getting vote response: %v", err)
		return "", err
	}

	validatorVoted := ""
	for _, value := range resp.Vote.Options {
		validatorVoted = value.Option.String()
	}

	return validatorVoted, nil
}

// getValidatorVoteV1beta1 to check validator voted for the proposal or not.
func getValidatorVoteV1beta1(ctx types.Context, client govv1beta1types.QueryClient, proposalID uint64, valAddr, chainName string) (string, error) {
	accAddrString, err := convertValAddrToAccAddr(ctx, valAddr, chainName)
	if err != nil {
		return "", err
	}

	resp, err := client.Vote(ctx.Context(), &govv1beta1types.QueryVoteRequest{
		ProposalId: proposalID,
		Voter:      accAddrString,
	})
	if err != nil {
		log.Printf("Error while getting vote response: %v", err)
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
	// api := ctx.Slacker().APIClient()
	// var blocks []slack.Block

	// for _, p := range proposals {
	// 	endTime, _ := time.Parse(time.RFC3339, p.votingEndTime)
	// 	daysLeft := int(time.Until(endTime).Hours() / 24)
	// 	if daysLeft == 0 {
	// 		daysLeft = 1
	// 	}
	// 	blocks = append(blocks, slack.NewSectionBlock(
	// 		slack.NewTextBlockObject(
	// 			"mrkdwn",
	// 			fmt.Sprintf("*%s*", p.pTitle),
	// 			false, false,
	// 		),
	// 		nil, nil))
	// 	var fields []*slack.TextBlockObject
	// 	mintscanName := chainName
	// 	if newName, ok := mint.RegisrtyNameToMintscanName[chainName]; ok {
	// 		mintscanName = newName
	// 	}
	// 	fields = append(fields, slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Proposal Id*\n *<https://mintscan.io/%s/proposals/%s| %s >* ", mintscanName, p.pID, p.pID), false, false))
	// 	fields = append(fields, slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Voting ends in* \n %d days ", daysLeft), false, false))
	// 	fields = append(fields, slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Validator* \n%s", p.accAddr), false, false))
	// 	blocks = append(blocks, slack.NewSectionBlock(nil, fields, nil, slack.SectionBlockOptionBlockID("")))
	// }
	// attachment := []slack.Block{
	// 	slack.NewHeaderBlock(slack.NewTextBlockObject("plain_text", fmt.Sprintf(" %s ", chainName), false, false)),
	// }
	// attachment = append(attachment, blocks...)

	// _, _, err := api.PostMessage(
	// 	ctx.Config().Slack.ChannelID,
	// 	slack.MsgOptionBlocks(attachment...),
	// )
	// if err != nil {
	// 	return err
	// }

	return nil
}
