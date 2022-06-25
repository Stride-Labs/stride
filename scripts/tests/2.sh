### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
GAIA1_EXEC="build/gaiad --home scripts-local/state/gaia"
STR1_EXEC="build/strided --home scripts-local/state/stride"
SCRIPT_DIR="./scripts-local/logs/"
STRIDE_ADDRESS="stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7"

$STR1_EXEC tx stakeibc liquid-stake 1000 uatom --keyring-backend test --from val1 -y --chain-id STRIDE