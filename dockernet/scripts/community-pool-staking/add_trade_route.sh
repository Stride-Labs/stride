#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

HOST_CHAIN=$REWARD_CONVERTER_HOST_ZONE
HOST_VAL_ADDRESS=$(${HOST_CHAIN}_ADDRESS)
HOST_CHAIN_ID=$(GET_VAR_VALUE ${HOST_CHAIN}_CHAIN_ID)
HOST_DENOM=$(GET_VAR_VALUE    ${HOST_CHAIN}_DENOM)

GAS="--gas-prices 0.1ustrd --gas auto --gas-adjustment 1.3"

echo "Determining relevant connections and channels..."
stride_to_noble_client=$(GET_CLIENT_ID_FROM_CHAIN_ID STRIDE NOBLE)
stride_to_noble_connection=$(GET_CONNECTION_ID_FROM_CLIENT_ID STRIDE $stride_to_noble_client)

stride_to_osmo_client=$(GET_CLIENT_ID_FROM_CHAIN_ID STRIDE OSMO)
stride_to_osmo_connection=$(GET_CONNECTION_ID_FROM_CLIENT_ID STRIDE $stride_to_osmo_client)

host_to_noble_client=$(GET_CLIENT_ID_FROM_CHAIN_ID $HOST_CHAIN NOBLE)
host_to_noble_connection=$(GET_CONNECTION_ID_FROM_CLIENT_ID $HOST_CHAIN $host_to_noble_client)
host_to_noble_channel=$(GET_TRANSFER_CHANNEL_ID_FROM_CONNECTION_ID $HOST_CHAIN $host_to_noble_connection)

host_to_osmo_client=$(GET_CLIENT_ID_FROM_CHAIN_ID $HOST_CHAIN OSMO)
host_to_osmo_connection=$(GET_CONNECTION_ID_FROM_CLIENT_ID $HOST_CHAIN $host_to_osmo_client)
host_to_osmo_channel=$(GET_TRANSFER_CHANNEL_ID_FROM_CONNECTION_ID $HOST_CHAIN $host_to_osmo_connection)

noble_to_host_client=$(GET_CLIENT_ID_FROM_CHAIN_ID NOBLE $HOST_CHAIN)
noble_to_host_connection=$(GET_CONNECTION_ID_FROM_CLIENT_ID NOBLE $noble_to_host_client)
noble_to_host_channel=$(GET_TRANSFER_CHANNEL_ID_FROM_CONNECTION_ID NOBLE $noble_to_host_connection)

noble_to_osmo_client=$(GET_CLIENT_ID_FROM_CHAIN_ID NOBLE OSMO)
noble_to_osmo_connection=$(GET_CONNECTION_ID_FROM_CLIENT_ID NOBLE $noble_to_osmo_client)
noble_to_osmo_channel=$(GET_TRANSFER_CHANNEL_ID_FROM_CONNECTION_ID NOBLE $noble_to_osmo_connection)

osmo_to_host_client=$(GET_CLIENT_ID_FROM_CHAIN_ID OSMO $HOST_CHAIN)
osmo_to_host_connection=$(GET_CONNECTION_ID_FROM_CLIENT_ID OSMO $osmo_to_host_client)
osmo_to_host_channel=$(GET_TRANSFER_CHANNEL_ID_FROM_CONNECTION_ID OSMO $osmo_to_host_connection)

osmo_to_noble_client=$(GET_CLIENT_ID_FROM_CHAIN_ID OSMO NOBLE)
osmo_to_noble_connection=$(GET_CONNECTION_ID_FROM_CLIENT_ID OSMO $osmo_to_noble_client)
osmo_to_noble_channel=$(GET_TRANSFER_CHANNEL_ID_FROM_CONNECTION_ID OSMO $osmo_to_noble_connection)

echo -e "\nSTRIDE -> NOBLE:"
echo "  Client: $stride_to_noble_client"
echo "  Connection: $stride_to_noble_connection"

echo -e "\nSTRIDE -> OSMO:"
echo "  Client: $stride_to_osmo_client"
echo "  Connection: $stride_to_osmo_connection"

echo -e "\n$HOST_CHAIN -> NOBLE:"
echo "  Client: $host_to_noble_client"
echo "  Connection: $host_to_noble_connection"
echo "  Transfer Channel: $host_to_noble_channel -> $noble_to_host_channel"

echo -e "\nNOBLE -> OSMO:"
echo "  Client: $noble_to_osmo_client"
echo "  Connection: $noble_to_osmo_connection"
echo "  Transfer Channel: $noble_to_osmo_channel"

echo -e "\nOSMO -> $HOST_CHAIN:"
echo "  Client: $osmo_to_host_client"
echo "  Connection: $osmo_to_host_connection"
echo "  Transfer Channel: $osmo_to_host_channel"

echo -e "\n$HOST_CHAIN -> OSMO:"
echo "  Client: $host_to_osmo_client"
echo "  Connection: $host_to_osmo_connection"
echo "  Transfer Channel: $host_to_osmo_channel"

echo -e "\nTransferring $USDC_DENOM to $HOST_CHAIN to create ibc denom..."
$NOBLE_MAIN_CMD tx ibc-transfer transfer transfer $noble_to_host_channel $HOST_VAL_ADDRESS 10000${USDC_DENOM} \
    --from ${NOBLE_VAL_PREFIX}1 -y | TRIM_TX
sleep 15

echo -e "\nTransferring $USDC_DENOM to OSMO to create ibc denom..."
$NOBLE_MAIN_CMD tx ibc-transfer transfer transfer $noble_to_osmo_channel $(OSMO_ADDRESS) 10000${USDC_DENOM} \
    --from ${NOBLE_VAL_PREFIX}1 -y | TRIM_TX
sleep 15

echo -e "\nTransferring $HOST_DENOM to OSMO to create ibc denom..."
$HOST_MAIN_CMD tx ibc-transfer transfer transfer $host_to_osmo_channel $(OSMO_ADDRESS) 10000${HOST_DENOM} \
    --from ${HOST_VAL_PREFIX}1 -y | TRIM_TX
sleep 15

echo -e "\nDetermining IBC Denoms..."
usdc_denom_on_host=$(GET_IBC_DENOM $HOST_CHAIN_ID $host_to_noble_channel $USDC_DENOM)
usdc_denom_on_osmo=$(GET_IBC_DENOM OSMO           $osmo_to_noble_channel $USDC_DENOM)
host_denom_on_osmo=$(GET_IBC_DENOM OSMO           $osmo_to_host_channel $HOST_DENOM)

echo "  ibc/$USDC_DENOM on Host:    $usdc_denom_on_host"
echo "  ibc/$USDC_DENOM on Osmosis: $usdc_denom_on_osmo"
echo "  ibc/$HOST_DENOM on Osmosis: $host_denom_on_osmo"

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
      "host_chain_id": "$HOST_CHAIN_ID",
      "stride_to_reward_connection_id": "$stride_to_noble_connection",
      "stride_to_trade_connection_id": "$stride_to_osmo_connection",
      "host_to_reward_transfer_channel_id": "$host_to_noble_channel",
      "reward_to_trade_transfer_channel_id": "$noble_to_osmo_channel",
      "trade_to_host_transfer_channel_id": "$osmo_to_host_channel",
      "reward_denom_on_host": "$usdc_denom_on_host",
      "reward_denom_on_reward": "$USDC_DENOM",
      "reward_denom_on_trade": "$usdc_denom_on_osmo",
      "host_denom_on_trade": "$host_denom_on_osmo",
      "host_denom_on_host": "$HOST_DENOM",
      "pool_id": 1,
      "max_allowed_swap_loss_rate": "0.15"
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