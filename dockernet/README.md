`/dockernet` contains (unmaintained) infrastructure that was used for early testing and development of the Stride protocol. The scripts here support docker-image based testing, some of which are heavily inspired by those used by Osmosis and Quicksilver (although there have been deviations from the original implementations since). The relevant licenses are included here.

## Dockernet
### Adding a new host zone
* Create a new dockerfile to `dockernet/dockerfiles` (named `Dockerfile.{new-host-zone}`). Use one of the other host zone's dockerfile's as a starting port to provide the certain boilerplate such as the package installs, adding user, exposing ports, etc. 
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
* Add the build command for that host zone in `dockernet/build.sh` (`n` is used as an example below - use the first letter of the host zone)
```
while getopts sgojhir{n} flag; do
   case "${flag}" in
   ...
   n) build_local_and_docker {new-host-zone} deps/{new-host-zone} ;;  
```
* Add the host zone to the docker compose file. Add 5 nodes and add the port forwarding to the first node only. Drop the RPC port number by 100, and the API/gRPC port by 10, relative to the last host zone that was added.
```
  {new-host-zone}1:
    image: stridezone:{new-host-zone}
    volumes:
      - ./dockernet/state/{new-host-zone}1:/home/{new-host-zone}/.{new-host-zone}
    ports:
      - "{rpc-port}:26657"
      - "{api-port}:1317"
      - "{grpc-port}:9090"

  {new-host-zone}2:
    image: stridezone:{new-host-zone}
    volumes:
      - ./dockernet/state/{new-host-zone}2:/home/{new-host-zone}/.{new-host-zone}

    ...

  {new-host-zone}5:
    image: stridezone:{new-host-zone}
    volumes:
      - ./dockernet/state/{new-host-zone}5:/home/{new-host-zone}/.{new-host-zone}
```
* Add the following parameters to `dockernet/config.sh`, where `CHAIN` is the ID of the new host zone
```
{CHAIN}_CHAIN_ID={NEW-HOST-ZONE}
{CHAIN}_NODE_PREFIX={new-host-zone}
{CHAIN}_NUM_NODES=3
{CHAIN}_CMD="$SCRIPT_DIR/../build/{new-host-zone}d"
{CHAIN}_VAL_PREFIX={n}val
{CHAIN}_ADDRESS_PREFIX=stars
{CHAIN}_REV_ACCT={n}rev1
{CHAIN}_DENOM={add denom as a constant at the top of the script and then reference here}
{CHAIN}_RPC_PORT={the one included in the docker-compose above}
{CHAIN}_MAIN_CMD="${CHAIN}_CMD --home $SCRIPT_DIR/state/${${CHAIN}_NODE_PREFIX}1"

RELAYER_{CHAIN}_EXEC="docker-compose run --rm relayer-{new-host-zone}"
RELAYER_{CHAIN}_ACCT=rly{add one since the account from the last host zone}
HOST_RELAYER_ACCTS=(... $RELAYER_{CHAIN}_ACCT)

RELAYER_{CHAIN}_MNEMONIC=""
RELAYER_MNEMONICS=(...,"$RELAYER_{CHAIN}_MNEMONIC")

```
* Add the IBC denom's for the host zone across each channel to `config.sh` (e.g. `IBC_{HOST}_CHANNEL_{N}_DENOM)`). You can use the following code block to generate the variables (just temporarily throw it in any of the test files, run it, and copy the output to `config.sh`)
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
* Add a section to the `dockernet/config/relayer_config.yaml`
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
* To enable the the new host zone, include it in the `HOST_CHAINS` array in `dockernet/config.sh`. **Note: You can only run up to 4 host zones at once.**
```
HOST_CHAINS=(GAIA {NEW-HOST-ZONE})
```
* And that's it! Just start the network as normal, and make sure to rebuild the new host zone when running for the first time.  
