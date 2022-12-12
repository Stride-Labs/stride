### IBC TRANSFER
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../config.sh

## IBC ATOM from GAIA to STRIDE

# $STRIDE_MAIN_CMD q stakeibc show-host-zone evmos_9001-2
# exit
# $STRIDE_MAIN_CMD q stakeibc show-host-zone GAIA
# $EVMOS_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) 99999999999uevmos --from ${EVMOS_VAL_PREFIX}1 -y 
# sleep 10
$STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS)

# $EVMOS_MAIN_CMD q bank balances evmos1e0ah3tdx555h2rx58qx2s7ct337nectaqr6nq8
# $EVMOS_MAIN_CMD q tx C39E390C569DBC9C91B1B02B6B934D098B478D11627DC21FA1D553541527D74F