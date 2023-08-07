package targets

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	authtypes "github.com/cosmos/cosmos-sdk/x/authz"
	"google.golang.org/grpc"

	"github.com/vitwit/authz-apps/voting-bot/types"
	"google.golang.org/grpc/credentials"
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

	grantsInterface, ok := jsonData["grants"]
	if !ok || grantsInterface == nil {
		// handle the case when "grants" is not present or nil
		return []interface{}(nil), nil
	}

	grants, ok := grantsInterface.([]interface{})
	if !ok {
		// handle the case when "grants" is not a slice of interfaces
		return []interface{}(nil), nil
	}
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
	creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: false})
	ctx1, _ := context.WithTimeout(context.Background(), 5*time.Second)
	grpcConn, err := grpc.DialContext(ctx1, "grpc.cosmos.dragonstake.io:443", grpc.WithTransportCredentials(creds))
	if err != nil {
		panic(err)
	}
	defer grpcConn.Close()
	queryclient := authtypes.NewQueryClient(grpcConn)

	for _, key := range keys {
		// validEndpoint, err := endpoints.GetValidEndpointForChain(key.ChainName)
		// if err != nil {
		// 	log.Printf("Error in getting valid LCD endpoints for %s chain", key.ChainName)
		// 	return err
		// }
		for _, val := range validators {
			if val.ChainName == key.ChainName {

				accAddr, err := convertValAddrToAccAddr(ctx, val.Address, key.ChainName)
				if err != nil {
					return err
				}
				v1beta1Res, err := queryclient.Grants(
					context.Background(),
					&authtypes.QueryGrantsRequest{
						Granter:    accAddr,
						Grantee:    key.KeyAddress,
						MsgTypeUrl: "&msg_url_type=/cosmos.gov.v1beta1.MsgVote",
					},
				)
				if err != nil {
					panic(err)
				}

				// ops := types.HTTPOptions{
				// 	Endpoint: validEndpoint + "/cosmos/authz/v1beta1/grants?granter=" + accAddr + "&grantee=" + key.KeyAddress + "&msg_url_type=/cosmos.gov.v1beta1.MsgVote",
				// 	Method:   http.MethodGet,
				// }
				// // g1, err := getAuthzGrants(ops.Endpoint)
				// if err != nil {
				// 	return err
				// }
				if len(v1beta1Res.Grants) > 0 {
					if err := ctx.Database().UpdateAuthzStatus("true", key.KeyAddress); err != nil {
						return err
					}
				}

				// ops = types.HTTPOptions{
				// 	Endpoint: validEndpoint + "/cosmos/authz/v1beta1/grants?granter=" + accAddr + "&grantee=" + key.KeyAddress + "&msg_url_type=/cosmos.gov.v1.MsgVote",
				// 	Method:   http.MethodGet,
				// }
				// g2, err := getAuthzGrants(ops.Endpoint)
				// if err != nil {
				// 	return err
				// }
				v1Res, err := queryclient.Grants(
					context.Background(),
					&authtypes.QueryGrantsRequest{
						Granter:    accAddr,
						Grantee:    key.KeyAddress,
						MsgTypeUrl: "&msg_url_type=/cosmos.gov.v1beta1.MsgVote",
					},
				)
				if err != nil {
					panic(err)
				}
				if len(v1Res.Grants) > 0 {
					if err := ctx.Database().UpdateAuthzStatus("true", key.KeyAddress); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
