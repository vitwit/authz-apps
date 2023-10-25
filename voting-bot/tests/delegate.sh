#!/bin/bash
source .env

VAL_ADDR=$(gaiad  q staking validators --output json | jq -r '.validators[0].operator_address')
sed -i "s/^val_addr=.*/val_addr=$VAL_ADDR/" .env


GRANTEE_ADDR=$(gaiad keys show "$GRANTEE" -a)
sed -i "s/^grantee_addr=.*/grantee_addr=$GRANTEE_ADDR/" .env


gaiad tx staking delegate "$VAL_ADDR" 100000stake --from "$KEY" -y

gaiad tx authz grant "$GRANTEE_ADDR" generic --msg-type /cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward --from "$VALIDATOR" -y 

gaiad tx authz grant "$GRANTEE_ADDR" generic --msg-type /cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission --from "$VALIDATOR" -y