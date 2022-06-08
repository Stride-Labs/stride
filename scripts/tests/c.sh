STR1_EXEC="docker-compose --ansi never exec -T stride1 strided --home /stride/.strided --chain-id STRIDE"
GETBAL() {
    head -n 1 | grep -o -E '[0-9]+'
  }
IBCATOM='ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E'

ibcaddr=$($STR1_EXEC q stakeibc module-address stakeibc | awk '{print $NF}') 
module_atom=$($STR1_EXEC q bank balances $ibcaddr --denom $IBCATOM | GETBAL)
echo $module_atom
echo $ibcaddr
echo $($STR1_EXEC q bank balances $ibcaddr --denom $IBCATOM)

docker-compose --ansi never exec -T gaia1 gaiad --home /gaia/.gaiad q bank balances cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2



@test "ibc transfer updates all balances" {
  # get initial balances
  str1_balance=$($STR1_EXEC q bank balances $STRIDE_ADDRESS_1 --denom ustrd | GETBAL)
  gaia1_balance=$($GAIA1_EXEC q bank balances $GAIA_ADDRESS_1 --denom $IBCSTRD | GETBAL)
  str1_balance_atom=$($STR1_EXEC q bank balances $STRIDE_ADDRESS_1 --denom $IBCATOM | GETBAL)
  gaia1_balance_atom=$($GAIA1_EXEC q bank balances $GAIA_ADDRESS_1 --denom uatom | GETBAL)
  # do IBC transfer
  $STR1_EXEC tx ibc-transfer transfer transfer channel-1 $GAIA_ADDRESS_1 1000ustrd --from val1 --chain-id STRIDE -y --keyring-backend test
  $GAIA1_EXEC tx ibc-transfer transfer transfer channel-0 $STRIDE_ADDRESS_1 1000uatom --from gval1 --chain-id GAIA -y --keyring-backend test
  sleep 20
  # get new balances
  str1_balance_new=$($STR1_EXEC q bank balances $STRIDE_ADDRESS_1 --denom ustrd | GETBAL)
  gaia1_balance_new=$($GAIA1_EXEC q bank balances $GAIA_ADDRESS_1 --denom $IBCSTRD | GETBAL)
  str1_balance_atom_new=$($STR1_EXEC q bank balances $STRIDE_ADDRESS_1 --denom $IBCATOM | GETBAL)
  gaia1_balance_atom_new=$($GAIA1_EXEC q bank balances $GAIA_ADDRESS_1 --denom uatom | GETBAL)
  # get all STRD balance diffs
  str1_diff=$(($str1_balance - $str1_balance_new))
  gaia1_diff=$(($gaia1_balance - $gaia1_balance_new))
  assert_equal "$str1_diff" '1000'
  assert_equal "$gaia1_diff" '-1000'
  # get all ATOM balance diffs
  str1_diff=$(($str1_balance_atom - $str1_balance_atom_new))
  gaia1_diff=$(($gaia1_balance_atom - $gaia1_balance_atom_new))
  assert_equal "$str1_diff" '-1000'
  assert_equal "$gaia1_diff" '1000'
}

@test "liquid stake mints stATOM" {
  str1_balance_atom=$($STR1_EXEC q bank balances $STRIDE_ADDRESS_1 --denom $IBCATOM | GETBAL)
  str1_balance_statom=$($STR1_EXEC q bank balances $STRIDE_ADDRESS_1 --denom $STATOM | GETBAL)
  # liquid stake
  $STR1_EXEC tx stakeibc liquid-stake 1000 $IBCATOM --keyring-backend test --from val1 -y
  sleep 5
  # make sure IBCATOM went down 
  str1_balance_atom_new=$($STR1_EXEC q bank balances $STRIDE_ADDRESS_1 --denom $IBCATOM | GETBAL)
  str1_atom_diff=$(($str1_balance_atom - $str1_balance_atom_new))
  assert_equal "$str1_atom_diff" '1000'
  # make sure STATOM went up
  str1_balance_statom_new=$($STR1_EXEC q bank balances $STRIDE_ADDRESS_1 --denom $STATOM | GETBAL)
  str1_statom_diff=$(($str1_balance_statom_new-$str1_balance_statom))
  assert_equal "$str1_statom_diff" '1000'
}

@test "liquid stake IBCs automatically" {
  # get module address 
  ibcaddr=$($STR1_EXEC q stakeibc module-address stakeibc | awk '{print $NF}') 
  module_atom=$($STR1_EXEC q bank balances $ibcaddr --denom $IBCATOM | GETBAL)
  assert_equal "$module_atom" '1000'
}

# add test to register host zone 
@test "host zone successfully registered" {
  run $STR1_EXEC q stakeibc show-host-zone GAIA
  host_zone_info=$($STR1_EXEC q stakeibc get-host-zone  | awk '{print $NF}')
  assert_line '  BaseDenom: ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9'
}

# add test to see if assets are properly being staked on host zone