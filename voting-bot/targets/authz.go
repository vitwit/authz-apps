package targets

import (
	"log"

	"github.com/vitwit/authz-apps/voting-bot/endpoints"
	"github.com/vitwit/authz-apps/voting-bot/types"
	"github.com/vitwit/authz-apps/voting-bot/utils"
)

const MSG_VOTE_TYPEURL_V1BETA1 = "/cosmos.gov.v1beta1.MsgVote"
const MSG_VOTE_TYPEURL_V1 = "/cosmos.gov.v1.MsgVote"

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

				granter, err := convertValAddrToAccAddr(ctx, val.Address, key.ChainName)
				if err != nil {
					return err
				}

				hasAuthz, err := utils.HasAuthzGrant(validEndpoint, granter, key.KeyAddress, MSG_VOTE_TYPEURL_V1BETA1)
				if err != nil {
					return err
				}

				if hasAuthz {
					if err := ctx.Database().UpdateAuthzStatus("true", key.KeyAddress); err != nil {
						return err
					}
				}

				hasAuthz, err = utils.HasAuthzGrant(validEndpoint, granter, key.KeyAddress, MSG_VOTE_TYPEURL_V1)
				if err != nil {
					return err
				}

				if hasAuthz {
					if err := ctx.Database().UpdateAuthzStatus("true", key.KeyAddress); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
