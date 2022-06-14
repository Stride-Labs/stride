### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
# source ${SCRIPT_DIR}/../account_vars.sh

GAIA_DELEGATE="cosmos19l6d3d7k2pel8epgcpxc9np6fsvjpaaa06nm65vagwxap0e4jezq05mmvu"
GAIA_DELEGATE_VAL="cosmosvaloper19e7sugzt8zaamk2wyydzgmg9n3ysylg6na6k6e"
GAIA_WITHDRAWAL="cosmos1lcnmjwjy2lnqged5pnrc0cstz0r88rttunla4zxv84mee30g2q3q48fm53"
GAIA1_EXEC="docker-compose --ansi never exec -T gaia1 gaiad --home /gaia/.gaiad"

GETSTAKE() {
    tail -n 2 | head -n 1 | grep -o -E '[0-9]+' | head -n 1
  }

# $STR1_EXEC q bank balances $STRIDE_ADDRESS_1

# $GAIA1_EXEC q bank balances $GAIA_WITHDRAWAL
# $GAIA1_EXEC q staking delegations $GAIA_WITHDRAWAL
# $GAIA1_EXEC q distribution rewards $GAIA_WITHDRAWAL

# echo "-----------------------------"
$STR1_EXEC strided q stakeibc list-host-zone
# $GAIA1_EXEC q bank balances $GAIA_DELEGATE
# echo "-----------------------------"
# # $GAIA1_EXEC q staking delegation $GAIA_DELEGATE $GAIA_DELEGATE_VAL | GETSTAKE
# $GAIA1_EXEC q distribution rewards $GAIA_DELEGATE
# echo "-----------------------------"
# $GAIA1_EXEC q staking delegations $GAIA_DELEGATE # $GAIA_DELEGATE_VAL #| GETSTAKE
# $GAIA1_EXEC q staking delegation $GAIA_DELEGATE $GAIA_DELEGATE_VAL | GETSTAKE

# $GAIA1_EXEC q bank balances $GAIA_DELEGATE
# $GAIA1_EXEC q staking delegations $GAIA_DELEGATE
 # cosmos19l6d3d7k2pel8epgcpxc9np6fsvjpaaa06nm65vagwxap0e4jezq05mmvu cosmosvaloper19e7sugzt8zaamk2wyydzgmg9n3ysylg6na6k6e
