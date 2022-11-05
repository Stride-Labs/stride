#### SETUP HOT WALLET (Only needs to be run once)
echo "$HOT_WALLET_1_MNEMONIC" | HOST_BINARY keys add hot --recover --keyring-backend test 


#### START RELAYERS
# Create connections and channels
docker-compose run --rm relayer-host rly transact link stride-host 

# (OR) If the go relayer isn't working, use hermes (you'll have to add the connections to the relayer config though in `scripts/state/relaye/config/config.yaml`)
# docker-compose run --rm hermes hermes create connection --a-chain HOST_CHAIN_ID --b-chain STRIDE_CHAIN_ID
# docker-compose run --rm hermes hermes create channel --a-chain STRIDE_CHAIN_ID --a-connection connection-0 --a-port transfer --b-port transfer

# Ensure Relayer Config is updated (`scripts/state/relaye/config/config.yaml`)
#    paths:
#     stride-host:
#       src:
#         chain-id: stride-1
#         client-id: 07-tendermint-0
#         connection-id: connection-0
#       dst:
#         chain-id: cosmoshub-4
#         client-id: {CLIENT-ID}
#         connection-id: {CONNECTION-ID}

# Get channel ID created on the host
build/strided --home scripts/state/stride1 q ibc channel channels 
transfer_channel=$(build/strided --home scripts/state/stride1 q ibc channel channels | grep channel-0 -A 4 | grep counterparty -A 1 | grep channel | awk '{print $2}') && echo $transfer_channel

# Start Hermes Relayer
docker-compose up -d hermes
docker-compose logs -f hermes | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> scripts/logs/hermes.log 2>&1 &

# Configure the Go Relayer to only run ICQ
sed -i -E "s|rule: \"\"|rule: allowlist|g" scripts/state/relayer/config/config.yaml

# Start Go Relayer (for ICQ)
docker-compose up -d relayer-host
docker-compose logs -f relayer-host | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> scripts/logs/relayer.log 2>&1 &


#### REGISTER HOST
# IBC Transfer from HOST to stride (from relayer account)
HOST_BINARY tx ibc-transfer transfer transfer $transfer_channel stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8 4000000HOST_DENOM --from hot --chain-id HOST_CHAIN_ID -y --keyring-backend test --node HOST_RPC_ENDPOINT

# Confirm funds were recieved on stride and get IBC denom
build/strided --home scripts/state/stride1 q bank balances stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8

# Register host zone
IBC_DENOM=$(build/strided --home scripts/state/stride1 q bank balances stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8 | grep ibc | awk '{print $2}' | tr -d '"') && echo $IBC_DENOM
build/strided --home scripts/state/stride1 tx stakeibc register-host-zone \
    connection-0 HOST_DENOM HOST_ACCOUNT_PREFIX $IBC_DENOM channel-0 1 \
    --from admin --gas 1000000 -y

# Add validator
build/strided --home scripts/state/stride1 tx stakeibc add-validator HOST_CHAIN_ID HOST_VAL_NAME_1 HOST_VAL_ADDRESS_1 10 5 --chain-id STRIDE_CHAIN_ID --keyring-backend test --from admin -y

# Confirm ICA channels were registered
build/strided --home scripts/state/stride1 q stakeibc list-host-zone

#### FLOW
## Go Through Flow
# Liquid stake (then wait and LS again)
build/strided --home scripts/state/stride1 tx stakeibc liquid-stake 1000000 HOST_DENOM --keyring-backend test --from admin -y --chain-id STRIDE_CHAIN_ID -y

# Confirm stTokens, StakedBal, and Redemption Rate
build/strided --home scripts/state/stride1 q bank balances stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8
build/strided --home scripts/state/stride1 q stakeibc list-host-zone

# Redeem
build/strided --home scripts/state/stride1 tx stakeibc redeem-stake 1000 HOST_CHAIN_ID HOT_WALLET_ADDRESS --from admin --keyring-backend test --chain-id STRIDE_CHAIN_ID -y

# Confirm stTokens and StakedBal
build/strided --home scripts/state/stride1 q bank balances stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8
build/strided --home scripts/state/stride1 q stakeibc list-host-zone

# Add another validator
build/strided --home scripts/state/stride1 tx stakeibc add-validator HOST_CHAIN_ID HOST_VAL_NAME_2 HOST_VAL_ADDRESS_2 10 5 --chain-id STRIDE_CHAIN_ID --keyring-backend test --from admin -y

# Liquid stake and confirm the stake was split 50/50 between the validators
build/strided --home scripts/state/stride1 tx stakeibc liquid-stake 1000000 HOST_DENOM --keyring-backend test --from admin -y --chain-id STRIDE_CHAIN_ID -y

# Change validator weights
build/strided --home scripts/state/stride1 tx stakeibc change-validator-weight HOST_CHAIN_ID HOST_VAL_ADDRESS_1 1 --from admin -y
build/strided --home scripts/state/stride1 tx stakeibc change-validator-weight HOST_CHAIN_ID HOST_VAL_ADDRESS_2 49 --from admin -y

# LS and confirm delegation aligned with new weights
build/strided --home scripts/state/stride1 tx stakeibc liquid-stake 1000000 HOST_DENOM --keyring-backend test --from admin -y --chain-id STRIDE_CHAIN_ID -y

# Call rebalance to and confirm new delegations
build/strided --home scripts/state/stride1 tx stakeibc rebalance-validators HOST_CHAIN_ID 5 --from admin

# Clear balances
fee_address=$(build/strided --home scripts/state/stride1 q stakeibc show-host-zone osmosis-1 | grep feeAccount -A 1 | grep address | awk '{print $2}') && echo $fee_address
balance=$(build/osmosisd --home scripts/state/stride1 q bank balances $fee_address | grep amount | awk '{print $3}' | tr -d '"') && echo $balance
build/strided --home scripts/state/stride1 tx stakeibc clear-balance HOST_CHAIN_ID $balance $transfer_channel --from admin

# Update delegations (just submit this query and confirm the ICQ callback displays in the stride logs)
# Must be submitted in ICQ window
build/strided --home scripts/state/stride1 tx stakeibc update-delegation HOST_CHAIN_ID HOST_VAL_ADDRESS_1 --from admin -y

#### MISC 
# If a channel closes, restore it with:
build/strided --home scripts/state/stride1 tx stakeibc restore-interchain-account HOST_CHAIN_ID {DELEGATION | WITHDRAWAL | FEE | REDEMPTION} --from admin