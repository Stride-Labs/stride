#!/bin/bash

# TODO(TEST-80) migrate tests to bats 

STR1_EXEC="docker-compose --ansi never exec -T stride1 strided --home /stride/.strided --chain-id STRIDE"
GAIA1_EXEC="docker-compose --ansi never exec -T gaia1 gaiad --home /gaia/.gaiad --chain-id GAIA"
ICQ_EXEC="docker-compose --ansi never exec -T icq interchain-queries"

# Set up: fund account on Stride
# TODO(TEST-83) pull these addresses dynamically using jq
$STR1_EXEC tx bank send val1 stride12vfkpj7lpqg0n4j68rr5kyffc6wu55dzqewda4 5000000ustrd \
     --keyring-backend test -y
sleep 10

# # Test: query balance
# TODO(TEST-83) pull these addresses dynamically using jq
$STR1_EXEC tx interchainquery query-balance GAIA cosmos1t2aqq3c6mt8fa6l5ady44manvhqf77sywjcldv uatom \
    --connection-id connection-1 --keyring-backend test -y --from val1
sleep 15
# $STR1_EXEC q txs --events message.module=interchainquery&limit=1
$STR1_EXEC q txs --events message.action=/stride.interchainquery.MsgSubmitQueryResponse&limit=1 # | jq '.logs'


# Test: query exchange rate
$STR1_EXEC tx interchainquery query-exchangerate GAIA --keyring-backend test -y --from val1
sleep 15
# $STR1_EXEC q txs --events message.module=interchainquery&limit=1
$STR1_EXEC q txs --events message.module=interchainquery&limit=1 # | jq '.logs'

# # Test: query delegated balance | NOTE: need to instantiate delegationAccount before this test will work!
# $STR1_EXEC tx interchainquery query-delegatedbalance GAIA --keyring-backend test -y --from val1
# sleep 15
# # $STR1_EXEC q txs --events message.module=interchainquery&limit=1
# $STR1_EXEC q txs --events message.module=interchainquery&limit=1 # | jq '.logs'