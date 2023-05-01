package voting

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	lensclient "github.com/strangelove-ventures/lens/client"
	registry "github.com/strangelove-ventures/lens/client/chain_registry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func GetChainName(chainID string, cr registry.ChainRegistry) (string, error) {
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
	return "", fmt.Errorf("chain name not found")
}

func ExecVote(chainID, pID, valAddr, vote, memo, gas, fees string) error {
	// Fetch chain info from chain registry
	cr := registry.DefaultChainRegistry(zap.New(zapcore.NewNopCore()))

	chainName, err := GetChainName(chainID, cr)
	if err != nil {
		log.Fatalf("chain name not found")
	}
	chainInfo, err := cr.GetChain(context.Background(), chainName)
	if err != nil {
		log.Fatalf("Failed to get chain info. Err: %v \n", err)
	}

	//	Use Chain info to select random endpoint
	rpc, err := chainInfo.GetRandomRPCEndpoint(context.Background())
	if err != nil {
		log.Fatalf("failed to get random RPC endpoint on chain %s. Err: %v", chainInfo.ChainID, err)
	}

	chainConfig := lensclient.ChainClientConfig{
		Key:            "default",
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
		log.Fatalf("error while getting current directory: %v", err)
	}
	// Create client object to pull chain info
	chainClient, err := lensclient.NewChainClient(zap.L(), &chainConfig, curDir, os.Stdin, os.Stdout)
	if err != nil {
		log.Fatalf("failed to build new chain client for %s. Err: %v", chainInfo.ChainID, err)
	}

	keyName := "my_key"
	keyOp, err := chainClient.AddKey(keyName, sdk.CoinType)
	if err != nil {
		log.Fatalf("error while adding key: %v", err)
	}
	chainConfig.Key = keyName

	proposalID, err := strconv.ParseUint(pID, 10, 64)
	if err != nil {
		return fmt.Errorf("unable to convert string to uint64. Err: %v", err)
	}
	voteOption, err := stringToVoteOption(vote)
	if err != nil {
		return fmt.Errorf("unable to convert vote option string to sdk vote option. Err: %v", err)
	}

	msgVote := v1.MsgVote{
		ProposalId: proposalID,
		Voter:      valAddr,
		Option:     voteOption,
		Metadata:   "",
	}
	msgAny, err := cdctypes.NewAnyWithValue(&msgVote)
	if err != nil {
		log.Fatalf("error on converting msg to Any: %v", err)
	}

	req := &authz.MsgExec{
		Grantee: keyOp.Address,
		Msgs:    []*cdctypes.Any{msgAny},
	}

	// Send msg and get response
	res, err := chainClient.SendMsg(context.Background(), req, memo)
	if err != nil {
		if res != nil {
			log.Fatalf("failed to vote on proposal: code(%d) msg(%s)", res.Code, res.Logs)
		}
		log.Fatalf("Failed to vote.Err: %v", err)
	}
	return nil
}

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
