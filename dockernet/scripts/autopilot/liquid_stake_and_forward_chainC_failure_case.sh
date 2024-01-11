#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

OSMO_ADDRESS="osmo1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrqyeqwq"
DENOM_ATOM_ON_STRIDE="ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
DENOM_ATOM_ON_OSMOSIS_THRU_STRIDE="ibc/6CDD4663F2F09CD62285E2D45891FC149A3568E316CE3EBBE201A71A78A69388"

# VERIFY AUTOPILOT LIQUID STAKING ATOM FROM GAIA WORKS
memo='{ "autopilot": { "receiver": "'"$(STRIDE_ADDRESS)"'", "stakeibc": { "action": "LiquidStake", "ibc_receiver": "'$(GAIA_ADDRESS)'" } } }'
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 "$memo" 1uatom --from ${GAIA_VAL_PREFIX}1 -y

sleep 10

# NOW TRY AUTOPILOT LIQUID STAKING ATOM FROM OSMOSIS (AFTER THAT ATOM WAS IBC'D THROUGH STRIDE).
# THIS SHOULD FAIL
# You should see this error in dockernet's `stride.log`
    # dockernet-stride1-1  | 1:25AM ERR Error liquid staking packet from autopilot for $OSMO_ADDRESS: the native token is not supported for liquid staking module=x/autopilot

## IBC ATOM from GAIA to STRIDE
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) 1000000uatom --from ${GAIA_VAL_PREFIX}1 -y 
sleep 10
$STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS)
## IBC ATOM from STRIDE to OSMOSIS
$STRIDE_MAIN_CMD tx ibc-transfer transfer transfer channel-1 $OSMO_ADDRESS "100$ATOM_ON_STRIDE_DENOM" --from ${STRIDE_VAL_PREFIX}1 -y 
sleep 10
$OSMO_MAIN_CMD q bank balances $OSMO_ADDRESS

memo='{ "autopilot": { "receiver": "'"$(STRIDE_ADDRESS)"'", "stakeibc": { "action": "LiquidStake", "ibc_receiver": "'$(GAIA_ADDRESS)'" } } }'
$OSMO_MAIN_CMD tx ibc-transfer transfer transfer channel-0 "$memo" "1$DENOM_ATOM_ON_OSMOSIS_THRU_STRIDE" --from ${OSMO_VAL_PREFIX}1 -y --gas 900000 --gas-prices 0.0025uosmo

