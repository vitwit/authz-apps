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

// Gets a valid LCD endpoint for a given chain
func GetValidEndpointForChain(chainName string) (validLCDEndpoint string, validV1Endpoint bool, err error) {
	cr := registry.DefaultChainRegistry(zap.New(zapcore.NewNopCore()))

	chainInfo, err := cr.GetChain(context.Background(), chainName)
	if err != nil {
		return "", false, err
	}

	AllLCDEndpoints, err := GetAllLCDEndpoints(chainInfo)
	if err != nil {
		return "", false, err
	}

	validLCDEndpoint, validV1Endpoint = GetValidLCDEndpoint(AllLCDEndpoints)
	if validLCDEndpoint == "" {
		return "", false, fmt.Errorf("valid LCD endpoint not found for chain %s", chainName)
	}

	validLCDEndpoint = strings.TrimSuffix(validLCDEndpoint, "/")

	return validLCDEndpoint, validV1Endpoint, nil
}

// Gets all LCD endpoints present in a chain
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
				log.Printf("invalid or unsupported url scheme: %v", u.Scheme)
			}
		} else {
			port = u.Port()
		}

		out = append(out, fmt.Sprintf("%s://%s:%s%s", u.Scheme, u.Hostname(), port, u.Path))
	}
	return
}

// Gets a valid LCD endpoint from all the LCD endpoints
func GetValidLCDEndpoint(endpoints []string) (endpoint string, validV1Endpoint bool) {
	for _, endpoint := range endpoints {
		validV1Endpoint = GetStatus(endpoint)
		if validV1Endpoint {
			return endpoint, true
		}
	}
	return "", false
}

// Gets proposals current stauts
func GetStatus(endpoint string) (govV1Support bool) {
	ops := HTTPOptions{
		Endpoint: endpoint + "/cosmos/gov/v1/proposals",
		Method:   http.MethodGet,
	}
	resp, err := HitHTTPTarget(ops)
	if err != nil {
		log.Printf("Error in external rpc: %v", err)
		log.Printf("⛔⛔ Unreachable to EXTERNAL RPC :: %s and the ERROR is : %v\n\n", ops.Endpoint, err.Error())
		return false
	}
	if resp.StatusCode == http.StatusOK {
		govV1Support = true
	} else {
		ops = HTTPOptions{
			Endpoint: endpoint + "/cosmos/gov/v1beta1/proposals",
			Method:   http.MethodGet,
		}
		resp, err = HitHTTPTarget(ops)
		if err != nil {
			log.Printf("Error in external rpc: %v", err)
			log.Printf("⛔⛔ Unreachable to EXTERNAL RPC :: %s and the ERROR is : %v\n\n", ops.Endpoint, err.Error())
			return false
		}
		if resp.StatusCode == http.StatusOK {
			govV1Support = true
		}
	}
	return
}
