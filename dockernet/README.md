`/dockernet` contains infrastructure that is used for testing and development of the Stride protocol. The scripts here support docker-image based testing, some of which are heavily inspired by those used by Osmosis and Quicksilver (although there have been large deviations from the original implementations since). The relevant licenses are included here.

## Dockernet
### Adding a new host zone
* Create a new dockerfile to `dockernet/dockerfiles` (named `Dockerfile.{new-host-zone}`). Use one of the other host zone's dockerfile's as a starting port to provide the certain boilerplate such as the package installs, adding user, exposing ports, etc. You can often find a dockerfile in the github directory of the host zone. In the dockerfile, set `COMMIT_HASH` to the current mainnet commit hash of the chain being tested (or the target commit hash, if we're launching the zone in the future after an upgrade). For newer chains, create a branch and a pull-request, but *do not* merge it (we don't maintain test versions of each chain).
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
* Add the build command for that host zone in `dockernet/build.sh`. For most zones, we use the first letter of the zone, for the new zone, just use `n` (since it won't be merged in, it won't conflict with anything).
```
while getopts sgojhir{n} flag; do
   case "${flag}" in
   ...
   n) build_local_and_docker {new-host-zone} deps/{new-host-zone} ;;  
```
* Add the host zone and relayer to `dockernet/docker-compose.yml`. Add 5 nodes, adding port forwarding to the first node only. Add the relayer. Drop the RPC port number by 100, and the API/gRPC port by 10, relative to the last host zone that was added.
```
  {new-host-zone}1:
    image: stridezone:{new-host-zone}
    volumes:
      - ./dockernet/state/{new-host-zone}1:/home/{new-host-zone}/.{new-host-zone}d
    ports:
      - "{rpc-port}:26657"
      - "{api-port}:1317"
      - "{grpc-port}:9090"

  {new-host-zone}2:
    image: stridezone:{new-host-zone}
    volumes:
      - ./dockernet/state/{new-host-zone}2:/home/{new-host-zone}/.{new-host-zone}d

    ...

  {new-host-zone}5:
    image: stridezone:{new-host-zone}
    volumes:
      - ./dockernet/state/{new-host-zone}5:/home/{new-host-zone}/.{new-host-zone}d
  ...
  relayer-{chain_id}:
    image: stridezone:relayer
    volumes:
      - ./state/relayer-{chain_id}:/home/relayer/.relayer
    restart: always
    command: [ "bash", "start.sh", "stride-{chain_id}" ]
```
* Add the following parameters to `dockernet/config.sh`, where `CHAIN` is the ID of the new host zone. For the relayer, you can use the mnemonic below or create your own. Note: you'll have to add the variables in the right places in `dockernet/config.sh`, as noted below.
```
# add to the top of dockernet/config.sh
{CHAIN}_DENOM="{min_denom}"
ST{CHAIN}_DENOM="st{min_denom}"

# add in the new chain's config section
{CHAIN}_CHAIN_ID={NEW-HOST-ZONE}
{CHAIN}_NODE_PREFIX={new-host-zone}
{CHAIN}_NUM_NODES=3
{CHAIN}_BINARY="$DOCKERNET_HOME/../build/{new-host-zone}d"
{CHAIN}_VAL_PREFIX={n}val
{CHAIN}_ADDRESS_PREFIX=stars
{CHAIN}_REV_ACCT={n}rev1
{CHAIN}_DENOM=${CHAIN}_DENOM
{CHAIN}_RPC_PORT={the one included in the docker-compose above}
{CHAIN}_MAIN_CMD="${CHAIN}_CMD --home $DOCKERNET_HOME/state/${${CHAIN}_NODE_PREFIX}1"
{CHAIN}_RECEIVER_ADDRESS={any random address on the chain}

# Optionally, if the chain has a micro-denom granularity beyond 6 digits, 
# specify the number of 0's in the following:
# e.g. evmos uses 18 digits, so 18 zero's should be included in the variable
# If this variable is excluded, it will default to 6 digits
{CHAIN}_MICRO_DENOM_UNITS=000000000000000000

# Add *below* the RELAYER section!
RELAYER_{CHAIN}_EXEC="docker-compose run --rm relayer-{new-host-zone}"
RELAYER_{CHAIN}_ACCT=rly{add one since the account from the last host zone}

# NOTE: Update the RELAYER_ACCTS variable directly!
RELAYER_ACCTS=(... $RELAYER_{CHAIN}_ACCT)

# stride1muwz5er4wq7svxnh5dgn2tssm92je5dwthxl7q
RELAYER_{CHAIN}_MNEMONIC="science depart where tell bus ski laptop follow child bronze rebel recall brief plug razor ship degree labor human series today embody fury harvest"
# NOTE: Update the RELAYER_MNEMONICS variable directly!
RELAYER_MNEMONICS=(
    ...
    "$RELAYER_{CHAIN}_MNEMONIC"
)

# Add the {CHAIN_ID}_ADDRESS function
${CHAIN_ID}_ADDRESS() { 
  $${CHAIN_ID}_MAIN_CMD keys show ${${CHAIN_ID}_VAL_PREFIX}1 --keyring-backend test -a 
}

```
* Add the IBC denom's for the host zone across each channel to `dockernet/config.sh` (e.g. `IBC_{HOST}_CHANNEL_{N}_DENOM)`). You can generate the variables by uncommenting `x/stakeibc/keeper/get_denom_traces_test.go`, specifying the ChainID and denom, and running `make test-unit`. Add the output to `dockernet/config.sh`. Note: You have to run the test using the "run test" button in VSCode, or pass in the `-v` flag and run the tests using `go test -mod=readonly ./x/stakeibc/...`, for the output to show up.
* Add a section to the `dockernet/config/relayer_config.yaml`. Most chains will use either the cosmos coin type (118) or eth coin type (60). If a new coin type is used, add it to the top of `config.sh` for future reference.
```
chains:
  ...
  {new-host-zone}:
    type: cosmos
    value:
      key: rly{N}
      chain-id: {CHAIN_ID}
      rpc-addr: http://{new-host-zone}1:26657
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
* To enable the the new host zone, include it in the `HOST_CHAINS` array in `dockernet/config.sh`. **Note: You can only run up to 4 host zones at once. Since this wont be merged, for simplicity, you can just run GAIA and the new host zone in the default case (see below).**
```bash
HOST_CHAINS=()  

if [[ "${ALL_HOST_CHAINS:-false}" == "true" ]]; then 
  HOST_CHAINS=(GAIA JUNO OSMO {NEW-HOST-ZONE})      # add here (this controls the hosts in `make start-docker-all`)
elif [[ "${#HOST_CHAINS[@]}" == "0" ]]; then 
  HOST_CHAINS=(GAIA {NEW-HOST-ZONE})                # add here (this controls the hosts in `make start-docker`)
fi
```
* Add the new host to the integration tests in `dockernet/tests/run_all_tests.sh`. When debugging, it's easiest to first test only the new host zone. You can comment out the existing chains and add the new host at the end. **Note: The transfer channel number will be 1 since it's the second host added (the first host is 0).** It should look something like:
``` bash
# CHAIN_NAME=GAIA TRANSFER_CHANNEL_NUMBER=0 $BATS $INTEGRATION_TEST_FILE
# CHAIN_NAME=JUNO TRANSFER_CHANNEL_NUMBER=1 $BATS $INTEGRATION_TEST_FILE
# CHAIN_NAME=OSMO TRANSFER_CHANNEL_NUMBER=2 $BATS $INTEGRATION_TEST_FILE
CHAIN_NAME={NEW-HOST-ZONE} TRANSFER_CHANNEL_NUMBER=1 $BATS $INTEGRATION_TEST_FILE
```
* Start the network as normal. Make sure to rebuild the new host zone when running for the first time. You can view the logs in `dockernet/logs/{new-host-zone}.log` to ensure the network started successfully.
```
make build-docker build=n
make start-docker
```
* After the chain is running, run the integration tests to confirm the new host zone is compatible with Stride
```
make test-integration-docker
```
* After the tests succeed, you can add back in the other hosts to the integration tests. **Note: The transfer channel for the new host will need to be updated from 1 to 3, since it is now the 4th host zone.**
```
CHAIN_NAME=GAIA TRANSFER_CHANNEL_NUMBER=0 $BATS $INTEGRATION_TEST_FILE
CHAIN_NAME=JUNO TRANSFER_CHANNEL_NUMBER=1 $BATS $INTEGRATION_TEST_FILE
CHAIN_NAME=OSMO TRANSFER_CHANNEL_NUMBER=2 $BATS $INTEGRATION_TEST_FILE
CHAIN_NAME={NEW-HOST-ZONE} TRANSFER_CHANNEL_NUMBER=3 $BATS $INTEGRATION_TEST_FILE
```
* Finally, restart dockernet with all hosts, and confirm all integration tests pass 
```
make start-docker-all 
make test-integration-docker
```
* If all tests pass, the host zone is good to go!

