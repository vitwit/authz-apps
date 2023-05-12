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
	registry "github.com/strangelove-ventures/lens/client/chain_registry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/likhita-809/lens-bot/config"
	"github.com/likhita-809/lens-bot/database"
)

type (
	Data struct {
		db  *database.Sqlitedb
		cfg *config.Config
	}

	MissedProposal struct {
		accAddr       string
		pID           string
		votingEndTime string
	}
)

// Gets proposals from the Registered chains and validators
func (a *Data) GetProposals(db *database.Sqlitedb) {
	var networksMap map[string]bool
	var networks []string
	vals, err := db.GetValidators()
	if err != nil {
		log.Fatalf("Error while getting validators: %v", err)
	}

	for _, val := range vals {
		if !networksMap[val.ChainName] {
			networks = append(networks, val.ChainName)
		}
	}
	err = a.AlertOnProposals(networks)
	if err != nil {
		log.Printf("Error while alerting on proposals: %s", err)
	}
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
			log.Printf("Error in getting valid LCD endpoints for %s chain", val.ChainName)

			return err
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

			var missedProposals []MissedProposal

			fmt.Println("pending proposals = ", len(p.Proposals), "  chain-name = ", val.ChainName)
			for _, proposal := range p.Proposals {
				validatorVote, err := a.GetValidatorVote(endpoint, proposal.ProposalID, val.Address, val.ChainName)
				if err != nil {
					return err
				}

				if validatorVote == "" {
					missedProposals = append(missedProposals, MissedProposal{
						accAddr:       val.Address,
						pID:           proposal.ProposalID,
						votingEndTime: proposal.VotingEndTime,
					})
				}

			}

			if len(missedProposals) > 0 {
				err = a.SendVotingPeriodProposalAlerts(val.ChainName, missedProposals)
				if err != nil {
					log.Printf("error on sending voting period proposals alert: %v", err)
					return err
				}
			}
		}
	}
	return nil
}

// GetValidatorVote to check validator voted for the proposal or not.
func (a *Data) GetValidatorVote(endpoint, proposalID, valAddr, chainName string) (string, error) {

	cr := registry.DefaultChainRegistry(zap.New(zapcore.NewNopCore()))
	chainInfo, err := cr.GetChain(context.Background(), chainName)
	if err != nil {
		return "", err
	}

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(chainInfo.Bech32Prefix, chainInfo.Bech32Prefix+"pub")
	config.SetBech32PrefixForValidator(chainInfo.Bech32Prefix+"valoper", chainInfo.Bech32Prefix+"valoperpub")

	addr, err := sdk.ValAddressFromBech32(valAddr)
	if err != nil {
		return "", err
	}

	accAddr, err := sdk.AccAddressFromHexUnsafe(hex.EncodeToString(addr.Bytes()))
	if err != nil {
		return "", err
	}

	ops := HTTPOptions{
		Endpoint: endpoint + "/cosmos/gov/v1beta1/proposals/" + proposalID + "/votes/" + accAddr.String(),
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

// SendVotingPeriodProposalAlerts which send alerts of voting period proposals
func (a *Data) SendVotingPeriodProposalAlerts(chainName string, proposals []MissedProposal) error {
	api := slack.New(a.cfg.Slack.BotToken)
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
				fmt.Sprintf("*%s* has not voted on proposal *%s* . Voting ends in *%d days.* ", p.accAddr, p.pID, daysLeft),
				false, false,
			),
			nil, nil))
	}
	attachment := []slack.Block{
		slack.NewHeaderBlock(slack.NewTextBlockObject("plain_text", fmt.Sprintf(" %s ", chainName), false, false)),
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
