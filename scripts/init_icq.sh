# this file should be called from the `stride` folder
# e.g. `sh ./scripts/init.sh`
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# import dependencies
source ${SCRIPT_DIR}/vars.sh
source ${SCRIPT_DIR}/account_vars.sh

echo "\nAdd ICQ relayer addresses for Stride and Gaia:"
# TODO(TEST-82) redefine stride-testnet in lens' config to $main_chain and gaia-testnet to $main-gaia-chain, then replace those below with $main_chain and $main_gaia_chain
$ICQ_RUN keys restore test "$ICQ_STRIDE_KEY" --chain stride-testnet
$ICQ_RUN keys restore test "$ICQ_GAIA_KEY" --chain gaia-testnet

echo "\nICQ addresses for Stride and Gaia:"
# TODO(TEST-83) pull these addresses dynamically using jq
ICQ_ADDRESS_STRIDE="stride12vfkpj7lpqg0n4j68rr5kyffc6wu55dzqewda4"
# echo $ICQ_ADDRESS_STRIDE
ICQ_ADDRESS_GAIA="cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf"
# echo $ICQ_ADDRESS_GAIA

STR1_EXEC="docker-compose --ansi never exec -T stride1 strided --home /stride/.strided --chain-id STRIDE"
$STR1_EXEC tx bank send val1 $ICQ_ADDRESS_STRIDE 5000000ustrd --chain-id $main_chain -y --keyring-backend test --home /stride/.strided
GAIA1_EXEC="docker-compose --ansi never exec -T gaia1 gaiad --home /gaia/.gaiad"
$GAIA1_EXEC tx bank send gval1 $ICQ_ADDRESS_GAIA 5000000uatom --chain-id $main_gaia_chain -y --keyring-backend test --home /gaia/.gaiad

echo "\nLaunch interchainquery relayer"
docker-compose up --force-recreate -d icq