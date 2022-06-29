SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source $SCRIPT_DIR/local_vars.sh

# send some ATOM over
$GAIA_CMD tx ibc-transfer transfer transfer channel-0 $STRIDE_ADDRESS 100000uatom --from $GAIA_VAL_ACCT --chain-id GAIA -y --keyring-backend test
CSLEEP 5
# query that token balance is there 
$STRIDE_CMD q bank balances $STRIDE_ADDRESS
CSLEEP 3
# check that the right validators are added, with the correct weight

$STRIDE_CMD tx stakeibc liquid-stake 1000 uatom --chain-id $STRIDE_CHAIN --keyring-backend test --from $STRIDE_VAL_ACCT -y
CSLEEP 3

$GAIA_CMD q bank balances cosmos19l6d3d7k2pel8epgcpxc9np6fsvjpaaa06nm65vagwxap0e4jezq05mmvu
