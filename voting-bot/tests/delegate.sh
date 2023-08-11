#!/bin/bash
source .env

VAL_ADDR=$(gaiad  q staking validators --output json | jq -r '.validators[0].operator_address')
sed -i "s/^val_addr=.*/val_addr=$VAL_ADDR/" .env


GRANTEE_ADDR=$(gaiad keys show "$GRANTEE" -a --keyring-backend test)
sed -i "s/^grantee_addr=.*/grantee_addr=$GRANTEE_ADDR/" .env


gaiad tx staking delegate "$VAL_ADDR" 100000stake --chain-id "$CHAINID" --from "$KEY" -y -b block --keyring-backend test

gaiad tx authz grant "$GRANTEE_ADDR" generic --msg-type /cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward --from "$KEY" --chain-id "$CHAINID" -y -b block --keyring-backend test

gaiad tx authz grant "$GRANTEE_ADDR" generic --msg-type /cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward --from "$VALIDATOR" --chain-id "$CHAINID" -y -b block --keyring-backend test

gaiad tx authz grant "$GRANTEE_ADDR" generic --msg-type /cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission --from "$VALIDATOR" --chain-id "$CHAINID" -y -b block --keyring-backend test