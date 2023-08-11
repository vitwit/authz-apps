package utils

import (
	"encoding/json"
	"net/http"
)

// HasAuthzGrant returns true if authz permission is exist for the provided
// parameters.
func HasAuthzGrant(lcd, granter, grantee, typeURL string) (bool, error) {
	response, err := http.Get(lcd + "/cosmos/authz/v1beta1/grants?granter=" + granter + "&grantee=" + grantee + "&msg_url_type=" + typeURL)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()

	var jsonData struct {
		Grants []interface{} `json:"grants"`
	}

	if err := json.NewDecoder(response.Body).Decode(&jsonData); err != nil {
		return false, err
	}

	if len(jsonData.Grants) > 0 {
		return true, nil
	}

	return false, nil
}
