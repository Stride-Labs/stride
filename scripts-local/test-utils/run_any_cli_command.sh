### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source ${SCRIPT_DIR}/../account_vars.sh


# transfer tokens to stride
# $GAIA_CMD q tx F382BB9C7B9970C41F0BE5F04CB59B85A68CF360949A97C599DFA92A80CAD5D0
# exit
# $GAIA_CMD tx ibc-transfer transfer transfer channel-0 stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7 10000uatom --from gval1 --chain-id GAIA -y --keyring-backend test
# exit

# check val1 balances
# $STRIDE_CMD keys list
# exit

# $STRIDE_CMD q bank balances stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7
# exit

# liquid stake
# $STRIDE_CMD tx stakeibc liquid-stake 10000 uatom --keyring-backend test --from val1 -y --chain-id $STRIDE_CHAIN
# exit

# redeem stake
# $STRIDE_CMD q tx B3C8E62837FCF9835EB131386A5F7FDE92A20AF4AB0C46E5495B3CDB9F6CF3C1
# exit

amt_to_redeem=3
$STRIDE_CMD tx stakeibc redeem-stake $amt_to_redeem GAIA $GAIA_RECEIVER_ACCT \
    --from val1 --keyring-backend test --chain-id $STRIDE_CHAIN -y

exit
