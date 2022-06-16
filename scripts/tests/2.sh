### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source ${SCRIPT_DIR}/../account_vars.sh

# $STR1_EXEC tx stakeibc liquid-stake 500000000000 uatom --keyring-backend test --from val1 -y
$STR1_EXEC tx stakeibc liquid-stake 1000 uatom --keyring-backend test --from val1 -y

# docker-compose --ansi never exec -T stride1 strided --home /stride/.strided --chain-id STRIDE tx stakeibc liquid-stake 1000 uatom --keyring-backend test --from val1 -y

# 5000000000000 <- IBC'd over
# 4998216206336 <- How much is there now 
# 1783793664 <- how much stATOM got minted
# 500000000000 <- how much stATOM we wanted to mint