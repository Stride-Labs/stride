STR1_EXEC="docker-compose --ansi never exec -T stride1 strided --home /stride/.strided --chain-id STRIDE"
GETBAL() {
    head -n 1 | grep -o -E '[0-9]+'
  }
IBCATOM='ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9'

$STR1_EXEC q stakeibc list-host-zone
# $STR1_EXEC q stakeibc module-address stakeibc
# $STR1_EXEC q bank balances stride1mvdq4nlupl39243qjz7sds5ez3rl9mnx253lza

exit 

docker-compose --ansi never exec -T stride1 strided --home /stride/.strided --chain-id STRIDE ibc-transfer transfer transfer channel-1 cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2 1000ustrd --from val1 --chain-id STRIDE -y --keyring-backend test

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