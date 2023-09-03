package utils

import (
	"encoding/json"
	"net/http"

	"github.com/vitwit/authz-apps/voting-bot/endpoints"
	"github.com/vitwit/authz-apps/voting-bot/types"
)

// HasAuthzGrant returns true if authz permission is exist for the provided
// parameters.
func HasAuthzGrant(endpoint, granter, grantee, typeURL string) (bool, error) {
	resp, err := endpoints.HitHTTPTarget(types.HTTPOptions{
		Endpoint:    endpoint + "/cosmos/authz/v1beta1/grants",
		Method:      http.MethodGet,
		QueryParams: types.QueryParams{"granter": granter, "grantee": grantee, "msg_type_url": typeURL},
	})
	if err != nil {
		return false, err
	}

	var authzResp types.Grants
	if err := json.Unmarshal(resp.Body, &authzResp); err != nil {
		return false, err
	}

	if len(authzResp.Grants) > 0 {
		return true, nil
	}

	return false, nil

}
