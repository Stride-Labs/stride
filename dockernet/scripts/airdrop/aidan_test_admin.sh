#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/setup.sh

# This script covers the Admin functions and Admin addresses sections in the airdrop spec testing table

# Uncomment to run partial script after first run
# source ${SCRIPT_DIR}/../../config.sh
# OS=$(uname)
# date_with_offset() {
# 	days=$1
# 	if [[ "$OS" == "Darwin" ]]; then
# 		date -u -v+"${days}d" +"%Y-%m-%dT%H:%M:%S"
# 	else
# 		date -u +"%Y-%m-%dT%H:%M:%S" -d "$days days"
# 	fi
# }
# AIRDROP_NAME="sttia"
# distributor_address=$($STRIDE_MAIN_CMD keys show distributor -a)
# allocator_address=$($STRIDE_MAIN_CMD keys show allocator -a)
# linker_address=$($STRIDE_MAIN_CMD keys show linker -a)


# Create airdrop
# -- DONE in setup.sh
# Create existing airdrop again should error
echo -e "\n>>> ADMIN TEST 1"
echo -e "\n>>> Creating airdrop..."
$STRIDE_MAIN_CMD tx airdrop create-airdrop $AIRDROP_NAME \
  	--distribution-start-date  $(date_with_offset 0) \
	--distribution-end-date    $(date_with_offset 19) \
	--clawback-date            $(date_with_offset 29) \
	--claim-type-deadline-date $(date_with_offset 9) \
	--early-claim-penalty      0.5 \
	--distributor-address      $distributor_address \
	--allocator-address        $allocator_address \
	--linker-address           $linker_address \
	--from admin -y | TRIM_TX


echo -e "\n>>> Airdrop creation should error above..."

# Create allocations
# -- DONE in setup.sh

echo -e "\n>>> ADMIN TEST 2"
echo -e "\n>>> Query Airdrop Allocations..."
$STRIDE_MAIN_CMD q airdrop all-allocations $AIRDROP_NAME
sleep 3

# Create allocation allocation already exists should error
echo -e "\n>>> ADMIN TEST 3"
echo -e "\n>>> Create allocation allocation already exists should ERROR..."
$STRIDE_MAIN_CMD tx airdrop add-allocations $AIRDROP_NAME ./dockernet/scripts/airdrop/allocations.csv \
 --from allocator -y --gas 1000000 | TRIM_TX
sleep 5

# Create allocation when airdrop doesn’t exist should error
echo -e "\n>>> Create allocation when airdrop doesn’t exist should ERROR..."
$STRIDE_MAIN_CMD tx airdrop add-allocations 'dne-airdrop' ./dockernet/scripts/airdrop/allocations.csv \
 --from allocator -y --gas 1000000 | TRIM_TX
sleep 5

# Create allocations with incorrect length should error
echo -e "\n>>> ADMIN TEST 4"
echo -e "\n>>> Create allocations with incorrect length should ERROR..."
$STRIDE_MAIN_CMD tx airdrop add-allocations $AIRDROP_NAME ./dockernet/scripts/airdrop/allocations_wrong_length.csv \
 --from allocator -y --gas 1000000 | TRIM_TX
sleep 5

# Update allocation
echo -e "\n>>> ADMIN TEST 5"
echo -e "\n>>> Update allocation..."
# allocation before
$STRIDE_MAIN_CMD q airdrop user-allocation $AIRDROP_NAME stride1qtzlx93h8xlmej42pjqyez6yp9nscfgxsmtt59 | head -n 30
$STRIDE_MAIN_CMD tx airdrop update-user-allocation $AIRDROP_NAME ./dockernet/scripts/airdrop/allocations_update.csv \
 --from allocator -y --gas 1000000 | TRIM_TX
sleep 5
# allocation after
$STRIDE_MAIN_CMD q airdrop user-allocation $AIRDROP_NAME stride1qtzlx93h8xlmej42pjqyez6yp9nscfgxsmtt59 | head -n 30


# # Only admin can create airdrop 
echo -e "\n>>> ADMIN TEST 6"
echo -e "\n>>> Only admin can create airdrop..."
$STRIDE_MAIN_CMD tx airdrop create-airdrop 'new-airdrop' \
  	--distribution-start-date  $(date_with_offset 0) \
	--distribution-end-date    $(date_with_offset 19) \
	--clawback-date            $(date_with_offset 29) \
	--claim-type-deadline-date $(date_with_offset 9) \
	--early-claim-penalty      0.5 \
	--distributor-address      $distributor_address \
	--allocator-address        $allocator_address \
	--linker-address           $linker_address \
	--from allocator -y | TRIM_TX
sleep 5

# Only allocator can add or update allocations
echo -e "\n>>> ADMIN TEST 7"
echo -e "\n>>> Only allocator can add or update allocations..."
$STRIDE_MAIN_CMD tx airdrop update-user-allocation $AIRDROP_NAME ./dockernet/scripts/airdrop/allocations_update.csv \
 --from admin -y --gas 1000000 | TRIM_TX
sleep 5
$STRIDE_MAIN_CMD tx airdrop add-allocations $AIRDROP_NAME ./dockernet/scripts/airdrop/allocations_wrong_length.csv \
 --from admin -y --gas 1000000 | TRIM_TX
sleep 5

# Only linker can link
echo -e "\n>>> ADMIN TEST 8"
echo -e "\n>>> Only linker can link"
$STRIDE_MAIN_CMD tx airdrop link-addresses $AIRDROP_NAME stride13k0vj64yr3dxq4e24v5s2ptqmnxmyl7xn5pz7q dym1np5x8s6lufkv8ghu8lzj5xtlgae5pwl8y8ne6x --from linker -y --gas 1000000 | TRIM_TX
sleep 5
$STRIDE_MAIN_CMD tx airdrop link-addresses $AIRDROP_NAME stride13k0vj64yr3dxq4e24v5s2ptqmnxmyl7xn5pz7q dym1np5x8s6lufkv8ghu8lzj5xtlgae5pwl8y8ne6x --from admin -y --gas 1000000 | TRIM_TX
sleep 5