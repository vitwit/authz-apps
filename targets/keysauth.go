package targets

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/likhita-809/lens-bot/database"
)

type GrantRetriever interface {
	GetGrants(string) ([]Grants, error)
}

func GetGrants(endpoint string) ([]interface{}, error) {
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

func KeyAuthorization(db *database.Sqlitedb) error {
	keys, err := db.GetKeys()
	if err != nil {
		return err
	}
	validators, err := db.GetValidators()
	if err != nil {
		return err
	}

	for _, key := range keys {
		validEndpoint, err := GetValidEndpointForChain(key.ChainName)
		if err != nil {
			log.Printf("Error in getting valid LCD endpoints for %s chain", key.ChainName)

			return err
		}
		for _, val := range validators {
			if val.ChainName == key.ChainName {
				ops := HTTPOptions{
					Endpoint: validEndpoint + "/cosmos/authz/v1beta1/grants?granter=" + val.ChainName + "&grantee=" + key.KeyAddress + "&msg_url_type=/cosmos.gov.v1beta1.MsgVote",
					Method:   http.MethodGet,
				}
				g1, err := GetGrants(ops.Endpoint)
				if err != nil {
					return err
				}
				var status string
				if len(g1) > 0 {
					status = "true"
				} else {

					ops := HTTPOptions{
						Endpoint: validEndpoint + "/cosmos/authz/v1beta1/grants?granter=" + val.ChainName + "&grantee=" + key.KeyAddress + "&msg_url_type=/cosmos.gov.v1beta1.MsgVote",
						Method:   http.MethodGet,
					}
					g2, err := GetGrants(ops.Endpoint)
					if err != nil {
						return err
					}
					if len(g2) > 0 {
						status = "true"
					} else {
						status = "false"
					}
					err = db.UpdateAuthzStatus(status, key.KeyAddress)
					if err != nil {
						return err
					}
				}
			}

		}
	}
	return err
}
