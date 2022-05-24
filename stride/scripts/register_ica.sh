#!/bin/bash

STRIDE_ACCT_1=stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7
STRIDE_CHAIN_ID=STRIDE_1
# $STR1_EXEC tx ibc-transfer transfer transfer channel-1 $GAIA_ADDRESS_1 1000ustrd --home /stride/.strided --keyring-backend test --from val1 --chain-id STRIDE_1 -y

# Register an ICA on gaia from stride
# --node tcp://localhost:16657
strided tx stakeibc register-ica --from $STRIDE_ACCT_1 --connection-id connection-0 --chain-id $STRIDE_CHAIN /stride/.strided --keyring-backend test

# Query the address of the interchain account
icad query intertx interchainaccounts connection-0 $DEMOWALLET_1 --home ./data/test-1 --node tcp://localhost:16657

# Store the interchain account address by parsing the query result: cosmos1hd0f4u7zgptymmrn55h3hy20jv2u0ctdpq23cpe8m9pas8kzd87smtf8al
export ICA_ADDR=$(icad query intertx interchainaccounts connection-0 $DEMOWALLET_1 --home ./data/test-1 --node tcp://localhost:16657 -o json | jq -r '.interchain_account_address') && echo $ICA_ADDR




# Register an interchain account on behalf of DEMOWALLET_1 where chain test-2 is the interchain accounts host
icad tx intertx register --from $DEMOWALLET_1 --connection-id connection-0 --chain-id test-1 --home ./data/test-1 --node tcp://localhost:16657 --keyring-backend test -y

# Query the address of the interchain account
icad query intertx interchainaccounts connection-0 $DEMOWALLET_1 --home ./data/test-1 --node tcp://localhost:16657

# Store the interchain account address by parsing the query result: cosmos1hd0f4u7zgptymmrn55h3hy20jv2u0ctdpq23cpe8m9pas8kzd87smtf8al
export ICA_ADDR=$(icad query intertx interchainaccounts connection-0 $DEMOWALLET_1 --home ./data/test-1 --node tcp://localhost:16657 -o json | jq -r '.interchain_account_address') && echo $ICA_ADDR

