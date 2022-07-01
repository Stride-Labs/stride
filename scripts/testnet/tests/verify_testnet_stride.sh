
set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source $SCRIPT_DIR/test_vars.sh

# SET GAIA ADDRESS TO THE DESIRED VALIDATOR
GAIA_VAL_ADDR=""
STRIDE_ACCT="val1"
STRIDE_ADDR="stride159atdlc3ksl50g0659w5tq42wwer334ajl7xnq"
IBCATOM="ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"

strided q bank balances $STRIDE_ADDR
sleep 5

strided tx stakeibc register-host-zone connection-0 uatom $IBCATOM channel-0 3 --chain-id STRIDE \
 --keyring-backend test --from val2 --gas 1000000 -y

sleep 5
strided tx stakeibc add-validator GAIA gval1 $GAIA_VAL_ADDR 10 5 --chain-id STRIDE --keyring-backend test --from STRIDE_ACCT -y

sleep 5
strided q stakeibc list-host-zone

#
#    0. Run the above command `q bank balances val1` to check that tokens were IBC'd over
#    1. Replace GAIA_VAL_ADDR with the cosmosvaloper address found in the other file
#    2. Run the above command register-host-zone
#    3. Run the above command add-validator
#    4. Run the `q stakeibc list-host-zone` function to see if the validator was added



