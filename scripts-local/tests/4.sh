### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source ${SCRIPT_DIR}/../account_vars.sh

# check balances before claiming redeemed stake
$GAIA_CMD q bank balances cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf

#claim stake
$STRIDE_CMD tx stakeibc claim-undelegated-tokens GAIA 3 --from val1 --keyring-backend test --chain-id STRIDE -y

CSLEEP 30
# check balances after claiming redeemed stake
$GAIA_CMD q bank balances cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf
