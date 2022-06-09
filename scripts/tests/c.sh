STR1_EXEC="docker-compose --ansi never exec -T stride1 strided --home /stride/.strided --chain-id STRIDE"
GAIA1_EXEC="docker-compose --ansi never exec -T gaia1 gaiad --home /gaia/.gaiad --chain-id GAIA"
GETBAL() {
    head -n 1 | grep -o -E '[0-9]+'
  }
IBCATOM='ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2'
STRIDE_ADDRESS_1='stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7'
GAIA_ADDRESS_1='cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2'
GAIA_DELEGATE='cosmos19l6d3d7k2pel8epgcpxc9np6fsvjpaaa06nm65vagwxap0e4jezq05mmvu'

# $STR1_EXEC q stakeibc list-host-zone
# MODULE_ADDR=$($STR1_EXEC q stakeibc module-address stakeibc | awk '{print $NF}') 
# echo $MODULE_ADDR
# $STR1_EXEC q bank balances $MODULE_ADDR

# $STR1_EXEC q bank balances $STRIDE_ADDRESS_1
# $GAIA1_EXEC q bank balances $GAIA_DELEGATE

$GAIA1_EXEC q staking delegation cosmos19l6d3d7k2pel8epgcpxc9np6fsvjpaaa06nm65vagwxap0e4jezq05mmvu cosmosvaloper19e7sugzt8zaamk2wyydzgmg9n3ysylg6na6k6e

exit 

docker-compose --ansi never exec -T stride1 strided --home /stride/.strided --chain-id STRIDE ibc-transfer transfer transfer channel-0 cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2 1000ustrd --from val1 --chain-id STRIDE -y --keyring-backend test

# query open channels
docker-compose --ansi never exec -T stride1 strided --home /stride/.strided --chain-id STRIDE q ibc channel channels 

docker-compose --ansi never exec -T stride1 strided --home /stride/.strided --chain-id STRIDE q ibc-transfer denom-trace C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9
docker-compose --ansi never exec -T stride1 strided tx stakeibc register-host-zone C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9 ATOM stATOM --chain-id STRIDE --home /stride/.strided --keyring-backend test --from val1 --gas 500000 -y


ibcaddr=$($STR1_EXEC q stakeibc module-address stakeibc | awk '{print $NF}') 
module_atom=$($STR1_EXEC q bank balances $ibcaddr --denom $IBCATOM | GETBAL)
echo $module_atom
echo $ibcaddr
echo $($STR1_EXEC q bank balances $ibcaddr --denom $IBCATOM)

docker-compose --ansi never exec -T gaia1 gaiad --home /gaia/.gaiad q bank balances cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2



docker-compose --ansi never exec -T gaia1 gaiad --home /gaia/.gaia q ibc channel channels