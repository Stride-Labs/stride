#!/bin/bash
set -eux

SOVEREIGN_HOME="$HOME/.sovereign"
CONSUMER_HOME="$HOME/.consumer"
CONSUMER_HOME1="$HOME/.consumer1"
PROVIDER_CHAIN_ID="provider"
CONSUMER_CHAIN_ID="consumer"
MONIKER="consumer"
VALIDATOR="validator"
VALIDATOR1="validator1"
KEYRING="--keyring-backend test"
TX_FLAGS="--gas-adjustment 100 --gas auto"
PROVIDER_BINARY="gaiad"
SOVEREIGN_BINARY="strided-sd"
CONSUMER_BINARY="strided-icdd"
NODE_IP="localhost"
PROVIDER_RPC_LADDR="$NODE_IP:26658"
PROVIDER_GRPC_ADDR="$NODE_IP:9091"
PROVIDER_RPC_LADDR1="$NODE_IP:26668"
PROVIDER_GRPC_ADDR1="$NODE_IP:9101"
SOVEREIGN_RPC_LADDR="$NODE_IP:26648"
SOVEREIGN_GRPC_ADDR="$NODE_IP:9081"
CONSUMER_RPC_LADDR="$NODE_IP:26638"
CONSUMER_GRPC_ADDR="$NODE_IP:9071"
CONSUMER_RPC_LADDR1="$NODE_IP:26628"
CONSUMER_GRPC_ADDR1="$NODE_IP:9061"
CONSUMER_USER="consumer"
SOVEREIGN_VALIDATOR="sovereign_validator"
PROVIDER_VALIDATOR="provider_validator"
PROVIDER_HOME="$HOME/.provider"
PROVIDER_HOME1="$HOME/.provider1"
PROVIDER_NODE_ADDRESS="tcp://localhost:26658"
PROVIDER_DELEGATOR=delegator


# Clean start
killall $SOVEREIGN_BINARY &> /dev/null || true
killall $CONSUMER_BINARY &> /dev/null || true
rm -rf $CONSUMER_HOME
rm -rf $CONSUMER_HOME1
rm -rf $SOVEREIGN_HOME

################SOVEREIGN############################
$SOVEREIGN_BINARY init --chain-id $CONSUMER_CHAIN_ID $MONIKER --home $SOVEREIGN_HOME
sleep 1

# Build genesis file and node directory structure
$SOVEREIGN_BINARY init --chain-id $CONSUMER_CHAIN_ID $MONIKER --home $CONSUMER_HOME
sleep 1


# Create user account keypair
$SOVEREIGN_BINARY keys add $CONSUMER_USER $KEYRING --home $SOVEREIGN_HOME --output json > $SOVEREIGN_HOME/consumer_keypair.json 2>&1
$SOVEREIGN_BINARY keys add $SOVEREIGN_VALIDATOR $KEYRING --home $SOVEREIGN_HOME --output json > $SOVEREIGN_HOME/sovereign_validator_keypair.json 2>&1
$SOVEREIGN_BINARY keys add $PROVIDER_VALIDATOR $KEYRING --home $CONSUMER_HOME --output json > $CONSUMER_HOME/consumer_validator_keypair.json 2>&1

# Add account in genesis (required by Hermes)
$SOVEREIGN_BINARY add-genesis-account $(jq -r .address $SOVEREIGN_HOME/consumer_keypair.json) 1000000000stake --home $SOVEREIGN_HOME
$SOVEREIGN_BINARY add-genesis-account $(jq -r .address $SOVEREIGN_HOME/sovereign_validator_keypair.json) 1000000000000stake --home $SOVEREIGN_HOME
$SOVEREIGN_BINARY add-genesis-account $(jq -r .address $CONSUMER_HOME/consumer_validator_keypair.json) 1000000000000stake --home $SOVEREIGN_HOME

# generate genesis for sovereign chain
$SOVEREIGN_BINARY gentx $SOVEREIGN_VALIDATOR 11000000000stake $KEYRING --chain-id=$CONSUMER_CHAIN_ID --home $SOVEREIGN_HOME

################CONSUMER############################

# Copy validator key files
cp $PROVIDER_HOME/config/priv_validator_key.json $CONSUMER_HOME/config/priv_validator_key.json
cp $PROVIDER_HOME/config/node_key.json $CONSUMER_HOME/config/node_key.json

$SOVEREIGN_BINARY add-genesis-account $(jq -r .address $CONSUMER_HOME/consumer_validator_keypair.json) 1000000000000stake --home $CONSUMER_HOME
$SOVEREIGN_BINARY gentx $PROVIDER_VALIDATOR 12000000000stake $KEYRING --chain-id=$CONSUMER_CHAIN_ID --home $CONSUMER_HOME

cp $CONSUMER_HOME/config/gentx/* $SOVEREIGN_HOME/config/gentx

#copy genesis
$SOVEREIGN_BINARY collect-gentxs --home $SOVEREIGN_HOME
sed -i '' 's/"voting_period": "172800s"/"voting_period": "20s"/g' $SOVEREIGN_HOME/config/genesis.json

cp $SOVEREIGN_HOME/config/genesis.json $CONSUMER_HOME/config/genesis.json

#######CHAIN2#######
$SOVEREIGN_BINARY init --chain-id $CONSUMER_CHAIN_ID $MONIKER --home $CONSUMER_HOME1
sleep 1
#copy genesis
cp $SOVEREIGN_HOME/config/genesis.json $CONSUMER_HOME1/config/genesis.json

# Copy validator key files
cp $PROVIDER_HOME1/config/priv_validator_key.json $CONSUMER_HOME1/config/priv_validator_key.json
cp $PROVIDER_HOME1/config/node_key.json $CONSUMER_HOME1/config/node_key.json

##########SET CONFIG.TOML#####################
# Set default client port
sed -i -r "/node =/ s/= .*/= \"tcp:\/\/${SOVEREIGN_RPC_LADDR}\"/" $SOVEREIGN_HOME/config/client.toml
sed -i -r "/node =/ s/= .*/= \"tcp:\/\/${CONSUMER_RPC_LADDR}\"/" $CONSUMER_HOME/config/client.toml
sed -i -r "/node =/ s/= .*/= \"tcp:\/\/${CONSUMER_RPC_LADDR1}\"/" $CONSUMER_HOME1/config/client.toml
sovereign=$($SOVEREIGN_BINARY tendermint show-node-id --home $SOVEREIGN_HOME)
node=$($SOVEREIGN_BINARY tendermint show-node-id --home $CONSUMER_HOME)
node1=$($SOVEREIGN_BINARY tendermint show-node-id --home $CONSUMER_HOME1)

# sed -i -r "/persistent_peers =/ s/= .*/= \"$node@localhost:26636,$node1@localhost:26626\"/" "$SOVEREIGN_HOME"/config/config.toml
# sed -i -r "/persistent_peers =/ s/= .*/= \"$sovereign@localhost:26646,$node1@localhost:26626\"/" "$CONSUMER_HOME"/config/config.toml
# sed -i -r "/persistent_peers =/ s/= .*/= \"$sovereign@localhost:26646,$node@localhost:26636\"/" "$CONSUMER_HOME1"/config/config.toml

sed -i -r "/persistent_peers =/ s/= .*/= \"$node@localhost:26636\"/" "$SOVEREIGN_HOME"/config/config.toml
sed -i -r "/pprof_laddr =/ s/= .*/= \"localhost:6061\"/" "$SOVEREIGN_HOME"/config/config.toml
sed -i -r "/persistent_peers =/ s/= .*/= \"$sovereign@localhost:26646\"/" "$CONSUMER_HOME"/config/config.toml
sed -i -r "/pprof_laddr =/ s/= .*/= \"localhost:6062\"/" "$CONSUMER_HOME"/config/config.toml

sed -i -r "126s/.*/address = \"tcp:\/\/0.0.0.0:1318\"/" "$CONSUMER_HOME"/config/app.toml
sed -i -r "126s/.*/address = \"tcp:\/\/0.0.0.0:1319\"/" "$CONSUMER_HOME1"/config/app.toml

# Start the chain
$SOVEREIGN_BINARY start \
       --home $SOVEREIGN_HOME \
       --rpc.laddr tcp://${SOVEREIGN_RPC_LADDR} \
       --grpc.address ${SOVEREIGN_GRPC_ADDR} \
       --address tcp://${NODE_IP}:26645 \
       --p2p.laddr tcp://${NODE_IP}:26646 \
       --grpc-web.enable=false \
       --pruning=nothing \
       --log_level debug \
       --trace \
       &> $SOVEREIGN_HOME/logs &

$SOVEREIGN_BINARY start \
       --home $CONSUMER_HOME \
       --rpc.laddr tcp://${CONSUMER_RPC_LADDR} \
       --grpc.address ${CONSUMER_GRPC_ADDR} \
       --address tcp://${NODE_IP}:26635 \
       --p2p.laddr tcp://${NODE_IP}:26636 \
       --grpc-web.enable=false \
       --pruning=nothing \
       --log_level debug \
       --trace \
       &> $CONSUMER_HOME/logs &

# $SOVEREIGN_BINARY start \
#        --home $CONSUMER_HOME1 \
#        --rpc.laddr tcp://${CONSUMER_RPC_LADDR1} \
#        --grpc.address ${CONSUMER_GRPC_ADDR1} \
#        --address tcp://${NODE_IP}:26625 \
#        --p2p.laddr tcp://${NODE_IP}:26626 \
#        --grpc-web.enable=false \
#        --log_level debug \
#        --trace \
#        &> $CONSUMER_HOME1/logs &        
sleep 10


########## GO RELAYER ########
RELAYER_DIR="./relayer"
MNEMONIC_1="trip ten ability cabbage artefact side brass field domain doll ritual easily"
MNEMONIC_2="guard lion kiss upper comic vital small bundle salon oxygen museum material enter idea high domain lend alert dish message federal egg upper coffee"

# # send tokens to relayers
$PROVIDER_BINARY tx bank send $PROVIDER_DELEGATOR cosmos1d35ujw5e509acpmfxf9tw859e4nkhq8qfs725a 1000000stake --chain-id=$PROVIDER_CHAIN_ID --keyring-backend=test -y --broadcast-mode=sync --node=http://$PROVIDER_RPC_LADDR --home=$PROVIDER_HOME
sleep 5
$SOVEREIGN_BINARY tx bank send $SOVEREIGN_VALIDATOR stride1qwhddffw7ljzmgn9cxcd76t40aq0e65p5fwafv 1000000stake --keyring-backend=test --chain-id=$CONSUMER_CHAIN_ID -y --broadcast-mode=sync --node=http://$SOVEREIGN_RPC_LADDR --home=$SOVEREIGN_HOME
sleep 5

# PROVIDER_CHAIN_ID="provider"
# CONSUMER_CHAIN_ID="consumer"

killall rly 2> /dev/null || true

# echo "Restoring accounts..."
rly keys restore provider rly1 "$MNEMONIC_1" --home $RELAYER_DIR &
rly keys restore consumer rly2 "$MNEMONIC_2" --home $RELAYER_DIR &

sleep 3

rly transact link consumer-provider --home $RELAYER_DIR
sleep 2
rly start --home $RELAYER_DIR &> $RELAYER_DIR/logs &

# Build consumer chain proposal file - unbonding period 21 days
tee $PROVIDER_HOME/consumer-proposal.json<<EOF
{
    "title": "Create a chain",
    "description": "Gonna be a great chain",
    "chain_id": "consumer",
    "initial_height": {
        "revision_number": 0,
        "revision_height": 39
    },
    "genesis_hash": "519df96a862c30f53e67b1277e6834ab4bd59dfdd08c781d1b7cf3813080fb28",
    "binary_hash": "09184916f3e85aa6fa24d3c12f1e5465af2214f13db265a52fa9f4617146dea5",
    "spawn_time": "2022-06-01T09:10:00.000000000-00:00", 
    "deposit": "10000001stake",
    "consumer_redistribution_fraction": "0.75",
    "blocks_per_distribution_transmission": 1000,
    "ccv_timeout_period": 2419200000000000,
    "transfer_timeout_period": 3600000000000,
    "historical_entries": 10000,
    "unbonding_period": 1814400000000000
}
EOF

$PROVIDER_BINARY tx gov submit-proposal consumer-addition $PROVIDER_HOME/consumer-proposal.json \
	--gas=100000000 --chain-id $PROVIDER_CHAIN_ID --node tcp://$PROVIDER_RPC_LADDR --from $VALIDATOR --home $PROVIDER_HOME --keyring-backend test -b block -y
sleep 5

# Vote yes to proposal
$PROVIDER_BINARY query gov proposals --node tcp://$PROVIDER_RPC_LADDR
$PROVIDER_BINARY tx gov vote 1 yes --from $VALIDATOR --chain-id $PROVIDER_CHAIN_ID --node tcp://$PROVIDER_RPC_LADDR --home $PROVIDER_HOME -b block -y --keyring-backend test
sleep 5

# ###########################UPGRADE TO SOVEREIGN CHAIN##########################

# $SOVEREIGN_BINARY tx gov submit-legacy-proposal software-upgrade "v12" --upgrade-height=36  \
# --upgrade-info="upgrade for consumer chain" \
# --title="upgrade to consumer chain" --description="upgrade to consumer chain description" \
# --from=$SOVEREIGN_VALIDATOR $KEYRING --chain-id=$CONSUMER_CHAIN_ID \
# --home=$SOVEREIGN_HOME --yes -b sync --deposit="100000000stake"

$SOVEREIGN_BINARY tx gov submit-legacy-proposal software-upgrade "v12" \
--upgrade-height=36 \
--upgrade-info="file:///Users/admin/Documents/j121/stride/interchain-security/upgrade-info.json?checksum=sha256:66a0bd2ce260927747e2f26522b937d773d13b79c7e5b07e2f5efa483d6fabe0" \
--title="upgrade to consumer chain" --description="upgrade to consumer chain description" \
--from=$SOVEREIGN_VALIDATOR $KEYRING --chain-id=$CONSUMER_CHAIN_ID --home=$SOVEREIGN_HOME --yes -b sync --deposit="100000000stake"

sleep 5

# $SOVEREIGN_BINARY query tx 1BA75A3E36276AFC530171A29D0D080A39834B85AA340B5EC18002C4608FBE7F --home=$SOVEREIGN_HOME
# $SOVEREIGN_BINARY tx gov submit-legacy-proposal software-upgrade "v12" \
# --upgrade-height=36 \
# --upgrade-info="file:///Users/admin/Documents/j121/stride/interchain-security/upgrade-info1.json?checksum=sha256:3f0309a1179afb7d7eb9090a13ce9fdca0ce46b39ac6908b415d901b822f046a" \
# --title="upgrade to consumer chain" --description="upgrade to consumer chain description" \
# --from=$SOVEREIGN_VALIDATOR $KEYRING --chain-id=$CONSUMER_CHAIN_ID --home=$SOVEREIGN_HOME --yes -b sync --deposit="100000000stake"


# Vote yes to proposal
$SOVEREIGN_BINARY tx gov vote 1 yes --from $SOVEREIGN_VALIDATOR --chain-id $CONSUMER_CHAIN_ID --node tcp://$SOVEREIGN_RPC_LADDR \
--home $SOVEREIGN_HOME -b sync -y $KEYRING
sleep 80

# ###########################START BINARIES AGAIN AFTER UPGRADE##########################
# $SOVEREIGN_BINARY query gov proposals --node tcp://$SOVEREIGN_RPC_LADDR
# $SOVEREIGN_BINARY status --node tcp://$SOVEREIGN_RPC_LADDR
# $SOVEREIGN_BINARY status --node tcp://$CONSUMER_RPC_LADDR
# # $SOVEREIGN_BINARY status --node tcp://$CONSUMER_RPC_LADDR1

killall $SOVEREIGN_BINARY &> /dev/null || true

# Add ccv section to SOVEREIGN_HOME genesis to be used on upgrade handler
if ! $PROVIDER_BINARY q provider consumer-genesis "$CONSUMER_CHAIN_ID" --node "$PROVIDER_NODE_ADDRESS" --output json > "$SOVEREIGN_HOME"/consumer_section.json; 
then
       echo "Failed to get consumer genesis for the chain-id '$CONSUMER_CHAIN_ID'! Finalize genesis failed. For more details please check the log file in output directory."
       exit 1
fi

jq -s '.[0].app_state.ccvconsumer = .[1] | .[0]' "$SOVEREIGN_HOME"/config/genesis.json "$SOVEREIGN_HOME"/consumer_section.json > "$SOVEREIGN_HOME"/genesis_consumer.json && \
	mv "$SOVEREIGN_HOME"/genesis_consumer.json "$SOVEREIGN_HOME"/config/consumer-genesis.json

# Modify genesis params
jq ".app_state.ccvconsumer.params.blocks_per_distribution_transmission = \"70\" | .app_state.tokenfactory.paused = { \"paused\": false }" \
  $SOVEREIGN_HOME/config/consumer-genesis.json > \
   $SOVEREIGN_HOME/edited_genesis.json && mv $SOVEREIGN_HOME/edited_genesis.json $SOVEREIGN_HOME/config/consumer-genesis.json
sleep 1

cp $SOVEREIGN_HOME/config/consumer-genesis.json $CONSUMER_HOME/config/consumer-genesis.json

$CONSUMER_BINARY start \
       --home $SOVEREIGN_HOME \
       --rpc.laddr tcp://${SOVEREIGN_RPC_LADDR} \
       --grpc.address ${SOVEREIGN_GRPC_ADDR} \
       --address tcp://${NODE_IP}:26645 \
       --p2p.laddr tcp://${NODE_IP}:26646 \
       --grpc-web.enable=false \
       --log_level debug \
       --trace \
       &> $SOVEREIGN_HOME/logs &

$CONSUMER_BINARY start \
       --home $CONSUMER_HOME \
       --rpc.laddr tcp://${CONSUMER_RPC_LADDR} \
       --grpc.address ${CONSUMER_GRPC_ADDR} \
       --address tcp://${NODE_IP}:26635 \
       --p2p.laddr tcp://${NODE_IP}:26636 \
       --grpc-web.enable=false \
       --log_level debug \
       --trace \
       &> $CONSUMER_HOME/logs &

# $CONSUMER_BINARY start \
#        --home $CONSUMER_HOME1 \
#        --rpc.laddr tcp://${CONSUMER_RPC_LADDR1} \
#        --grpc.address ${CONSUMER_GRPC_ADDR1} \
#        --address tcp://${NODE_IP}:26625 \
#        --p2p.laddr tcp://${NODE_IP}:26626 \
#        --grpc-web.enable=false \
#        --log_level debug \
#        --trace \
#        &> $CONSUMER_HOME1/logs &        
sleep 30

rly transact connection consumer-provider-ics --home $RELAYER_DIR
rly transact channel consumer-provider-ics --src-port consumer --dst-port provider --order ordered --version 1 --home $RELAYER_DIR

# ############################################################

PROVIDER_VALIDATOR_ADDRESS=$(jq -r .address $PROVIDER_HOME/keypair.json)
DELEGATIONS=$($PROVIDER_BINARY q staking delegations $PROVIDER_VALIDATOR_ADDRESS --home $PROVIDER_HOME --node tcp://${PROVIDER_RPC_LADDR} -o json)
OPERATOR_ADDR=$(echo $DELEGATIONS | jq -r .delegation_responses[0].delegation.validator_address)

$PROVIDER_BINARY tx staking delegate $OPERATOR_ADDR 32000000stake \
       --from $VALIDATOR \
       $KEYRING \
       --home $PROVIDER_HOME \
       --node tcp://${PROVIDER_RPC_LADDR} \
       --chain-id $PROVIDER_CHAIN_ID -y -b block
sleep 1

$PROVIDER_BINARY status --node=tcp://${PROVIDER_RPC_LADDR}
# $PROVIDER_BINARY status --node=tcp://${PROVIDER_RPC_LADDR1}

$CONSUMER_BINARY status --node tcp://$SOVEREIGN_RPC_LADDR
$CONSUMER_BINARY status --node tcp://$CONSUMER_RPC_LADDR

# $CONSUMER_BINARY query staking params --node=tcp://$CONSUMER_RPC_LADDR
# $PROVIDER_BINARY query staking params --node=tcp://${PROVIDER_RPC_LADDR}