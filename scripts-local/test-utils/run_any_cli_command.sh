### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source ${SCRIPT_DIR}/../account_vars.sh


# $GAIA_CMD tx ibc-transfer transfer transfer channel-0 stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7 100000uatom --from gval1 --chain-id GAIA -y --keyring-backend test
# $GAIA_CMD q tx DC7D62F7D9AD1F8CBE48F95C1E3DADDBAD7FE85DFB6979B458A086FE5ED56A8C
# exit

# $STRIDE_CMD q stakeibc show-host-zone OSMO
$OSMO_CMD q bank balances osmo1cx04p5974f8hzh2lqev48kjrjugdxsxy7mzrd0eyweycpr90vk8q8d6f3h
exit

# $STRIDE_CMD q tx 48DF6A5FA859BC4DE0A5D891798DEBA36DF7A61561090F635B823B38557E5C41
# exit

# liquid stake
# $STRIDE_CMD tx stakeibc liquid-stake 1000 uatom --keyring-backend test --from val1 -y --chain-id $STRIDE_CHAIN
# exit

# redeem stake
# amt_to_redeem=5
# $STRIDE_CMD tx stakeibc redeem-stake $amt_to_redeem GAIA $GAIA_RECEIVER_ACCT \
#     --from val1 --keyring-backend test --chain-id $STRIDE_CHAIN -y
# exit

EPOCH=9
SENDER_ACCT=stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7
$STRIDE_CMD tx stakeibc claim-undelegated-tokens GAIA $EPOCH $SENDER_ACCT --from val1 --keyring-backend test --chain-id $STRIDE_CHAIN -y
