#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

AIRDROP_NAME="sttia"

# --------------------------------------------------------------------------------

# echo -e "\n>>> Query Airdrop..."
# $STRIDE_MAIN_CMD q airdrop airdrop $AIRDROP_NAME
# sleep 3

# echo -e "\n>>> Update Airdrop (claim penalty to 25%)..."
# $STRIDE_MAIN_CMD tx airdrop update-airdrop $AIRDROP_NAME \
#   --claim-type-deadline-date "2024-08-01" \
#   --clawback-date "2024-11-01" \
#   --distribution-address stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8 \
#   --distribution-end-date "2024-10-01" \
#   --distribution-start-date "2024-07-01" \
#   --early-claim-penalty "0.25" \
#   --reward-denom ustrd \
#  --from admin -y --gas 1000000
# sleep 5

# echo -e "\n>>> Query Airdrop..."
# $STRIDE_MAIN_CMD q airdrop airdrop $AIRDROP_NAME


# --------------------------------------------------------------------------------

# echo -e "\n>>> Query Airdrop Allocations..."
# $STRIDE_MAIN_CMD q airdrop all-allocations $AIRDROP_NAME
# sleep 3

# echo -e "\n>>> Adding Allocations..."
# $STRIDE_MAIN_CMD tx airdrop add-allocations $AIRDROP_NAME ./dockernet/scripts/airdrop/allocations.csv \
#  --from admin -y --gas 1000000 | TRIM_TX
# sleep 5

# echo -e "\n>>> Query Airdrop..."
# $STRIDE_MAIN_CMD q airdrop airdrop $AIRDROP_NAME
# echo -e "\n>>> Query Airdrop Allocations..."
# $STRIDE_MAIN_CMD q airdrop all-allocations $AIRDROP_NAME

# --------------------------------------------------------------------------------

# echo -e "\n>>> Query Airdrop Allocation to stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7..."
# $STRIDE_MAIN_CMD q airdrop user-allocation $AIRDROP_NAME stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7
# sleep 3

# # echo -e "\n>>> Update User Allocations..."
# $STRIDE_MAIN_CMD tx airdrop update-user-allocation $AIRDROP_NAME "stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7" ./dockernet/scripts/airdrop/allocations_v2.csv \
#  --from admin -y --gas 1000000 | TRIM_TX
# sleep 5

# echo -e "\n>>> Query Airdrop Allocation for user stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7; first few days should be 999999..."
# $STRIDE_MAIN_CMD q airdrop user-allocation $AIRDROP_NAME stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7

# --------------------------------------------------------------------------------

# echo -e "\n>>> Query Airdrop Allocation to stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7..."
# $STRIDE_MAIN_CMD q airdrop user-allocation $AIRDROP_NAME stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7
# sleep 3

# # echo -e "\n>>> Adding Allocations..."
# $STRIDE_MAIN_CMD tx airdrop link-addresses $AIRDROP_NAME "stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7" "dym1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhr6zxkau" \
#  --from admin -y --gas 1000000 | TRIM_TX
# sleep 5

# echo -e "\n>>> Query Airdrop Allocation to stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7..."
# $STRIDE_MAIN_CMD q airdrop user-allocation $AIRDROP_NAME stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7
# sleep 3

# echo -e "\n>>> Query Airdrop Allocation to stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7..."
# $STRIDE_MAIN_CMD q airdrop user-allocation $AIRDROP_NAME dym1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhr6zxkau
# sleep 3