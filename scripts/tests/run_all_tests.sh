#!/bin/bash
BASE_SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

$BASE_SCRIPT_DIR/bats/bin/bats $BASE_SCRIPT_DIR/basic_tests.bats

# echo $STR1_EXEC
# echo $GAIA_ADDRESS_1
# docker-compose run hermes hermes -c /tmp/hermes.toml tx raw create-client STRIDE GAIA
# docker-compose run -T hermes hermes -c /tmp/hermes.toml create channel --port-a transfer --port-b transfer GAIA connection-0 
# docker-compose run -T hermes hermes -c /tmp/hermes.toml version
# docker-compose run hermes hermes -c /tmp/hermes.toml keys restore --mnemonic "$RLY_MNEMONIC_1" $STRIDE_CHAIN
# docker-compose run hermes hermes -c /tmp/hermes.toml keys restore --mnemonic "$RLY_MNEMONIC_2" $GAIA_CHAIN
# docker-compose --ansi never exec -T gaia1 gaiad version
# docker-compose --ansi never exec -T stride1 strided tx ibc-transfer transfer channel-1 cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2: 1000ustrd --home /stride/.strided --keyring-backend test --from val1 --chain-id STRIDE -y
