package voting

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/likhita-809/lens-bot/endpoints"
	"github.com/likhita-809/lens-bot/types"
	"github.com/shomali11/slacker"
	lensclient "github.com/strangelove-ventures/lens/client"
	registry "github.com/strangelove-ventures/lens/client/chain_registry"
	"go.uber.org/zap"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

var RegisrtyNameToMintscanName = map[string]string{
	"cosmos":        "cosmos",
	"cosmoshub":     "cosmos",
	"osmosis":       "osmosis",
	"regen":         "regen",
	"akash":         "akash",
	"stride":        "stride",
	"juno":          "juno",
	"umee":          "umee",
	"omniflixhub":   "omniflix",
	"axelar":        "axelar",
	"bandchain":     "bandchain",
	"comdex":        "comdex",
	"desmos":        "desmos",
	"emoney":        "emoney",
	"evmos":         "evmos",
	"gravitybridge": "gravity-bridge",
	"tgrade":        "tgrade",
	"stargaze":      "stargaze",
	"sentinel":      "sentinel",
	"quicksilver":   "quicksilver",
	"persistence":   "persistence",
}

// GetChainInfo related to chain Name
func GetChainInfo(ctx types.Context, name string) (registry.ChainInfo, error) {
	return ctx.ChainRegistry().GetChain(ctx.Context(), name)
}

func getChainDenom(chainInfo registry.ChainInfo) (string, error) {
	assetList, err := chainInfo.GetAssetList(context.Background())
	if err != nil {
		return "", err
	}
	assets := assetList.Assets
	if len(assets) > 0 {
		denom := assets[0].Base
		return denom, nil
	} else {
		return "", fmt.Errorf("no assets found for %s chain", chainInfo.ChainName)
	}
}

// Votes on the proposal using the given data and key
func ExecVote(ctx types.Context, chainName, pID, granter, vote,
	fromKey, metadata, memo, gasPrices string, responseWriter slacker.ResponseWriter,
) (string, error) {
	defer func() {
		if r := recover(); r != nil {
			responseWriter.Reply(fmt.Sprintf("Recovered from panic: %v", r))
			log.Println("Recovered from panic:", r)
		}
	}()

	// Fetch chain info from chain registry
	chainInfo, err := GetChainInfo(ctx, chainName)
	if err != nil {
		return "", fmt.Errorf("chain info not found: %v", err)
	}

	//	Use Chain info to select random endpoint
	rpc, err := chainInfo.GetRandomRPCEndpoint(ctx.Context())
	if err != nil {
		return "", fmt.Errorf("failed to get random RPC endpoint on chain %s. Err: %v", chainInfo.ChainID, err)
	}
	denom, err := getChainDenom(chainInfo)
	if err != nil {
		return "", fmt.Errorf("failed to get denom from chain %s: %v", chainInfo.ChainID, err)
	}
	coins, err := sdk.ParseDecCoins(gasPrices)
	if err != nil {
		fmt.Printf("Error while parsing gasPrices :%v\nInvalid fee format, using default fee", err)
		gasPrices = "0.25" + denom
	} else {
		if coins.Empty() {
			gasPrices = "0.25" + denom
		}
	}

	chainConfig := lensclient.ChainClientConfig{
		Key:            fromKey,
		ChainID:        chainInfo.ChainID,
		RPCAddr:        rpc,
		AccountPrefix:  chainInfo.Bech32Prefix,
		KeyringBackend: "test",
		GasPrices:      gasPrices,
		Debug:          true,
		Timeout:        "20s",
		GasAdjustment:  1.4,
		OutputFormat:   "json",
		SignModeStr:    "direct",
		Modules:        lensclient.ModuleBasics,
	}

	curDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error while getting current directory: %v", err)
	}

	// Create client object to pull chain info
	chainClient, err := lensclient.NewChainClient(zap.L(), &chainConfig, curDir, os.Stdin, os.Stdout)
	if err != nil {
		return "", fmt.Errorf("failed to build new chain client for %s. Err: %v", chainInfo.ChainID, err)
	}

	keyAddr, err := ctx.Database().GetKeyAddress(fromKey)
	if err != nil {
		return "", fmt.Errorf("error while getting address of %s key", fromKey)
	}

	proposalID, err := strconv.ParseUint(pID, 10, 64)
	if err != nil {
		return "", fmt.Errorf("unable to convert string to uint64. Err: %v", err)
	}
	voteOption, err := stringToVoteOption(vote)
	if err != nil {
		return "", fmt.Errorf("unable to convert vote option string to sdk vote option. Err: %v", err)
	}

	var msgAny *cdctypes.Any

	// TODO: handle v1beta1 and v1 types
	// validV1Endpoint, err := v.GetValidV1Endpoint(chainInfo)
	// if err != nil {
	// 	return fmt.Errorf("error while getting valid gov v1 endpoint: %v", err)
	// }
	// if validV1Endpoint {
	// 	msgVote := v1.MsgVote{
	// 		ProposalId: proposalID,
	// 		Voter:      valAddr,
	// 		Option:     voteOption,
	// 		Metadata:   metadata,
	// 	}
	// 	msgAny, err = cdctypes.NewAnyWithValue(&msgVote)
	// 	if err != nil {
	// 		return fmt.Errorf("error on converting msg to Any: %v", err)
	// 	}

	// } else {
	msgVote := v1beta1.MsgVote{
		ProposalId: proposalID,
		Voter:      granter,
		Option:     v1beta1.VoteOption(voteOption),
	}
	msgAny, err = cdctypes.NewAnyWithValue(&msgVote)
	if err != nil {
		return "", fmt.Errorf("error on converting msg to Any: %v", err)
	}
	// }

	req := &authz.MsgExec{
		Grantee: keyAddr,
		Msgs:    []*cdctypes.Any{msgAny},
	}

	responseWriter.Reply(fmt.Sprintf("voting %s on %s proposal %d", voteOption, chainName, proposalID))
	// Send msg and get response
	res, err := chainClient.SendMsg(context.Background(), req, memo)
	if err != nil {
		if res != nil {
			return "", fmt.Errorf("failed to vote on proposal: code(%d) msg(%s)", res.Code, res.Logs)
		}
		return "", fmt.Errorf("failed to vote.Err: %v", err)
	}
	pTitle, err := getProposalTitle(chainName, pID)
	if err != nil {
		fmt.Printf("failed to store logs: %v", err)
	} else {
		if err = ctx.Database().AddLog(chainName, pTitle, pID, vote); err != nil {
			fmt.Printf("failed to store logs: %v", err)
		}
	}

	mintscanName := chainName
	if newName, ok := RegisrtyNameToMintscanName[chainName]; ok {
		mintscanName = newName
	}

	return fmt.Sprintf("Trasaction broadcasted: https://mintscan.io/%s/txs/%s", mintscanName, res.TxHash), nil
}

func getProposalTitle(chainName, pID string) (string, error) {
	endpoint, err := endpoints.GetValidEndpointForChain(chainName)
	if err != nil {
		log.Printf("Error in getting valid LCD endpoints for %s chain", chainName)

		return "", err
	}
	var p types.Proposal
	ops := types.HTTPOptions{
		Endpoint: endpoint + "/cosmos/gov/v1beta1/proposals/" + pID,
		Method:   http.MethodGet,
	}
	resp, err := endpoints.HitHTTPTarget(ops)
	if err != nil {
		log.Printf("Error while getting http response: %v", err)
		return "", err
	}

	err = json.Unmarshal(resp.Body, &p)
	if err != nil {
		log.Printf("Error while unmarshalling the proposals: %v", err)
		return "", err
	}

	return p.Content.Title, nil
}

// Converts the string to a acceptable vote format
func stringToVoteOption(str string) (v1.VoteOption, error) {
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

// Gets valid v1 endpoints from Chain registry
func GetValidV1Endpoint(chainInfo registry.ChainInfo) (bool, error) {
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
