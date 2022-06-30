
set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source $SCRIPT_DIR/test_vars.sh

# SET STRIDE ADDRESS TO THE DESIRED STRIDE USER

gaiad tx ibc-transfer transfer transfer channel-0 $STRIDE_ADDRESS 100000uatom --from gval1 --chain-id GAIA -y --keyring-backend test
sleep 3

gaiad q staking validators

#
#    1. Get stride address, replace $STRIDE_ADDRESS above with that
#    2. Run the above command `ibc-transfer`
#    3. Run the `q staking validators` command and grab the GAIA validator address
#         this should start with "cosmosvaloper"
#    4. Move to verify_testnet_stride.sh
#

