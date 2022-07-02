### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source ${SCRIPT_DIR}/../account_vars.sh

## GAIA
#  ibc over atoms to stride
STRIDE_ADDRESS=$($STRIDE_CMD keys show $STRIDE_VAL_ACCT --keyring-backend test -a)

$GAIA_CMD tx ibc-transfer transfer transfer channel-0 $STRIDE_ADDRESS 100000uatom --from $GAIA_VAL_ACCT --chain-id $GAIA_CHAIN -y --keyring-backend test
CSLEEP 5
$STRIDE_CMD q bank balances $STRIDE_ADDRESS

# register our validators

sleep 5
$GAIA_CMD tx bank send gval1 $GAIA_VAL_2_ADDR 10000uatom --chain-id $GAIA_CHAIN --keyring-backend test -y
sleep 5
$GAIA_CMD tx bank send gval1 $GAIA_VAL_3_ADDR 10000uatom --chain-id $GAIA_CHAIN --keyring-backend test -y

sleep 5
$STRIDE_CMD tx stakeibc add-validator GAIA gval1 $GAIA_DELEGATE_VAL 10 5 --chain-id $STRIDE_CHAIN --keyring-backend test --from $STRIDE_VAL_ACCT -y
sleep 5
$STRIDE_CMD tx stakeibc add-validator GAIA gval2 $GAIA_DELEGATE_VAL_2 10 10 --chain-id $STRIDE_CHAIN --keyring-backend test --from $STRIDE_VAL_ACCT -y
