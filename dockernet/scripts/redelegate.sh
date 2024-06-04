SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../config.sh

echo ">>> gaiad tx staking redelegate"
$GAIA_MAIN_CMD tx staking redelegate $(GET_VAL_ADDR GAIA 1) $(GET_VAL_ADDR GAIA 2) 100000uatom --from user -y --gas 500000 | TRIM_TX
sleep 5

echo -e "\nDelegations"
$GAIA_MAIN_CMD q staking delegations $($GAIA_MAIN_CMD keys show -a user)
sleep 1

echo -e "\nRedelegations"
$GAIA_MAIN_CMD q staking redelegations $($GAIA_MAIN_CMD keys show -a user)