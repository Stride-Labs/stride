### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source ${SCRIPT_DIR}/../account_vars.sh

$GAIA_CMD q staking delegation cosmos19l6d3d7k2pel8epgcpxc9np6fsvjpaaa06nm65vagwxap0e4jezq05mmvu cosmosvaloper1pcag0cj4ttxg8l7pcg0q4ksuglswuuedadj7ne

# delegation $DELEGATION_ICA_ADDR cosmosvaloper1pcag0cj4ttxg8l7pcg0q4ksuglswuuedadj7ne


# $GAIA_CMD q staking delegation $DELEGATION_ICA_ADDR cosmosvaloper1pcag0cj4ttxg8l7pcg0q4ksuglswuuedadj7ne
