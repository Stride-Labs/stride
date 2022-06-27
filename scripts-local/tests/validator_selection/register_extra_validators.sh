SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source $SCRIPT_DIR/local_vars.sh

# first register two validators
echo $GAIA_VAL_MNEMONIC_2 | $GAIA_CMD keys add $GAIA_VAL_ACCT_2 --recover --keyring-backend=test
sleep 1
echo $GAIA_VAL_MNEMONIC_3 | $GAIA_CMD keys add $GAIA_VAL_ACCT_3 --recover --keyring-backend=test

# # now move tokens to them
$GAIA_CMD tx bank send gval1 $GAIA_VAL_2_ADDR 10000uatom --chain-id $GAIA_CHAIN --keyring-backend test -y
sleep 1
$GAIA_CMD tx bank send gval1 $GAIA_VAL_3_ADDR 10000uatom --chain-id $GAIA_CHAIN --keyring-backend test -y
sleep 1

# lastly add these validators
$STRIDE_CMD tx stakeibc add-validator GAIA gval1 $GAIA_VAL_ADDR 10 5 --chain-id $STRIDE_CHAIN --keyring-backend test --from $STRIDE_VAL_ACCT -y
sleep 1
$STRIDE_CMD tx stakeibc add-validator GAIA gval2 $GAIA_VAL_2_ADDR 10 10 --chain-id $STRIDE_CHAIN --keyring-backend test --from $STRIDE_VAL_ACCT -y
sleep 1
$STRIDE_CMD tx stakeibc add-validator GAIA gval3 $GAIA_VAL_3_ADDR 10 10 --chain-id $STRIDE_CHAIN --keyring-backend test --from $STRIDE_VAL_ACCT -y
