### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source ${SCRIPT_DIR}/../account_vars.sh

# balances
# $STRIDE_CMD q bank balances stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7
# exit

$STRIDE_CMD q staking validators
exit

# Test 0
# delegate from whitelisted address to a validator
$STRIDE_CMD tx staking delegate 
# wait two blocks
sleep 30
# query rewards



# Test 1

# Test 2

# Test 3

# val1 
# val2
# whitelisted address stride1xv9jfnsn3cdktwwcy6rx49el6748hnwh3ezvdk
# blacklisted address stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7

# stake from blacklisted to val2
# stake from whitelisted to val1

# query outstanding rewards for whitelisted address
# query outstanding rewards for blacklisted address