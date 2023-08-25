#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

liquid_staked_address="cosmos1x92tnm6pfkl3gsfy0rfaez5myq5zh99aek2jmd"
validator_address="cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p"
delegation_account="cosmos1wdplq6qjh2xruc7qqagma9ya665q6qhcwju3ng"
stride_address="stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7"

###### Transactions
# echo "Delegate"
# $GAIA_MAIN_CMD tx staking delegate $validator_address 5000000uatom --from staker -y | TRIM_TX && echo ""
# sleep 5

# echo "Tokenize to liquid staker:"
# $GAIA_MAIN_CMD tx staking tokenize-share $validator_address 3000000uatom $liquid_staked_address --from staker -y --gas auto | TRIM_TX && echo ""
# sleep 5

# echo "Tokenize to delegation account:"
# $GAIA_MAIN_CMD tx staking tokenize-share $validator_address 1000000uatom $delegation_account --from staker -y --gas auto | TRIM_TX && echo ""
# sleep 5

# echo "Withdraw Rewards from record:"
# $GAIA_MAIN_CMD tx distribution withdraw-tokenize-share-rewards 2 --from staker -y | TRIM_TX
# sleep 5

# echo "Redeem Tokens from User:"
# $GAIA_MAIN_CMD tx staking redeem-tokens 1000000cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p/1 --from staker -y --gas auto | TRIM_TX
# sleep 5

# echo "Redeem Tokens from Delegation Account:"
# $GAIA_MAIN_CMD tx staking redeem-tokens 1000000cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p/1 --from grev1 -y --gas auto | TRIM_TX
# sleep 5

# echo "Send Token:"
# $GAIA_MAIN_CMD tx bank send $liquid_staked_address $delegation_account 1000000cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p/1 --from staker -y | TRIM_TX
# sleep 5

# echo "IBC Transfer:"
# $GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $stride_address 3000000cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p/1 --from staker -y | TRIM_TX
# sleep 5

# echo "Transfer Rewards:"
# $GAIA_MAIN_CMD tx staking transfer-tokenize-share-record 1 $delegation_account --from staker -y --gas auto | TRIM_TX
# sleep 5


######## Queries
# echo "Validator Shares:"
# $GAIA_MAIN_CMD q staking validators && echo ""

# echo "Delegations:"
# $GAIA_MAIN_CMD q staking delegations $liquid_staked_address && echo ""

# echo "Rewards:"
# $GAIA_MAIN_CMD q distribution rewards $liquid_staked_address $validator_address && echo ""

# echo "User Bank balance:"
# $GAIA_MAIN_CMD q bank balances $liquid_staked_address && echo ""

# echo "User Tokenized shares:"
# $GAIA_MAIN_CMD q distribution tokenize-share-record-rewards $liquid_staked_address && echo ""

# echo "Delegation Account Bank balance:"
# $GAIA_MAIN_CMD q bank balances $delegation_account && echo ""

# echo "Delegation Account Tokenized shares:"
# $GAIA_MAIN_CMD q distribution tokenize-share-record-rewards $delegation_account && echo ""