`/scripts` contains (unmaintained) infrastructure that was used for early testing and development of the Stride protocol. The scripts here support docker-image based testing, some of which are heavily inspired by those used by Osmosis and Quicksilver (although there have been deviations from the original implementations since). The relevant licenses are included here.

## Dockernet
### Adding a new host zone
* Create a new dockerfile at the root level (`Dockerfile.{new-host-zone})
* Add the repo as a submodule
```
git submodule add {repo-url} deps/{new-host-zone}
```
* Update the commit hash
```
cd deps/{new-host-zone}
git checkout {commit-hash}
cd ..
```
* Add a comment to `.gitmodules` with the commit hash
* Add the build command for that host zone in `scripts/build.sh`
```
while getopts sgojhir{n} flag; do
   case "${flag}" in
   ...
   n) build_local_and_docker {new-host-zone} deps/{new-host-zone} ;;  
```
* Add the host zone to the docker compose file at the root level. Add the port forwarding to the first node. Add 5 nodes here. Drop the RPC port number by 100, and the API/gRPC port by 10 since the last host zone.
```
  {new-host-zone}1:
    image: stridezone:{new-host-zone}
    volumes:
      - ./scripts/state/{new-host-zone}1:/home/{new-host-zone}/.{new-host-zone}
    ports:
      - "26257:26657"
      - "1277:1317"
      - "9050:9090"

  {new-host-zone}2:
    image: stridezone:{new-host-zone}
    volumes:
      - ./scripts/state/{new-host-zone}2:/home/{new-host-zone}/.{new-host-zone}

    ...

  {new-host-zone}5:
    image: stridezone:{new-host-zone}
    volumes:
      - ./scripts/state/{new-host-zone}5:/home/{new-host-zone}/.{new-host-zone}
```
* Add the following parameters to `scripts/vars.sh`, where `CHAIN_ID` is the ID of the new host zone
```
{CHAIN_ID}_CHAIN_ID={NEW-HOST-ZONE}
{CHAIN_ID}_NODE_PREFIX={new-host-zone}
{CHAIN_ID}_NUM_NODES=3
{CHAIN_ID}_CMD="$SCRIPT_DIR/../build/{new-host-zone}d"
{CHAIN_ID}_VAL_PREFIX={n}val
{CHAIN_ID}_ADDRESS_PREFIX=stars
{CHAIN_ID}_REV_ACCT={n}rev1
{CHAIN_ID}_DENOM={add denom as a constant at the top of the script and then reference here}
{CHAIN_ID}_RPC_PORT={the one included in the docker-compose above}
{CHAIN_ID}_MAIN_CMD="${CHAIN_ID}_CMD --home $SCRIPT_DIR/state/${${CHAIN_ID}_NODE_PREFIX}1"

{CHAIN_ID}_REV_MNEMONIC=""
{CHAIN_ID}_VAL_MNEMONIC_1=""
{CHAIN_ID}_VAL_MNEMONIC_2=""
{CHAIN_ID}_VAL_MNEMONIC_3=""
{CHAIN_ID}_VAL_MNEMONIC_4=""
{CHAIN_ID}_VAL_MNEMONIC_5=""
{CHAIN_ID}_VAL_MNEMONICS=("${CHAIN_ID}_VAL_MNEMONIC_1","${CHAIN_ID}_VAL_MNEMONIC_2","${CHAIN_ID}_VAL_MNEMONIC_3","${CHAIN_ID}_VAL_MNEMONIC_4","${CHAIN_ID}_VAL_MNEMONIC_5")

HERMES_${CHAIN_ID}_ACCT=rly{add one since the account from the last host zone}
HERMES_${CHAIN_ID}_MNEMONIC=""

RELAYER_{CHAIN_ID}_EXEC="docker-compose run --rm relayer-{new-host-zone}"
RELAYER_{CHAIN_ID}_ACCT=rly{add one since the account from the last host zone}
RELAYER_{CHAIN_ID}_MNEMONIC=""

```
* Add the IBC denom's for the host zone across each channel. You can use the following code block (just temporarily throw it in any of the test files and run it)
```
import transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"

func (s *KeeperTestSuite) TestIBCDenom() {
	chainId := {CHAIN_ID}
	denom := {DENOM}
	for i := 0; i < 4; i++ {
		sourcePrefix := transfertypes.GetDenomPrefix("transfer", fmt.Sprintf("channel-%d", i))
		prefixedDenom := sourcePrefix + denom

		fmt.Printf("IBC_%s_CHANNEL_%d_DENOM='%s'\n", chainId, i, transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom())
	}
}
```
* Add a section to the `scripts/config/relayer_config.yaml`
```
chains:
  ...
  {new-host-zone}:
    type: cosmos
    value:
      key: rly{N}
      chain-id: {CHAIN_ID}
      rpc-addr: http://{NODE_PREFIX}1:26657
      account-prefix: {ACCOUNT_PREFIX}
      keyring-backend: test
      gas-adjustment: 1.2
      gas-prices: 0.01{DENOM}
      debug: false
      timeout: 20s
      output-format: json
      sign-mode: direct
  ...
paths:
  ...
    stride-{new-host-zone}:
    src:
      chain-id: STRIDE
    dst:
      chain-id: {CHAIN_ID}
    src-channel-filter:
      rule: ""
      channel-list: []
```
* Add a section to hermes
```
[[chains]]
id = '{CHAIN_ID}'
rpc_addr = 'http://{NODE_PREFIX}1:26657'
grpc_addr = 'http://{NODE_PREFIX}1:9090'
websocket_addr = 'ws://{NODE_PREFIX}:26657/websocket'
rpc_timeout = '10s'
account_prefix = '{ADDRESS_PREFIX}'
key_name = 'hrly{next relayer ID}'
store_prefix = 'ibc'
default_gas = 100000
max_gas = 3000000
gas_price = { price = 0.000, denom = '{DENOM}' }
gas_multiplier = 1.1
max_msg_num = 30
max_tx_size = 2097152
clock_drift = '5s'
max_block_time = '10s'
trusting_period = '119s'
trust_threshold = { numerator = '1', denominator = '3' }
address_type = { derivation = 'cosmos' }
```
* Finally add the execution of the `init_chain` script for this host zone in `scripts/start_network.sh`, and add it to the array of `HOST_CHAINS`
```
sh ${SCRIPT_DIR}/init_chain.sh {NEW-HOST-ZONE}
HOST_CHAINS=(GAIA JUNO OSMO ... {NEW-HOST-ZONE})
```
* And that's it! Just start the network as normal, and make sure to rebuild the new host zone when running for the first time.  