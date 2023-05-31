package targets

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/likhita-809/lens-bot/database"
)

type GrantRetriever interface {
	GetGrants(string) ([]Grants, error)
}

func GetGrants(endpoint string) ([]Grants, error) {
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var g []Grants
	err = json.Unmarshal(body, &g)
	if err != nil {
		log.Printf("Error while unmarshalling the grants : %v", err)
		return nil, err
	}
	return g, nil
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
		validEndpoints, err := GetValidLCDEndpoints(key.ChainName)
		if err != nil {
			log.Printf("Error in getting valid LCD endpoints for %s chain", key.ChainName)

			return err
		}
		for _, val := range validators {
			if val.ChainName == key.ChainName {
				for _, endpoint := range validEndpoints {
					ops := HTTPOptions{
						Endpoint: endpoint + "/cosmos/authz/v1beta1/grants?granter=" + val.ChainName + "&grantee=" + key.KeyAddress,
						Method:   http.MethodGet,
					}
					g1, err := GetGrants(ops.Endpoint)
					if err != nil {
						return err
					}
					g := g1[0]
					// resp, err := HitHTTPTarget(ops)
					// if err != nil {
					// 	log.Printf("Error while getting http response: %v", err)
					// 	return err
					// }

					layout := "2006-01-02"
					expireDate := g.Grants.Authorization.Expiration
					var expireTime int64
					end, err := time.Parse(layout, expireDate)
					if err != nil {
						return err
					}
					expireTime = end.Unix()
					var status string

					if g.Grants.Authorization.Type == "/cosmos.authz.v1beta1.GenericAuthorization" &&
						(g.Grants.Authorization.Msg == "/cosmos.gov.v1beta1.MsgVote" ||
							g.Grants.Authorization.Msg == "/cosmos.gov.v1.MsgVote") {
						if expireTime > time.Now().UTC().Unix() {
							status = "true"
						} else {
							status = "false"
						}
						db.UpdateKey(status, key.KeyAddress)
					}

				}

			}

		}
	}
	return err
}
