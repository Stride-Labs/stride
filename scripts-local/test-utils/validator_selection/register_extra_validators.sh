SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source $SCRIPT_DIR/local_vars.sh

# # now move tokens to validators
$GAIA_CMD tx bank send gval1 $GAIA_VAL_2_ADDR 10000uatom --chain-id $GAIA_CHAIN --keyring-backend test -y
sleep 3
$GAIA_CMD tx bank send gval1 $GAIA_VAL_3_ADDR 10000uatom --chain-id $GAIA_CHAIN --keyring-backend test -y
sleep 3

# lastly register these as validators on Stride
$STRIDE_CMD tx stakeibc add-validator GAIA gval1 $GAIA_DELEGATE_VAL 10 5 --chain-id $STRIDE_CHAIN --keyring-backend test --from $STRIDE_VAL_ACCT -y
sleep 3
$STRIDE_CMD tx stakeibc add-validator GAIA gval2 $GAIA_DELEGATE_VAL_2 10 10 --chain-id $STRIDE_CHAIN --keyring-backend test --from $STRIDE_VAL_ACCT -y
