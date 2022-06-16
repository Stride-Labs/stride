#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# import dependencies
source ${SCRIPT_DIR}/vars.sh

docker-compose up -d stride1 stride2 stride3 gaia1 gaia2 gaia3
echo "Chains creating..."
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
echo "Waiting for hermes to be ready..."

echo "\nBuild interchainquery relayer service (this takes ~120s...)"
rm -rf ./icq/keys
docker-compose build icq --no-cache
ICQ_RUN="docker-compose --ansi never run --rm -T icq interchain-queries"

echo "\nAdd ICQ relayer addresses for Stride and Gaia:"
# TODO(TEST-82) redefine stride-testnet in lens' config to $STRIDE_CHAIN and gaia-testnet to $main-gaia-chain, then replace those below with $STRIDE_CHAIN and $GAIA_CHAIN
$ICQ_RUN keys restore test "$ICQ_STRIDE_KEY" --chain stride-testnet
$ICQ_RUN keys restore test "$ICQ_GAIA_KEY" --chain gaia-testnet

echo "\nICQ addresses for Stride and Gaia:"
# TODO(TEST-83) pull these addresses dynamically using jq
ICQ_ADDRESS_STRIDE="stride12vfkpj7lpqg0n4j68rr5kyffc6wu55dzqewda4"
# echo $ICQ_ADDRESS_STRIDE
ICQ_ADDRESS_GAIA="cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf"
# echo $ICQ_ADDRESS_GAIA

STR1_EXEC="docker-compose --ansi never exec -T stride1 strided --home /stride/.strided --chain-id STRIDE"
$STR1_EXEC tx bank send val1 $ICQ_ADDRESS_STRIDE 5000000ustrd --chain-id $STRIDE_CHAIN -y --keyring-backend test --home /stride/.strided
GAIA1_EXEC="docker-compose --ansi never exec -T gaia1 gaiad --home /gaia/.gaiad"
$GAIA1_EXEC tx bank send gval1 $ICQ_ADDRESS_GAIA 5000000uatom --chain-id $GAIA_CHAIN -y --keyring-backend test --home /gaia/.gaiad

echo "\nLaunch interchainquery relayer"
docker-compose up --force-recreate -d icq
