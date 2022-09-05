* Build
```
make build-local build=s
```
* Start stride
```
bash scripts-local/start_network.sh
```
* Create connections and channels
./build/hermes/release/hermes --config scripts-local/hermes/config.toml create connection --a-chain local-test-1 --b-chain cosmoshub-4
./build/hermes/release/hermes --config scripts-local/hermes/config.toml create channel --a-chain cosmoshub-4 --a-connection {CONNECTION-ID} --a-port transfer --b-port transfer
* Start hermes
```
 ./build/hermes/release/hermes --config scripts-local/hermes/config.toml start >> scripts-local/logs/hermes.log 2>&1 &
```
* Start ICQ
```
bash scripts-local/icq_startup.sh &
```
* Create admin account
```
echo $STRIDE_ADMIN_MNEMONIC | build/strided --home scripts-local/state/stride keys add admin --recover
```
* Create gaia account
* IBC Transfer tokens from GAIA
```
gaiad tx ibc-transfer transfer transfer channel-386 stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8 100000uatom --from fleet --chain-id cosmoshub-4 -y --keyring-backend test
```
* Register host zone
```
build/strided --home scripts-local/state/stride tx stakeibc register-host-zone \
    connection-0 uatom cosmos {IBC_DENOM} channel-0 1 \
    --from admin --gas 1000000 -y
```
* Add validator
```
build/strided --home scripts-local/state/stride tx stakeibc add-validator cosmoshub-4 imperator cosmosvaloper1vvwtk805lxehwle9l4yudmq6mn0g32px9xtkhc 10 5 --chain-id local-test-1 --keyring-backend test --from admin -y
```
* Liquid stake!
```
build/strided --home scripts-local/state/stride tx stakeibc liquid-stake 10000 uatom --keyring-backend test --from admin -y --chain-id local-test-1 -y
```
* Redeem!!
```
build/strided --home scripts-local/state/stride tx stakeibc redeem-stake 89 cosmoshub-4 cosmos1gvxswlup5ejmlwch9q00zv8ne28gq6vuzzw0j5 --from admin --keyring-backend test --chain-id local-test-1 -y
```