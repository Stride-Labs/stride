#!/bin/bash

# TODO(TEST-80) migrate tests to bats 

STR1_EXEC="docker-compose --ansi never exec -T stride1 strided --home /stride/.strided --chain-id STRIDE"
GAIA1_EXEC="docker-compose --ansi never exec -T gaia1 gaiad --home /gaia/.gaiad --chain-id GAIA"
ICQ_EXEC="docker-compose --ansi never exec -T icq interchain-queries"


# Test: query exchange rate
$STR1_EXEC tx interchainquery query-exchangerate GAIA --keyring-backend test -y --from val1
sleep 15
$STR1_EXEC q txs --events message.action=/stride.interchainquery.MsgSubmitQueryResponse --limit=1

# # Test: query delegated balance 
$STR1_EXEC tx interchainquery query-delegatedbalance GAIA --keyring-backend test -y --from val1
sleep 15
$STR1_EXEC q txs --events message.action=/stride.interchainquery.MsgSubmitQueryResponse --limit=1
