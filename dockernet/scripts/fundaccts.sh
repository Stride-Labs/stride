### IBC TRANSFER
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../config.sh

# fund the test accounts

COSMOS_WALLET="cosmos1x92tnm6pfkl3gsfy0rfaez5myq5zh99aek2jmd"
STRIDE_WALLET="stride1x92tnm6pfkl3gsfy0rfaez5myq5zh99a6a2w0p"

$GAIA_MAIN_CMD q bank balances $COSMOS_WALLET
sleep 10
$GAIA_MAIN_CMD tx bank send cosmos1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrgl2scj $COSMOS_WALLET 1000000uatom --from ${GAIA_VAL_PREFIX}1 -y 
sleep 10
$GAIA_MAIN_CMD q bank balances $COSMOS_WALLET


$STRIDE_MAIN_CMD q bank balances $STRIDE_WALLET
sleep 10
$STRIDE_MAIN_CMD tx bank send stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7 $STRIDE_WALLET 1000000ustrd --from ${STRIDE_VAL_PREFIX}1 -y 
sleep 10
$STRIDE_MAIN_CMD q bank balances $STRIDE_WALLET