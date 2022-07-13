SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source $SCRIPT_DIR/local_vars.sh

# $GAIA_CMD q staking validators
# $GAIA_CMD keys list
# $GAIA_CMD tx staking create-validator 
# $GAIA_CMD keys add --algo ed25519 --backend test
# --amount 10000uatom --from $GAIA_VAL_ACCT_2 --pubkey $GAIA_VAL_2_PUBKEY --chain-id $GAIA_CHAIN --keyring-backend test -y --commission-max-change-rate 0.1 --commission-max-rate 0.1 --commission-rate 0.1 --min-self-delegation 1

$STRIDE_CMD keys add gval4 --keyring-backend test --algo ed25519
# Create a gentx.
# $GAIA_CMD gentx gval2 100000000uatom --chain-id $GAIA_CHAIN --keyring-backend test --pubkey $GAIA_VAL_2_PUBKEY --output-document=$STATE/gaia/config/gentx/gval2.json

# Add the gentx to the genesis file.
# simd collect-gentxs

# lastly add these as validators
# $STRIDE_CMD tx stakeibc add-validator GAIA gval1 $GAIA_VAL_ADDR 10 5 --chain-id $STRIDE_CHAIN --keyring-backend test --from $STRIDE_VAL_ACCT -y
# sleep 1
# $STRIDE_CMD tx stakeibc add-validator GAIA gval2 $GAIA_VAL_2_ADDR 10 10 --chain-id $STRIDE_CHAIN --keyring-backend test --from $STRIDE_VAL_ACCT -y
# sleep 1
# $STRIDE_CMD tx stakeibc add-validator GAIA gval3 $GAIA_VAL_3_ADDR 10 10 --chain-id $STRIDE_CHAIN --keyring-backend test --from $STRIDE_VAL_ACCT -y
