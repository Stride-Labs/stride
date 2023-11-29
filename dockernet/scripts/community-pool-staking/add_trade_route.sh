#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

GAS="--gas-prices 0.1ustrd --gas auto --gas-adjustment 1.3"

echo "Determining relevant connections and channels..."
stride_to_noble_client=$(GET_CLIENT_ID_FROM_CHAIN_ID STRIDE NOBLE)
stride_to_noble_connection=$(GET_CONNECTION_ID_FROM_CLIENT_ID STRIDE $stride_to_noble_client)

stride_to_osmo_client=$(GET_CLIENT_ID_FROM_CHAIN_ID STRIDE OSMO)
stride_to_osmo_connection=$(GET_CONNECTION_ID_FROM_CLIENT_ID STRIDE $stride_to_osmo_client)

dydx_to_noble_client=$(GET_CLIENT_ID_FROM_CHAIN_ID DYDX NOBLE)
dydx_to_noble_connection=$(GET_CONNECTION_ID_FROM_CLIENT_ID DYDX $dydx_to_noble_client)
dydx_to_noble_channel=$(GET_TRANSFER_CHANNEL_ID_FROM_CONNECTION_ID DYDX $dydx_to_noble_connection)

noble_to_dydx_client=$(GET_CLIENT_ID_FROM_CHAIN_ID NOBLE DYDX)
noble_to_dydx_connection=$(GET_CONNECTION_ID_FROM_CLIENT_ID NOBLE $noble_to_dydx_client)
noble_to_dydx_channel=$(GET_TRANSFER_CHANNEL_ID_FROM_CONNECTION_ID NOBLE $noble_to_dydx_connection)

noble_to_osmo_client=$(GET_CLIENT_ID_FROM_CHAIN_ID NOBLE OSMO)
noble_to_osmo_connection=$(GET_CONNECTION_ID_FROM_CLIENT_ID NOBLE $noble_to_osmo_client)
noble_to_osmo_channel=$(GET_TRANSFER_CHANNEL_ID_FROM_CONNECTION_ID NOBLE $noble_to_osmo_connection)

osmo_to_dydx_client=$(GET_CLIENT_ID_FROM_CHAIN_ID OSMO DYDX)
osmo_to_dydx_connection=$(GET_CONNECTION_ID_FROM_CLIENT_ID OSMO $osmo_to_dydx_client)
osmo_to_dydx_channel=$(GET_TRANSFER_CHANNEL_ID_FROM_CONNECTION_ID OSMO $osmo_to_dydx_connection)

osmo_to_noble_client=$(GET_CLIENT_ID_FROM_CHAIN_ID OSMO NOBLE)
osmo_to_noble_connection=$(GET_CONNECTION_ID_FROM_CLIENT_ID OSMO $osmo_to_noble_client)
osmo_to_noble_channel=$(GET_TRANSFER_CHANNEL_ID_FROM_CONNECTION_ID OSMO $osmo_to_noble_connection)

echo -e "\nSTRIDE -> NOBLE:"
echo "  Client: $stride_to_noble_client"
echo "  Connection: $stride_to_noble_connection"

echo -e "\nSTRIDE -> OSMO:"
echo "  Client: $stride_to_osmo_client"
echo "  Connection: $stride_to_osmo_connection"

echo -e "\nDYDX -> NOBLE:"
echo "  Client: $dydx_to_noble_client"
echo "  Connection: $dydx_to_noble_connection"
echo "  Transfer Channel: $dydx_to_noble_channel -> $noble_to_dydx_channel"

echo -e "\nNOBLE -> OSMO:"
echo "  Client: $noble_to_osmo_client"
echo "  Connection: $noble_to_osmo_connection"
echo "  Transfer Channel: $noble_to_osmo_channel"

echo -e "\nOSMO -> DYDX:"
echo "  Client: $osmo_to_dydx_client"
echo "  Connection: $osmo_to_dydx_connection"
echo "  Transfer Channel: $osmo_to_dydx_channel"

echo -e "\nTransferring usdc to dydx to create ibc denom..."
$NOBLE_MAIN_CMD tx ibc-transfer transfer transfer $noble_to_dydx_channel $(DYDX_ADDRESS) 10000${USDC_DENOM} \
    --from ${NOBLE_VAL_PREFIX}1 -y | TRIM_TX
sleep 15

echo -e "\nDetermining IBC Denoms..."
usdc_denom_on_dydx=$(GET_IBC_DENOM DYDX $dydx_to_noble_channel $USDC_DENOM)
usdc_denom_on_osmo=$(GET_IBC_DENOM OSMO $osmo_to_noble_channel $USDC_DENOM)
dydx_denom_on_osmo=$(GET_IBC_DENOM OSMO $osmo_to_dydx_channel $DYDX_DENOM)

echo "  ibc/usdc on dYdX:    $usdc_denom_on_dydx"
echo "  ibc/usdc on Osmosis: $usdc_denom_on_osmo"
echo "  ibc/dydx on Osmosis: $dydx_denom_on_osmo"

proposal_file=${STATE}/${STRIDE_NODE_PREFIX}1/trade_route.json
cat << EOF > $proposal_file
{
  "title": "Create a new trade route for host chain X",
  "metadata": "Create a new trade route for host chain X",
  "summary": "Create a new trade route for host chain X",
  "messages": [
    {
      "@type": "/stride.stakeibc.MsgCreateTradeRoute",
      "authority": "stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl",
      "host_chain_id": "$DYDX_CHAIN_ID",
      "stride_to_reward_connection_id": "$stride_to_noble_connection",
      "stride_to_trade_connection_id": "$stride_to_osmo_connection",
      "host_to_reward_transfer_channel_id": "$dydx_to_noble_channel",
      "reward_to_trade_transfer_channel_id": "$noble_to_osmo_channel",
      "trade_to_host_transfer_channel_id": "$osmo_to_dydx_channel",
      "reward_denom_on_host": "$usdc_denom_on_dydx",
      "reward_denom_on_reward": "$USDC_DENOM",
      "reward_denom_on_trade": "$usdc_denom_on_osmo",
      "host_denom_on_trade": "$dydx_denom_on_osmo",
      "host_denom_on_host": "$DYDX_DENOM",
      "pool_id": 1,
      "max_allowed_swap_loss_rate": "0.05"
    }
  ],
  "deposit": "2000000000ustrd"
}
EOF

echo -e "\nCreate trade route proposal file:"
cat $proposal_file

echo -e "\n>>> Submitting proposal to register trade route..."
$STRIDE_MAIN_CMD tx gov submit-proposal $proposal_file --from val1 -y $GAS | TRIM_TX
sleep 5

echo -e "\n>>> Voting on proposal..."
proposal_id=$(GET_LATEST_PROPOSAL_ID STRIDE)
$STRIDE_MAIN_CMD tx gov vote $proposal_id yes --from val1 -y | TRIM_TX
$STRIDE_MAIN_CMD tx gov vote $proposal_id yes --from val2 -y | TRIM_TX
$STRIDE_MAIN_CMD tx gov vote $proposal_id yes --from val3 -y | TRIM_TX

echo -e "\nProposal Status:"
WATCH_PROPOSAL_STATUS STRIDE $proposal_id
