package targets

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	registry "github.com/strangelove-ventures/lens/client/chain_registry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func GetValidLCDEndpoints() (validEndpoints []string, err error) {
	cr := registry.DefaultChainRegistry(zap.New(zapcore.NewNopCore()))

	chains, err := cr.ListChains(context.Background())
	if err != nil {
		return []string{}, err
	}

	for _, chainName := range chains {
		chainInfo, _ := cr.GetChain(context.Background(), chainName)
		AllLCDEndpoints, _ := GetAllLCDEndpoints(chainInfo)
		validLCDEndpoint, _ := GetValidLCDEndpoint(AllLCDEndpoints)
		validLCDEndpoint = strings.TrimSuffix(validLCDEndpoint, "/")
		validEndpoints = append(validEndpoints, validLCDEndpoint)

	}
	return validEndpoints, nil
}

func GetAllLCDEndpoints(c registry.ChainInfo) (out []string, err error) {
	for _, endpoint := range c.Apis.Rest {
		u, err := url.Parse(endpoint.Address)
		if err != nil {
			return nil, err
		}
		var port string
		if u.Port() == "" {
			switch u.Scheme {
			case "https":
				port = "443"
			case "http":
				port = "80"
			default:
				return nil, fmt.Errorf("invalid or unsupported url scheme: %v", u.Scheme)
			}
		} else {
			port = u.Port()
		}

		out = append(out, fmt.Sprintf("%s://%s:%s%s", u.Scheme, u.Hostname(), port, u.Path))
	}
	return
}

func GetValidLCDEndpoint(endpoints []string) (string, error) {
	var validEndpoint bool
	for _, endpoint := range endpoints {
		validEndpoint = GetStatus(endpoint)
		if validEndpoint {
			return endpoint, nil
		}
	}
	return "", nil
}

func GetStatus(endpoint string) bool {
	ops := HTTPOptions{
		Endpoint: endpoint + "/cosmos/gov/v1beta1/proposals",
		Method:   http.MethodGet,
	}

	resp, err := HitHTTPTarget(ops)
	if err != nil {
		log.Printf("Error in external rpc: %v", err)
		fmt.Printf("⛔⛔ Unreachable to EXTERNAL RPC :: %s and the ERROR is : %v\n\n", ops.Endpoint, err.Error())
		return false
	}
	return resp.StatusCode == http.StatusOK
}
