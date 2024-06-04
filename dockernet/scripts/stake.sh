SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../config.sh

echo ">>> gaiad tx staking delegate"
$GAIA_MAIN_CMD tx staking delegate $(GET_VAL_ADDR GAIA 1) 1000000uatom --from user -y | TRIM_TX
sleep 5

echo -e "\nDelegations:"
$GAIA_MAIN_CMD q staking delegations $($GAIA_MAIN_CMD keys show -a user)