#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

LIQUIDITY=1000000000000
GAS="--gas-prices 0.1uosmo --gas auto --gas-adjustment 1.3"

echo "Determining relevant channels..."
dydx_to_osmo_client=$(GET_CLIENT_ID_FROM_CHAIN_ID DYDX OSMO)
dydx_to_osmo_connection=$(GET_CONNECTION_ID_FROM_CLIENT_ID DYDX $dydx_to_osmo_client)
dydx_to_osmo_channel=$(GET_TRANSFER_CHANNEL_ID_FROM_CONNECTION_ID DYDX $dydx_to_osmo_connection)
osmo_to_dydx_channel=$(GET_COUNTERPARTY_TRANSFER_CHANNEL_ID DYDX $dydx_to_osmo_channel)

echo -e "\nDYDX -> OSMO:"
echo "  Client: $dydx_to_osmo_client"
echo "  Connection: $dydx_to_osmo_connection"
echo "  Transfer Channel: $dydx_to_osmo_channel -> $osmo_to_dydx_channel"

noble_to_osmo_client=$(GET_CLIENT_ID_FROM_CHAIN_ID NOBLE OSMO)
noble_to_osmo_connection=$(GET_CONNECTION_ID_FROM_CLIENT_ID NOBLE $noble_to_osmo_client)
noble_to_osmo_channel=$(GET_TRANSFER_CHANNEL_ID_FROM_CONNECTION_ID NOBLE $noble_to_osmo_connection)
osmo_to_noble_channel=$(GET_COUNTERPARTY_TRANSFER_CHANNEL_ID NOBLE $noble_to_osmo_channel)

echo -e "\nNOBLE -> OSMO:"
echo "  Client: $noble_to_osmo_client"
echo "  Connection: $noble_to_osmo_connection"
echo "  Transfer Channel: $noble_to_osmo_channel -> $osmo_to_noble_channel"

echo -e "\nSending dydx/usdc to osmosis for initial liquidity..."

echo ">>> DYDX to Osmosis:"
$DYDX_MAIN_CMD tx ibc-transfer transfer transfer $dydx_to_osmo_channel $(OSMO_ADDRESS) ${LIQUIDITY}${DYDX_DENOM} \
    --from ${DYDX_VAL_PREFIX}1 -y | TRIM_TX

echo ">>> USDC to Osmosis:"
$NOBLE_MAIN_CMD tx ibc-transfer transfer transfer $noble_to_osmo_channel $(OSMO_ADDRESS) ${LIQUIDITY}${USDC_DENOM} \
    --from ${NOBLE_VAL_PREFIX}1 -y | TRIM_TX
sleep 15

echo ">>> Balances:"
$OSMO_MAIN_CMD q bank balances $(OSMO_ADDRESS)

echo -e "\nDetermining IBC Denoms..."
dydx_denom_on_osmo=$(GET_IBC_DENOM OSMO $osmo_to_dydx_channel $DYDX_DENOM)
usdc_denom_on_osmo=$(GET_IBC_DENOM OSMO $osmo_to_noble_channel $USDC_DENOM)

echo "  ibc/dydx on Osmosis: $dydx_denom_on_osmo"
echo "  ibc/usdc on Osmosis: $usdc_denom_on_osmo"

echo -e "\nCreating dydx/usdc pool on osmosis..."
pool_file=${STATE}/${OSMO_NODE_PREFIX}1/pool.json
cat << EOF > $pool_file
{
	"weights": "5${dydx_denom_on_osmo},5${usdc_denom_on_osmo}",
	"initial-deposit": "1000000000000${dydx_denom_on_osmo},1000000000000${usdc_denom_on_osmo}",
	"swap-fee": "0.01",
	"exit-fee": "0.0",
	"future-governor": ""
}
EOF

$OSMO_MAIN_CMD tx gamm create-pool --pool-file $pool_file --from ${OSMO_VAL_PREFIX}1 -y $GAS | TRIM_TX
sleep 5

echo -e "\n>>> Pools:"
$OSMO_MAIN_CMD q gamm pools
