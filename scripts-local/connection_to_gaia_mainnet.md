* Build
```
make build-local build=s
```
* Start stride
```
bash scripts-local/start_network.sh
```
* Fund hermes and ICQ addresses on GAIA via keplr
* Create connections and channels
./build/hermes/release/hermes --config scripts-local/hermes/config.toml create connection --a-chain local-test-1 --b-chain cosmoshub-4
./build/hermes/release/hermes --config scripts-local/hermes/config.toml create channel --a-chain cosmoshub-4 --a-connection {CONNECTION-ID} --a-port transfer --b-port transfer
* Start hermes
```
 ./build/hermes/release/hermes --config scripts-local/hermes/config.toml start >> scripts-local/logs/hermes.log 2>&1 &
```
* Create admin account
```
echo $STRIDE_ADMIN_MNEMONIC | build/strided --home scripts-local/state/stride keys add admin --recover
```
* Transfer tokens from GAIA
* Register host zone
```
build/strided --home scripts-local/state/stride tx stakeibc register-host-zone \
    {CONNECTION} uatom cosmos {IBC_DENOM} {CHANNEL} 1 \
    --from admin --gas 1000000 -y
```