# Connecting to Mainnet
## Setup Scripts
* Update stride and host chain IDs in vars.sh and set num nodes to 1
* In `init_chain.sh`
    * Comment line that sets the unbonding time
    * Get rid of if branches on CHAIN_ID and keep just the stride logic
* Update go relayer
    * Update stride and host chain IDs (in chains and paths sections)
    * Set stride gas price to 0
    * Update endpoints for host
* Update hermes 
    * Updated chain IDs
    * Comment out other hosts
    * Update the following:
```
[[chains]]
id = '{CHAIN_ID}'
rpc_addr = 'https://{ENDPOINT}'
grpc_addr = 'http://{ENDPOINT}:9090'
websocket_addr = 'ws:///{ENDPOINT}/websocket'
max_tx_size = 200000
trusting_period = '12days'
[chains.packet_filter]
policy = 'allow'
list = [
  ['ica*', '*'],
  ['transfer', 'channel-*'],
]
```
* Comment out all but juno in hermes and the go relayer
## Start Stride Local
* Build
```
make build-docker build=srh
```
* Start stride
```
bash scripts/start_local_to_main.sh
```
## Create channels and start relayers
* Fund go and hermes relayer addresses on host
* Create connections and channels
```
docker-compose run --rm relayer-juno rly transact link stride-juno > scripts/logs/relayer-juno.log 2>&1
# Or if it's not working with go, use hermes
docker-compose run --rm hermes hermes create connection --a-chain local-test-1 --b-chain juno-1
docker-compose run --rm hermes hermes create channel --a-chain juno-1 --a-connection {CONNECTION-ID} --a-port transfer --b-port transfer
```
* Get channel ID created on the host
```
build/strided --home scripts/state/stride1 q ibc channel channels 
```
* Start Hermes Relayer
```
docker-compose up -d hermes
docker-compose logs -f hermes | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> scripts/logs/hermes-juno.log 2>&1 &
```
* Configure the Go Relayer to only run ICQ
```
# add this to scripts/state/relayer/config/config.yaml
src-channel-filter:
    rule: allowlist
    channel-list: []
```
* Start Go Relayer (for ICQ)
```
docker-compose up -d relayer-juno
docker-compose logs -f relayer-juno | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> scripts/logs/relayer-juno.log 2>&1 &
```
## Register Host
* IBC Transfer from HOST to stride (from relayer account)
```
# use one of the relayer mnemonics that has a juno balance
build/junod keys add juno1 --recover --keyring-backend test 
build/junod tx ibc-transfer transfer transfer {CHANNEL-ID} stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8 300000000ujuno --from juno1 --chain-id juno-1 -y --keyring-backend test --node {JUNO_RPC}
build/junod q tx {TX_HASH} --node {JUNO_RPC}
```
* Confirm funds were recieved on stride and get IBC denom
```
build/strided --home scripts/state/stride1 q bank balances stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8
```
* Register host zone
```
build/strided --home scripts/state/stride1 tx stakeibc register-host-zone \
    connection-0 ujuno juno {IBC-DENOM} channel-0 1 \
    --from admin --gas 1000000 -y
```
* Add validator
```
build/strided --home scripts/state/stride1 tx stakeibc add-validator juno-1 imperator junovaloper17n3w6v5q3n0tws4xv8upd9ul4qqes0nlg7q0xd 10 5 --chain-id local-test-1 --keyring-backend test --from admin -y
```
## Go Through Flow
* Liquid stake (then wait and LS again)
```
build/strided --home scripts/state/stride1 tx stakeibc liquid-stake 50000000 ujuno --keyring-backend test --from admin -y --chain-id local-test-1 -y
```
* Confirm stTokens, StakedBal, and Redemption Rate
```
build/strided --home scripts/state/stride1 q bank balances stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8
build/strided --home scripts/state/stride1 q stakeibc list-host-zone
```
* Redeem
```
build/strided --home scripts/state/stride1 tx stakeibc redeem-stake 1000 juno-1 {HOT_WALLET_ADDRESS} --from admin --keyring-backend test --chain-id local-test-1 -y
```
* Confirm stTokens and StakedBal
```
build/strided --home scripts/state/stride1 q bank balances stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8
build/strided --home scripts/state/stride1 q stakeibc list-host-zone
```
* Add another validator
```
build/strided --home scripts/state/stride1 tx stakeibc add-validator juno-1 cosmostation junovaloper1t8ehvswxjfn3ejzkjtntcyrqwvmvuknzmvtaaa 10 5 --chain-id local-test-1 --keyring-backend test --from admin -y
```
* Liquid stake and confirm the stake was split 50/50 between the validators
```
build/strided --home scripts/state/stride1 tx stakeibc liquid-stake 50000000 ujuno --keyring-backend test --from admin -y --chain-id local-test-1 -y
```
* Change validator weights
```
build/strided --home scripts/state/stride1 tx stakeibc change-validator-weight juno-1 junovaloper17n3w6v5q3n0tws4xv8upd9ul4qqes0nlg7q0xd 1 --from admin -y
build/strided --home scripts/state/stride1 tx stakeibc change-validator-weight juno-1 junovaloper1t8ehvswxjfn3ejzkjtntcyrqwvmvuknzmvtaaa 49 --from admin -y
```
* LS and confirm delegation aligned with new weights
```
build/strided --home scripts/state/stride1 tx stakeibc liquid-stake 50000000 ujuno --keyring-backend test --from admin -y --chain-id local-test-1 -y
```
* Call rebalance to and confirm new delegations
```
build/strided --home scripts/state/stride1 tx stakeibc rebalance-validators juno-1 5 --from admin
```
* Clear balances
```
build/strided --home scripts/state/stride1 tx stakeibc clear-balance juno-1 1 {CHANNEL-ID} --from admin
```
* Update balances 
```
```