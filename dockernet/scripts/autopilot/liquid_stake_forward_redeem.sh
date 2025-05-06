#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

# High level test: LS + forward, then autopilot redeem twice
# (Step 1) Liquid stake from GAIA and forward back to original address
#   - verify stATOM balance for GAIA_ADDRESS increased
# (Step 2) Redeem once
#   - verify stATOM balance for GAIA_ADDRESS decreased
#   - verify EUR / HZU / URR were all updated correctly
# (Step 3) Redeem a second time
#   - verify stATOM balance for GAIA_ADDRESS decreased
#   - verify EUR / HZU / URR were all updated correctly

DENOM_STATOM_ON_GAIA="ibc/054A44EC8D9B68B9A6F0D5708375E00A5569A28F21E0064FF12CADC3FEF1D04F"

# (Step 1)
# echo newline

echo "LS + FORWARD (STEP 1)"
# Autopilot liquid stake and forward from GAIA to STRIDE
# store the GAIA_ADDRESS stATOM balance before liquid staking
GAIA_ADDRESS_BALANCE_BEFORE=$(GET_BALANCE GAIA $(GAIA_ADDRESS) $DENOM_STATOM_ON_GAIA)
memo='{ "autopilot": { "receiver": "'"$(STRIDE_ADDRESS)"'", "stakeibc": { "action": "LiquidStake", "ibc_receiver": "'$(GAIA_ADDRESS)'" } } }'
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 "$memo" 5uatom --from ${GAIA_VAL_PREFIX}1 -y
sleep 10
GAIA_ADDRESS_BALANCE_AFTER=$(GET_BALANCE GAIA $(GAIA_ADDRESS) $DENOM_STATOM_ON_GAIA)
# subtract GAIA_ADDRESS_BALANCE_AFTER - GAIA_ADDRESS_BALANCE_BEFORE and verify the result is 5uatom
echo "GAIA_ADDRESS_BALANCE_AFTER - GAIA_ADDRESS_BALANCE_BEFORE == 5uatom: $(($GAIA_ADDRESS_BALANCE_AFTER - $GAIA_ADDRESS_BALANCE_BEFORE == 5))"

# (Step 2)
echo -e "\n\n\n\n\n"
echo "REDEEM ONCE (STEP 2)"
# Autopilot redeem once
GAIA_ADDRESS_BALANCE_BEFORE=$(GET_BALANCE GAIA $(GAIA_ADDRESS) $DENOM_STATOM_ON_GAIA)
memo='{ "autopilot": { "receiver": "'"$(STRIDE_ADDRESS)"'",  "stakeibc": { "action": "RedeemStake", "ibc_receiver": "'$(GAIA_ADDRESS)'" } } }'
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 "$memo" 5$DENOM_STATOM_ON_GAIA --from ${GAIA_VAL_PREFIX}1 -y
sleep 6
GAIA_ADDRESS_BALANCE_AFTER=$(GET_BALANCE GAIA $(GAIA_ADDRESS) $DENOM_STATOM_ON_GAIA)
# subtract GAIA_ADDRESS_BALANCE_AFTER - GAIA_ADDRESS_BALANCE_BEFORE and verify the result is 5uatom
echo "GAIA_ADDRESS_BALANCE_AFTER - GAIA_ADDRESS_BALANCE_BEFORE == -5stuatom: $(($GAIA_ADDRESS_BALANCE_AFTER - $GAIA_ADDRESS_BALANCE_BEFORE == -5))"

# Check the EpochUnbondingRecord / HostZoneUnbonding / UserRedemptionRecord
# get the current epoch
EPOCH=$($STRIDE_MAIN_CMD q epochs current-epoch day | awk -F'"' '{print $2}')
$STRIDE_MAIN_CMD q records show-epoch-unbonding-record $EPOCH
URR_KEY=$GAIA_CHAIN_ID.$EPOCH.$(GAIA_ADDRESS)
$STRIDE_MAIN_CMD q records show-user-redemption-record $URR_KEY

# (Step 3)
echo -e "\n\n\n\n\n"
echo "REDEEM TWICE (STEP 3)"
# Autopilot redeem a second time
GAIA_ADDRESS_BALANCE_BEFORE=$(GET_BALANCE GAIA $(GAIA_ADDRESS) $DENOM_STATOM_ON_GAIA)
memo='{ "autopilot": { "receiver": "'"$(STRIDE_ADDRESS)"'",  "stakeibc": { "action": "RedeemStake", "ibc_receiver": "'$(GAIA_ADDRESS)'" } } }'
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 "$memo" 5$DENOM_STATOM_ON_GAIA --from ${GAIA_VAL_PREFIX}1 -y
sleep 6
GAIA_ADDRESS_BALANCE_AFTER=$(GET_BALANCE GAIA $(GAIA_ADDRESS) $DENOM_STATOM_ON_GAIA)
# subtract GAIA_ADDRESS_BALANCE_AFTER - GAIA_ADDRESS_BALANCE_BEFORE and verify the result is 5uatom
echo "GAIA_ADDRESS_BALANCE_AFTER - GAIA_ADDRESS_BALANCE_BEFORE == -5stuatom: $(($GAIA_ADDRESS_BALANCE_AFTER - $GAIA_ADDRESS_BALANCE_BEFORE == -5))"

# Check the records again
EPOCH=$($STRIDE_MAIN_CMD q epochs current-epoch day | awk -F'"' '{print $2}')
$STRIDE_MAIN_CMD q records show-epoch-unbonding-record $EPOCH
URR_KEY=$GAIA_CHAIN_ID.$EPOCH.$(GAIA_ADDRESS)
$STRIDE_MAIN_CMD q records show-user-redemption-record $URR_KEY
