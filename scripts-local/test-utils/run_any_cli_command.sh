### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source ${SCRIPT_DIR}/../account_vars.sh


# transfer tokens to stride
# $GAIA_CMD tx ibc-transfer transfer transfer channel-0 stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7 1000000uatom --from gval1 --chain-id GAIA -y --keyring-backend test
# exit

# $STRIDE_CMD q stakeibc show-host-zone OSMO
# $OSMO_CMD q bank balances osmo1cx04p5974f8hzh2lqev48kjrjugdxsxy7mzrd0eyweycpr90vk8q8d6f3h
# exit

# $STRIDE_CMD q ibc channel channels
# exit

# $STRIDE_CMD q stakeibc list-host-zone
# exit
# $STRIDE_CMD tx stakeibc liquid-stake 10 uatom --keyring-backend test --from val1 -y --chain-id $STRIDE_CHAIN
# exit
$STRIDE_CMD q bank balances stride1755g4dkhpw73gz9h9nwhlcefc6sdf8kcmvcwrk4rxfrz8xpxxjms7savm8
exit


# liquid stake
# $STRIDE_CMD tx stakeibc liquid-stake 100000 uatom --keyring-backend test --from val1 -y --chain-id $STRIDE_CHAIN
# exit

# clear balances
$STRIDE_CMD tx stakeibc clear-balance GAIA 66 channel-0 --keyring-backend test --from val1 --chain-id $STRIDE_CHAIN
exit

# redeem stake
# amt_to_redeem=5
# $STRIDE_CMD tx stakeibc redeem-stake $amt_to_redeem GAIA $GAIA_RECEIVER_ACCT \
#     --from val1 --keyring-backend test --chain-id $STRIDE_CHAIN -y
# exit

EPOCH=9
SENDER_ACCT=stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7
$STRIDE_CMD tx stakeibc claim-undelegated-tokens GAIA $EPOCH $SENDER_ACCT --from val1 --keyring-backend test --chain-id $STRIDE_CHAIN -y
