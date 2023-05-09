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
# $LSM_MAIN_CMD tx staking delegate $validator_address 5000000stake --from staker -y | TRIM_TX && echo ""
# sleep 5

# echo "Tokenize to liquid staker:"
# $LSM_MAIN_CMD tx staking tokenize-share $validator_address 3000000stake $liquid_staked_address --from staker -y --gas auto | TRIM_TX && echo ""
# sleep 5

# echo "Tokenize to delegation account:"
# $LSM_MAIN_CMD tx staking tokenize-share $validator_address 1000000stake $delegation_account --from staker -y --gas auto | TRIM_TX && echo ""
# sleep 5

# echo "Withdraw Rewards from record:"
# $LSM_MAIN_CMD tx distribution withdraw-tokenize-share-rewards 2 --from staker -y | TRIM_TX
# sleep 5

# echo "Redeem Tokens from User:"
# $LSM_MAIN_CMD tx staking redeem-tokens 1000000cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p/1 --from staker -y --gas auto | TRIM_TX
# sleep 5

# echo "Redeem Tokens from Delegation Account:"
# $LSM_MAIN_CMD tx staking redeem-tokens 1000000cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p/1 --from lsrev1 -y --gas auto | TRIM_TX
# sleep 5

# echo "Send Token:"
# $LSM_MAIN_CMD tx bank send $liquid_staked_address $delegation_account 1000000cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p/1 --from staker -y | TRIM_TX
# sleep 5

# echo "IBC Transfer:"
# $LSM_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $stride_address 3000000cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p/1 --from staker -y | TRIM_TX
# sleep 5

# echo "Transfer Rewards:"
# $LSM_MAIN_CMD tx staking transfer-tokenize-share-record 1 $delegation_account --from staker -y --gas auto | TRIM_TX
# sleep 5


######## Queries
# echo "Validator Shares:"
# $LSM_MAIN_CMD q staking validators && echo ""

# echo "Delegations:"
# $LSM_MAIN_CMD q staking delegations $liquid_staked_address && echo ""

# echo "Rewards:"
# $LSM_MAIN_CMD q distribution rewards $liquid_staked_address $validator_address && echo ""

# echo "User Bank balance:"
# $LSM_MAIN_CMD q bank balances $liquid_staked_address && echo ""

# echo "User Tokenized shares:"
# $LSM_MAIN_CMD q distribution tokenize-share-record-rewards $liquid_staked_address && echo ""

# echo "Delegation Account Bank balance:"
# $LSM_MAIN_CMD q bank balances $delegation_account && echo ""

# echo "Delegation Account Tokenized shares:"
# $LSM_MAIN_CMD q distribution tokenize-share-record-rewards $delegation_account && echo ""