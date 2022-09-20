* Build
```
make build-docker build=sr
```
* Start stride
```
bash scripts/start_network.sh
```
* Fund relayer STARS address
* Create connections and channels
```
docker-compose run --rm relayer-stars rly transact link stride-stars > scripts/logs/relayer-stars.log 2>&1
```
* Get channel ID created on the host
```
build/strided --home scripts/state/stride1 q ibc channel channels 
```
* Start Relayer
```
docker-compose up -d relayer-stars
docker-compose logs -f relayer-stars | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> scripts/logs/relayer-stars.log 2>&1 &
```
* Create stars account
* IBC Transfer tokens from STARS
```
build/starsd keys add stars1 --recover --keyring-backend test 
build/starsd tx ibc-transfer transfer transfer {CHANNEL-ID} stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8 100000ustars --from stars1 --chain-id stargaze-1 -y --keyring-backend test --node https://stargaze-rpc.polkachu.com:443
build/starsd q tx {TX_HASH} --node https://stargaze-rpc.polkachu.com:443
```
* Confirm funds were recieved on stride and get IBC denom
```
build/strided --home scripts/state/stride1 q bank balances stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8
```
* Register host zone
```
build/strided --home scripts/state/stride1 tx stakeibc register-host-zone \
    connection-0 {HOST_DENOM} {ADDRESS_PREFIX} {IBC_DENOM} channel-0 1 \
    --from admin --gas 1000000 -y
```
* Add validator
```
build/strided --home scripts/state/stride1 tx stakeibc add-validator stargaze-1 imperator starsvaloper1y3cxrze7kmktj93atd42g9rffyg823g0qjqelc 10 5 --chain-id local-test-3 --keyring-backend test --from admin -y
```
* Liquid stake!
```
build/strided --home scripts/state/stride1 tx stakeibc liquid-stake 10000 ustars --keyring-backend test --from admin -y --chain-id local-test-3 -y
```
* Redeem!!
```
build/strided --home scripts/state/stride1 tx stakeibc redeem-stake 89 stargaze-1 stars1kwll0uet4mkj867s4q8dgskp03txgjnse0aa2l --from admin --keyring-backend test --chain-id local-test-3 -y
```