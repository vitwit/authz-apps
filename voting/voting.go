package voting

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/likhita-809/lens-bot/database"
	lensclient "github.com/strangelove-ventures/lens/client"
	registry "github.com/strangelove-ventures/lens/client/chain_registry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

type Vote struct {
	db *database.Sqlitedb
}

func (v *Vote) GetChainName(chainID string, cr registry.ChainRegistry) (string, error) {
	chains, err := cr.ListChains(context.Background())
	if err != nil {
		return "", err
	}

	for _, chainName := range chains {
		chainInfo, _ := cr.GetChain(context.Background(), chainName)
		if chainInfo.ChainID == chainID {
			return chainName, nil
		}
	}
	log.Fatalf("chain name not found")
	return "", nil
}

func (v *Vote) ExecVote(chainID, pID, valAddr, vote, fromKey, metadata, memo, gas, fees string) error {
	// Fetch chain info from chain registry
	cr := registry.DefaultChainRegistry(zap.New(zapcore.NewNopCore()))

	chainName, err := v.GetChainName(chainID, cr)
	if err != nil {
		fmt.Errorf("chain name not found")
	}
	chainInfo, err := cr.GetChain(context.Background(), chainName)
	if err != nil {
		fmt.Errorf("Failed to get chain info. Err: %v \n", err)
	}

	//	Use Chain info to select random endpoint
	rpc, err := chainInfo.GetRandomRPCEndpoint(context.Background())
	if err != nil {
		fmt.Errorf("failed to get random RPC endpoint on chain %s. Err: %v", chainInfo.ChainID, err)
	}

	chainConfig := lensclient.ChainClientConfig{
		Key:            fromKey,
		ChainID:        chainInfo.ChainID,
		RPCAddr:        rpc,
		AccountPrefix:  chainInfo.Bech32Prefix,
		KeyringBackend: "test",
		GasPrices:      gas,
		Debug:          true,
		Timeout:        "20s",
		OutputFormat:   "json",
		SignModeStr:    "direct",
		Modules:        lensclient.ModuleBasics,
	}

	curDir, err := os.Getwd()
	if err != nil {
		fmt.Errorf("error while getting current directory: %v", err)
	}
	// Create client object to pull chain info
	chainClient, err := lensclient.NewChainClient(zap.L(), &chainConfig, curDir, os.Stdin, os.Stdout)
	if err != nil {
		fmt.Errorf("failed to build new chain client for %s. Err: %v", chainInfo.ChainID, err)
	}

	keyAddr, err := v.db.GetKeyAddress(fromKey)
	if err != nil {
		fmt.Errorf("error while getting address of %s key", fromKey)
	}

	proposalID, err := strconv.ParseUint(pID, 10, 64)
	if err != nil {
		fmt.Errorf("unable to convert string to uint64. Err: %v", err)
	}
	voteOption, err := v.stringToVoteOption(vote)
	if err != nil {
		fmt.Errorf("unable to convert vote option string to sdk vote option. Err: %v", err)
	}

	var msgAny *cdctypes.Any

	validV1Endpoint, err := v.GetValidV1Endpoint(chainInfo)
	if err != nil {
		fmt.Errorf("error while getting valid gov v1 endpoint: %v", err)
	}
	if validV1Endpoint {
		msgVote := v1.MsgVote{
			ProposalId: proposalID,
			Voter:      valAddr,
			Option:     voteOption,
			Metadata:   metadata,
		}
		msgAny, err = cdctypes.NewAnyWithValue(&msgVote)
		if err != nil {
			fmt.Errorf("error on converting msg to Any: %v", err)
		}

	} else {
		msgVote := v1beta1.MsgVote{
			ProposalId: proposalID,
			Voter:      valAddr,
			Option:     v1beta1.VoteOption(voteOption),
		}
		msgAny, err = cdctypes.NewAnyWithValue(&msgVote)
		if err != nil {
			fmt.Errorf("error on converting msg to Any: %v", err)
		}
	}

	req := &authz.MsgExec{
		Grantee: keyAddr,
		Msgs:    []*cdctypes.Any{msgAny},
	}

	// Send msg and get response
	res, err := chainClient.SendMsg(context.Background(), req, memo)
	if err != nil {
		if res != nil {
			fmt.Errorf("failed to vote on proposal: code(%d) msg(%s)", res.Code, res.Logs)
		}
		fmt.Errorf("Failed to vote.Err: %v", err)
	}
	return nil
}

func (v *Vote) stringToVoteOption(str string) (v1.VoteOption, error) {
	str = strings.ToLower(str)
	switch str {
	case "yes":
		return v1.OptionYes, nil
	case "no":
		return v1.OptionNo, nil
	case "abstain":
		return v1.OptionAbstain, nil
	case "no_with_veto":
		return v1.OptionNoWithVeto, nil
	default:
		return v1.VoteOption(0), fmt.Errorf("invalid vote option: %s", str)

	}
}

func (v *Vote) GetValidV1Endpoint(chainInfo registry.ChainInfo) (bool, error) {
	var out []string
	for _, endpoint := range chainInfo.Apis.Rest {
		u, err := url.Parse(endpoint.Address)
		if err != nil {
			return false, err
		}
		var port string
		if u.Port() == "" {
			switch u.Scheme {
			case "https":
				port = "443"
			case "http":
				port = "80"
			default:
				return false, fmt.Errorf("invalid or unsupported url scheme: %v", u.Scheme)

			}
		} else {
			port = u.Port()
		}

		out = append(out, fmt.Sprintf("%s://%s:%s%s", u.Scheme, u.Hostname(), port, u.Path))
	}
	validEndpoint := false
	for _, endpoint := range out {
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			return false, err
		}
		client := &http.Client{
			Timeout: 10 * time.Second, // set the time so that we should get response within that time
		}

		resp, err := client.Do(req)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()
		validEndpoint = true
		if validEndpoint {
			break
		}
	}
	return validEndpoint, nil
}
