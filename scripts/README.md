`/scripts` contains infrastructure that is used for testing and development of the Stride protocol. The scripts here support docker-image based testing, some of which are heavily inspired by those used by Osmosis and Quicksilver (although there have been large deviations from the original implementations since). The relevant licenses are included here.

## Dockernet
### Adding a new host zone
* Create a new dockerfile at the root level (`Dockerfile.{new-host-zone}`). Use one of the other host zone's dockerfile's as a starting port to provide the certain boilerplate such as the package installs, adding user, exposing ports, etc. You can often find a dockerfile in the github directory of the host zone. In the dockerfile, set `COMMIT_HASH` to the current mainnet commit hash of the chain being tested (or the target commit hash, if we're launching the zone in the future after an upgrade). For newer chains, create a branch and a pull-request, but *do not* merge it (we don't maintain test versions of each chain).
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
* Add the build command for that host zone in `scripts/build.sh`. For most zones, we use the first letter of the zone, for the new zone, just use `n` (since it won't be merged in, it won't conflict with anything).
```
while getopts sgojhir{n} flag; do
   case "${flag}" in
   ...
   n) build_local_and_docker {new-host-zone} deps/{new-host-zone} ;;  
```
* Add the host zone to `docker-compose.yml` in the root level of the stride directory. Add 5 nodes and add port forwarding to the first node only. Drop the RPC port number by 100, and the API/gRPC port by 10, relative to the last host zone that was added.
```
  {new-host-zone}1:
    image: stridezone:{new-host-zone}
    volumes:
      - ./scripts/state/{new-host-zone}1:/home/{new-host-zone}/.{new-host-zone}
    ports:
      - "{rpc-port}:26657"
      - "{api-port}:1317"
      - "{grpc-port}:9090"

  {new-host-zone}2:
    image: stridezone:{new-host-zone}
    volumes:
      - ./scripts/state/{new-host-zone}2:/home/{new-host-zone}/.{new-host-zone}d

    ...

  {new-host-zone}5:
    image: stridezone:{new-host-zone}
    volumes:
      - ./scripts/state/{new-host-zone}5:/home/{new-host-zone}/.{new-host-zone}
```
* Add the following parameters to `scripts/vars.sh`, where `CHAIN` is the ID of the new host zone. For the relayer, you can use the mnemonic below or create your own. Note: you'll have to add the variables in the right places in `scripts/vars.sh`, as noted below.
```
# Add at the top of vars.sh
{CHAIN}_DENOM="{min_denom}"
ST{CHAIN}_DENOM="st{min_denom}"

# Add below the final host zone
{CHAIN}_CHAIN_ID={NEW-HOST-ZONE}
{CHAIN}_NODE_PREFIX={new-host-zone}
{CHAIN}_NUM_NODES=3
{CHAIN}_CMD="$SCRIPT_DIR/../build/{new-host-zone}d"
{CHAIN}_VAL_PREFIX={n}val
{CHAIN}_ADDRESS_PREFIX=stars
{CHAIN}_REV_ACCT={n}rev1
{CHAIN}_DENOM=${CHAIN}_DENOM
{CHAIN}_RPC_PORT={the one included in the docker-compose above}
{CHAIN}_MAIN_CMD="${CHAIN}_CMD --home $SCRIPT_DIR/state/${${CHAIN}_NODE_PREFIX}1"

# Add *below* the RELAYER section!
RELAYER_{CHAIN}_EXEC="docker-compose run --rm relayer-{new-host-zone}"
RELAYER_{CHAIN}_ACCT=rly{add one since the account from the last host zone}
# NOTE: Update the HOST_RELAYER_ACCTS variable directly!
HOST_RELAYER_ACCTS=(... $RELAYER_{CHAIN}_ACCT)
# stride1muwz5er4wq7svxnh5dgn2tssm92je5dwthxl7q
RELAYER_{CHAIN}_MNEMONIC="science depart where tell bus ski laptop follow child bronze rebel recall brief plug razor ship degree labor human series today embody fury harvest"
# NOTE: Update the RELAYER_MNEMONICS variable directly!
RELAYER_MNEMONICS=(...,"$RELAYER_{CHAIN}_MNEMONIC")

```
* Add the IBC denom's for the host zone across each channel to `vars.sh` (e.g. `IBC_{HOST}_CHANNEL_{N}_DENOM)`). You can generate the variables by uncommenting `x/stakeibc/keeper/get_denom_traces_test.go` and running `make test-unit`. Add the output to `scripts/vars.sh`. Note: You have to run the test using the "run test" button in VSCode, or pass in the `-v` flag and run the tests using `go test -mod=readonly ./x/stakeibc/...`, for the output to show up.
* Add a section to the `scripts/config/relayer_config.yaml`
```
chains:
  ...
  {new-host-zone}:
    type: cosmos
    value:
      key: rly{N}
      chain-id: {CHAIN_ID}
      rpc-addr: http://{node_prefix}1:26657
      account-prefix: {bech32_hrp_account_prefix}
      keyring-backend: test
      gas-adjustment: 1.2
      gas-prices: 0.01{minimal_denom}
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
* To enable the the new host zone, include it in the `HOST_CHAINS` array in `scripts/vars.sh`. **Note: You can only run up to 4 host zones at once. You can just run GAIA and the new host zone, for simplicity (see below).**
```
HOST_CHAINS=(GAIA {NEW-HOST-ZONE})
```
* And that's it! Just start the network as normal, and make sure to rebuild the new host zone when running for the first time. 
```
make build-docker build=n
make start-docker
```
