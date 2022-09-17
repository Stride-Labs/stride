`/scripts` contains (unmaintained) infrastructure that was used for early testing and development of the Stride protocol. The scripts here support docker-image based testing, some of which are heavily inspired by those used by Osmosis and Quicksilver (although there have been deviations from the original implementations since). The relevant licenses are included here.

## Dockernet
### Adding a new host zone
* Create a new dockerfile at the root level (`Dockerfile.{new-host-zone})
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
* Add the host zone as a submodule in `deps`
* Add the build command for that host zone in `scripts/build.sh`
```
while getopts sgojhir{n} flag; do
   case "${flag}" in
   ...
   n) build_local_and_docker {new-host-zone} deps/{new-host-zone} ;;  
```
* Add the following parameters to `scripts/vars.sh`, where `CHAIN_ID` is the ID of the new host zone
```
{CHAIN_ID}_CHAIN_ID={NEW-HOST-ZONE}
{CHAIN_ID}_NODE_PREFIX={new-host-zone}
{CHAIN_ID}_NUM_NODES=3
{CHAIN_ID}_CMD="$SCRIPT_DIR/../build/{new-host-zone}d"
{CHAIN_ID}_VAL_PREFIX={n}val
{CHAIN_ID}_REV_ACCT={n}rev1
{CHAIN_ID}_DENOM=
{CHAIN_ID}_IBC_DENOM=
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

ICQ_${CHAIN_ID}_ACCT=rly{add one since the account from the last host zone}
ICQ_${CHAIN_ID}_MNEMONIC=""
```
* Finally add the execution of the `init_chain` script for this host zone in `scripts/start_network.sh`, and add it to the array of `HOST_CHAINS`
```
sh ${SCRIPT_DIR}/init_chain.sh {NEW-HOST-ZONE}
HOST_CHAINS=(GAIA JUNO OSMO ... {NEW-HOST-ZONE})
```
* And that's it! Just start the network as normal, and make sure to rebuild the new host zone when running for the first time.  