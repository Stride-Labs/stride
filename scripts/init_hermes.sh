#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# import dependencies
source ${SCRIPT_DIR}/vars.sh

CSLEEP 10
echo "Restoring keys"
docker-compose run --rm hermes hermes -c /tmp/hermes.toml keys restore --mnemonic "$RLY_MNEMONIC_1" $STRIDE_CHAIN
docker-compose run --rm hermes hermes -c /tmp/hermes.toml keys restore --mnemonic "$RLY_MNEMONIC_2" $GAIA_CHAIN

echo "creating hermes identifiers"
docker-compose run --rm hermes hermes -c /tmp/hermes.toml tx raw create-client $STRIDE_CHAIN $GAIA_CHAIN > /dev/null
docker-compose run --rm hermes hermes -c /tmp/hermes.toml tx raw conn-init $STRIDE_CHAIN $GAIA_CHAIN 07-tendermint-0 07-tendermint-0 > /dev/null

echo "Creating connection $STRIDE_CHAIN <> $GAIA_CHAIN"
docker-compose run --rm -T hermes hermes -c /tmp/hermes.toml create connection $STRIDE_CHAIN $GAIA_CHAIN > /dev/null

echo "Creating transfer channel"
docker-compose run --rm -T hermes hermes -c /tmp/hermes.toml create channel --port-a transfer --port-b transfer $GAIA_CHAIN connection-0 > /dev/null
# docker-compose run hermes hermes -c /tmp/hermes.toml tx raw chan-open-init $STRIDE_CHAIN $GAIA_CHAIN connection-0 transfer transfer > /dev/null

echo "Starting hermes relayer"
docker-compose up --force-recreate -d hermes
