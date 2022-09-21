* Build
```
make build-docker build=srh
```
* Start stride
```
bash scripts/start_network.sh
```
* Fund go and hermes relayer STARS address
* Create connections and channels
```
docker-compose run --rm relayer-stars rly transact link stride-stars > scripts/logs/relayer-stars.log 2>&1
```
* Get channel ID created on the host
```
build/strided --home scripts/state/stride1 q ibc channel channels 
```
* Start Hermes Relayer
```
docker-compose up -d hermes
docker-compose logs -f hermes | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> scripts/logs/hermes-stars.log 2>&1 &
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
docker-compose up -d relayer-stars
docker-compose logs -f relayer-stars | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> scripts/logs/relayer-stars.log 2>&1 &
```
* IBC Transfer STARS to stride (from relayer account)
```
build/starsd keys add stars1 --recover --keyring-backend test 
build/starsd tx ibc-transfer transfer transfer {CHANNEL-ID} stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8 300000000ustars --from stars1 --chain-id stargaze-1 -y --keyring-backend test --node https://stargaze-rpc.polkachu.com:443
build/starsd q tx {TX_HASH} --node https://stargaze-rpc.polkachu.com:443
```
* Confirm funds were recieved on stride and get IBC denom
```
build/strided --home scripts/state/stride1 q bank balances stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8
```
* Register host zone
```
build/strided --home scripts/state/stride1 tx stakeibc register-host-zone \
    connection-0 ustars stars ibc/49BAE4CD2172833F14000627DA87ED8024AD46A38D6ED33F6239F22B5832F958 channel-0 1 \
    --from admin --gas 1000000 -y
```
* Add validator
```
build/strided --home scripts/state/stride1 tx stakeibc add-validator stargaze-1 imperator starsvaloper1y3cxrze7kmktj93atd42g9rffyg823g0qjqelc 10 5 --chain-id local-test-10 --keyring-backend test --from admin -y
```
* Liquid stake (then wait and LS again)
```
build/strided --home scripts/state/stride1 tx stakeibc liquid-stake 50000000 ustars --keyring-backend test --from admin -y --chain-id local-test-10 -y
```
* Confirm stTokens, StakedBal, and Redemption Rate
```
build/strided --home scripts/state/stride1 q bank balances stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8
build/strided --home scripts/state/stride1 q stakeibc list-host-zone
```
* Redeem
```
build/strided --home scripts/state/stride1 tx stakeibc redeem-stake 1000 stargaze-1 stars1kwll0uet4mkj867s4q8dgskp03txgjnse0aa2l --from admin --keyring-backend test --chain-id local-test-10 -y
```
* Confirm stTokens and StakedBal
```
build/strided --home scripts/state/stride1 q bank balances stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8
build/strided --home scripts/state/stride1 q stakeibc list-host-zone
```
* Add another validator
```
build/strided --home scripts/state/stride1 tx stakeibc add-validator stargaze-1 figment starsvaloper13htkxk8nw6qwhfdugllp8ldtgt5nm80xf679h5 10 5 --chain-id local-test-10 --keyring-backend test --from admin -y
```
* Liquid stake and confirm the stake was split 50/50 between the validators
```
build/strided --home scripts/state/stride1 tx stakeibc liquid-stake 50000000 ustars --keyring-backend test --from admin -y --chain-id local-test-10 -y
```
* Change validator weights
```
build/strided --home scripts/state/stride1 tx stakeibc change-validator-weight stargaze-1 starsvaloper1y3cxrze7kmktj93atd42g9rffyg823g0qjqelc 1 --from admin -y
build/strided --home scripts/state/stride1 tx stakeibc change-validator-weight stargaze-1 starsvaloper13htkxk8nw6qwhfdugllp8ldtgt5nm80xf679h5 49 --from admin -y
```
* LS and confirm delegation aligned with new weights
```
build/strided --home scripts/state/stride1 tx stakeibc liquid-stake 50000000 ustars --keyring-backend test --from admin -y --chain-id local-test-10 -y
```
* Call rebalance to and confirm new delegations
```
build/strided --home scripts/state/stride1 tx stakeibc rebalance-validators stargaze-1 5 --from admin
```
* Clear balances
```
build/strided --home scripts/state/stride1 tx stakeibc clear-balance stargaze-1 1 {CHANNEL-ID} --from admin
```
* Update balances 
```
```
