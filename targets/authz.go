package targets

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/likhita-809/lens-bot/endpoints"
	"github.com/likhita-809/lens-bot/types"
)

func getAuthzGrants(endpoint string) ([]interface{}, error) {
	response, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("Failed to read response body:", err)
		return nil, err
	}
	var jsonData map[string]interface{}
	err = json.Unmarshal(body, &jsonData)
	if err != nil {
		log.Println("Failed to parse JSON data:", err)
		return nil, err
	}
	grants := jsonData["grants"].([]interface{})
	return grants, nil
}

// SyncAuthzStatus iterates over all validators account and update
// authz grant status
func SyncAuthzStatus(ctx types.Context) error {
	keys, err := ctx.Database().GetKeys()
	if err != nil {
		return err
	}
	validators, err := ctx.Database().GetValidators()
	if err != nil {
		return err
	}

	for _, key := range keys {
		validEndpoint, err := endpoints.GetValidEndpointForChain(key.ChainName)
		if err != nil {
			log.Printf("Error in getting valid LCD endpoints for %s chain", key.ChainName)
			return err
		}
		for _, val := range validators {
			if val.ChainName == key.ChainName {
				ops := types.HTTPOptions{
					Endpoint: validEndpoint + "/cosmos/authz/v1beta1/grants?granter=" + val.Address + "&grantee=" + key.KeyAddress + "&msg_url_type=/cosmos.gov.v1beta1.MsgVote",
					Method:   http.MethodGet,
				}
				g1, err := getAuthzGrants(ops.Endpoint)
				if err != nil {
					return err
				}
				if len(g1) > 0 {
					if err := ctx.Database().UpdateAuthzStatus("true", key.KeyAddress); err != nil {
						return err
					}
				}

				ops = types.HTTPOptions{
					Endpoint: validEndpoint + "/cosmos/authz/v1beta1/grants?granter=" + val.Address + "&grantee=" + key.KeyAddress + "&msg_url_type=/cosmos.gov.v1.MsgVote",
					Method:   http.MethodGet,
				}
				g2, err := getAuthzGrants(ops.Endpoint)
				if err != nil {
					return err
				}

				if len(g2) > 0 {
					if err := ctx.Database().UpdateAuthzStatus("true", key.KeyAddress); err != nil {
						return err
					}
				}
				if err := ctx.Database().UpdateAuthzStatus("false", key.KeyAddress); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
