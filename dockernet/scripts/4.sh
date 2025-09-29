### CLAIM
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/../config.sh

# check balances before claiming redeemed stake
$CELESTIA_MAIN_CMD q bank balances $CELESTIA_RECEIVER_ADDRESS

#claim stake
EPOCH=$($STRIDE_MAIN_CMD q records list-user-redemption-record  | grep -Fiw 'epoch_number' | head -n 1 | grep -o -E '[0-9]+')
SENDER=stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7
$STRIDE_MAIN_CMD tx stakeibc claim-undelegated-tokens CELESTIA $EPOCH $CELESTIA_RECEIVER_ADDRESS --from ${STRIDE_VAL_PREFIX}1 -y 

CSLEEP 30
# check balances after claiming redeemed stake
$CELESTIA_MAIN_CMD q bank balances $CELESTIA_RECEIVER_ADDRESS
