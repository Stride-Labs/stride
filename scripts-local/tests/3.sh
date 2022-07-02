### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source ${SCRIPT_DIR}/../account_vars.sh

# check balances of receiver acct before issuing a redemption
# build/gaiad --home ./scripts-local/state/gaia q bank balances cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf

# issue a redemption
# build/strided --home ./scripts-local/state/stride tx stakeibc redeem-stake 100 stuatom  cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf --from val1 --keyring-backend test --chain-id STRIDE
$STRIDE_CMD tx stakeibc redeem-stake 89 GAIA cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf --from $STRIDE_VAL_ACCT --keyring-backend test --chain-id $STRIDE_CHAIN -y


build/gaiad --home ./scripts-local/state/gaia q bank balances cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf