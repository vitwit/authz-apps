package endpoints

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/likhita-809/lens-bot/types"
	registry "github.com/strangelove-ventures/lens/client/chain_registry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Gets a valid LCD endpoint for a given chain
func GetValidEndpointForChain(chainName string) (validLCDEndpoint string, err error) {
	cr := registry.DefaultChainRegistry(zap.New(zapcore.NewNopCore()))

	chainInfo, err := cr.GetChain(context.Background(), chainName)
	if err != nil {
		return "", err
	}

	AllLCDEndpoints, err := GetAllLCDEndpoints(chainInfo)
	if err != nil {
		return "", err
	}

	validLCDEndpoint, err = GetValidLCDEndpoint(AllLCDEndpoints)
	if err != nil {
		return "", err
	}

	validLCDEndpoint = strings.TrimSuffix(validLCDEndpoint, "/")

	return validLCDEndpoint, nil
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

// Gets proposals current stauts
func GetStatus(endpoint string) bool {
	ops := types.HTTPOptions{
		Endpoint: endpoint + "/cosmos/gov/v1beta1/proposals",
		Method:   http.MethodGet,
	}

	resp, err := HitHTTPTarget(ops)
	if err != nil {
		log.Printf("Error in external rpc: %v", err)
		log.Printf("⛔⛔ Unreachable to EXTERNAL RPC :: %s and the ERROR is : %v\n\n", ops.Endpoint, err.Error())
		return false
	}
	return resp.StatusCode == http.StatusOK
}

// HitHTTPTarget to hit the target and get response
func HitHTTPTarget(ops types.HTTPOptions) (*types.PingResp, error) {
	req, err := newHTTPRequest(ops)
	if err != nil {
		return nil, err
	}

	httpcli := http.Client{Timeout: time.Duration(30 * time.Second)}
	resp, err := httpcli.Do(req)
	if err != nil {
		return nil, err
	}

	res, err := makeResponse(resp)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Adds the Query parameters
func addQueryParameters(req *http.Request, queryParams types.QueryParams) {
	q := req.URL.Query()
	for key, value := range queryParams {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()
}

// newHTTPRequest to make a new http request
func newHTTPRequest(ops types.HTTPOptions) (*http.Request, error) {
	// make new request
	req, err := http.NewRequest(ops.Method, ops.Endpoint, bytes.NewBuffer(ops.Body))
	if err != nil {
		return nil, err
	}

	// Add any query parameters to the URL.
	if len(ops.QueryParams) != 0 {
		addQueryParameters(req, ops.QueryParams)
	}

	return req, nil
}

// Creates response
func makeResponse(res *http.Response) (*types.PingResp, error) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return &types.PingResp{}, err
	}

	response := &types.PingResp{
		StatusCode: res.StatusCode,
		Body:       body,
	}
	_ = res.Body.Close()
	return response, nil
}
