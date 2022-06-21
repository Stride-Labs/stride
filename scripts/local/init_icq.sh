# this file should be called from the `stride` folder
# e.g. `sh ./scripts/init.sh`
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# import dependencies
source ${SCRIPT_DIR}/vars.sh

echo "\nAdd ICQ relayer addresses for Stride and Gaia:"
# TODO(TEST-82) redefine stride-testnet in lens' config to $STRIDE_CHAIN and gaia-testnet to $main-gaia-chain, then replace those below with $STRIDE_CHAIN and $GAIA_CHAIN
$ICQ_RUN keys restore test "$ICQ_STRIDE_KEY" --chain stride-testnet
$ICQ_RUN keys restore test "$ICQ_GAIA_KEY" --chain gaia-testnet

echo "\nICQ addresses for Stride and Gaia:"
# TODO(TEST-83) pull these addresses dynamically using jq
ICQ_ADDRESS_STRIDE="stride12vfkpj7lpqg0n4j68rr5kyffc6wu55dzqewda4"
ICQ_ADDRESS_GAIA="cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf"

$STRIDE1_EXEC tx bank send val1 $ICQ_ADDRESS_STRIDE 5000000ustrd --chain-id $STRIDE_CHAIN -y --keyring-backend test --home /stride/.strided
$GAIA1_EXEC tx bank send gval1 $ICQ_ADDRESS_GAIA 5000000uatom --chain-id $GAIA_CHAIN -y --keyring-backend test --home /gaia/.gaiad

echo "\nStarting interchainquery relayer"
docker-compose up --force-recreate -d icq
