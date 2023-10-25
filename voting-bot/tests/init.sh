#!/bin/bash

source .env

# validate dependencies are installed
command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }

# remove existing daemon and client
rm -rf ~/.gaia*

# setup gaia config
gaiad config chain-id $CHAINID
gaiad config broadcast-mode block
gaiad config keyring-backend test

# if $KEY exists it should be deleted
gaiad keys add "$KEY"

# Create a named pipe (FIFO)
mkfifo mypipe
gaiad keys add "$VALIDATOR" --recover < mypipe &
echo "$VALIDATOR_MNEMONIC" > mypipe
sleep 1
rm mypipe

mkfifo mypipe
gaiad keys add "$GRANTEE" --recover < mypipe &
echo "$GRANTEE_MNEMONIC" > mypipe
sleep 1
rm mypipe


# Set moniker and chain-id for Ethermint (Moniker can be anything, chain-id must be an integer)
gaiad init "$MONIKER" --chain-id "$CHAINID"

# Allocate genesis accounts (cosmos formatted addresses)
gaiad add-genesis-account "$KEY" 100000000000000000000000000stake 
gaiad add-genesis-account "$VALIDATOR" 100000000000000000000000000stake
gaiad add-genesis-account "$GRANTEE" 100000000000000000000000000stake

# Sign genesis transaction
gaiad gentx "$VALIDATOR" 1000000000000000000000stake --chain-id "$CHAINID"

# Collect genesis tx
gaiad collect-gentxs

# Run this to ensure everything worked and that the genesis file is setup correctly
gaiad validate-genesis

# Start the node
gaiad start