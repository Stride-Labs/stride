#!/bin/bash

echo "$STRIDE_ADDRESS_1 balances"
$STR1_EXEC q bank balances $STRIDE_ADDRESS_1 --denom ustrd

# transfer 100strd from val1 on stride to gval1 on gaia
echo "transferring 1000strd tokens..."
$STR1_EXEC tx ibc-transfer transfer transfer channel-1 $GAIA_ADDRESS_1 1000ustrd --home /stride/.strided --keyring-backend test --from val1 --chain-id STRIDE_1 -y

sleep 9
echo "$STRIDE_ADDRESS_1 balances"
$STR1_EXEC q bank balances $STRIDE_ADDRESS_1 --denom ustrd
