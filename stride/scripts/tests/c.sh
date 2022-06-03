STR1_EXEC="docker-compose --ansi never exec -T stride1 strided --home /stride/.strided --chain-id STRIDE"
GETBAL() {
    head -n 1 | grep -o -E '[0-9]+'
  }
IBCATOM='ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E'

ibcaddr=$($STR1_EXEC q stakeibc module-address stakeibc | awk '{print $NF}') 
module_atom=$($STR1_EXEC q bank balances $ibcaddr --denom $IBCATOM | GETBAL)
echo $module_atom
echo $ibcaddr

docker-compose --ansi never exec -T gaia1 gaiad --home /gaia/.gaiad q bank balances cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2