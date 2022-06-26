### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/../account_vars.sh

GAIA_DELEGATE="cosmos19l6d3d7k2pel8epgcpxc9np6fsvjpaaa06nm65vagwxap0e4jezq05mmvu"
GAIA_DELEGATE_VAL="cosmosvaloper19e7sugzt8zaamk2wyydzgmg9n3ysylg6na6k6e"
GAIA_WITHDRAWAL="cosmos1lcnmjwjy2lnqged5pnrc0cstz0r88rttunla4zxv84mee30g2q3q48fm53"
GAIA1_EXEC="docker-compose --ansi never exec -T gaia1 gaiad --home /gaia/.gaiad"

$GAIA_CMD q bank balances $GAIA_DELEGATE
echo "-----------------------------"
$GAIA_CMD q distribution rewards $GAIA_DELEGATE
echo "-----------------------------"
$GAIA_CMD q staking delegations $GAIA_DELEGATE 
