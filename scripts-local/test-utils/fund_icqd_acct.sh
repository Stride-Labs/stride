### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source ${SCRIPT_DIR}/../account_vars.sh

echo "Funding the account we query using ICQ...\n\n"
$GAIA_CMD tx bank send $GAIA_VAL_ADDR $WITHDRAWAL_ICA_ADDR 1234uatom --from $GAIA_VAL_ACCT -y
sleep 3
echo "Funded the account we query using ICQ, querying to verify...\n\n"
$GAIA_CMD q bank balances $WITHDRAWAL_ICA_ADDR 

