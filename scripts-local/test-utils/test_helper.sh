### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source ${SCRIPT_DIR}/../account_vars.sh


# $STRIDE_CMD q tx 85A47929321B6909F76F3B7037F3557C3F0AA0E103E27DDA27013859CDB51605
# $GAIA_CMD q bank balances $GAIA_ADDRESS
# $GAIA_CMD q ibc channel channels

# $GAIA_CMD q tx 101026FF938277A12B148E62CF8A11BE99780618C75E6D1F83469EE4FE6C390E

$STRIDE_CMD q bank balances $STRIDE_ADDRESS
# $GAIA_CMD q bank balances cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2
exit



$GAIA_CMD tx ibc-transfer transfer transfer channel-0 $STRIDE_ADDRESS 10000uatom --from gval1 --chain-id GAIA -y --keyring-backend test
$STRIDE_CMD q bank balances $STRIDE_ADDRESS
