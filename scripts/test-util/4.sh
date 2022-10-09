### LIQ STAKE 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/../vars.sh

# check balances before claiming redeemed stake
$GAIA_MAIN_CMD q bank balances $GAIA_RECEIVER_ACCT

#claim stake
EPOCH=5
SENDER=stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7
$STRIDE_MAIN_CMD tx stakeibc claim-undelegated-tokens GAIA $EPOCH $(STRIDE_ADDRESS) --from ${STRIDE_VAL_PREFIX}1 -y

CSLEEP 30
# check balances after claiming redeemed stake
$GAIA_MAIN_CMD q bank balances $GAIA_RECEIVER_ACCT
