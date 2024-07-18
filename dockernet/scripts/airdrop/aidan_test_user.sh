#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

# source ${SCRIPT_DIR}/setup.sh

# This script covers the Success cases and Error cases sections in the airdrop spec testing table

AIRDROP_NAME="sttia"
OS=$(uname)

# stride1wxv8jnusl5s2che9kp33jrsxz3kxcwm6fgfs0z
DISTRIBUTOR_MNEMONIC="wire elephant soldier improve million minor image identify analyst black kangaroo word dose run olive heavy usual sound copy diamond market accident other clean"
# stride1v0t265yce4pwdvmzs4ew0ll4fnk2t6updfg6hf
ALLOCATOR_MNEMONIC="board stamp journey usage pen warfare burst that detect rain solid skill affair cactus forward insect video bitter goddess tank syrup hunt pond flat"
# stride1krn5q0s83qfx9v67awsj7leaql99tmyydcukhl
LINKER_MNEMONIC="add special provide zebra gasp above tube shield judge vapor tortoise snow receive vibrant couch true tide snack goat fee risk viable coil mutual"

# stride1qtzlx93h8xlmej42pjqyez6yp9nscfgxsmtt59
CLAIMER_1_MNEMONIC="breeze reason effort latin mask orbit ball raw security gown category royal copper scheme fiction flame few wise siege car text snake famous render"
# stride183g7tx3u4lmtwv7ph9fpq862e6dyapamexywru
CLAIMER_2_MNEMONIC="glance trigger upgrade keep nature glad wreck snake grief trap utility curtain bracket large drama ridge loud token service idea smart crisp flavor carpet"
# stride13k0vj64yr3dxq4e24v5s2ptqmnxmyl7xn5pz7q
CLAIMER_3_MNEMONIC="pet garlic cram security clock element truth soda stomach ugly you dress narrow black space grab concert cancel depend crawl corn worry miss submit"

LINKED_CLAIMER_ADDRESS="dym1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhr6zxkau"


current_time_with_minute_offset() {
	minutes=$1
	if [[ "$OS" == "Darwin" ]]; then
		date -u -v+"${minutes}M" +"%Y-%m-%dT%H:%M:%S"
	else
		date -u +"%Y-%m-%dT%H:%M:%S" -d "$minutes minutes"
	fi
}

echo ">>> Creating admin accounts"
echo $DISTRIBUTOR_MNEMONIC | $STRIDE_MAIN_CMD keys add distributor --recover -- | grep -E "address|name"
echo $ALLOCATOR_MNEMONIC | $STRIDE_MAIN_CMD keys add allocator --recover | grep -E "address|name"
echo $LINKER_MNEMONIC | $STRIDE_MAIN_CMD keys add linker --recover | grep -E "address|name"

echo ">>> Creating claimer accounts"
echo $CLAIMER_1_MNEMONIC | $STRIDE_MAIN_CMD keys add claimer1 --recover | grep -E "address|name"
echo $CLAIMER_2_MNEMONIC | $STRIDE_MAIN_CMD keys add claimer2 --recover | grep -E "address|name"
echo $CLAIMER_3_MNEMONIC | $STRIDE_MAIN_CMD keys add claimer3 --recover | grep -E "address|name"

distributor_address=$($STRIDE_MAIN_CMD keys show distributor -a)
allocator_address=$($STRIDE_MAIN_CMD keys show allocator -a)
linker_address=$($STRIDE_MAIN_CMD keys show linker -a)

claimer_1_address=$($STRIDE_MAIN_CMD keys show claimer1 -a)
claimer_2_address=$($STRIDE_MAIN_CMD keys show claimer2 -a)
claimer_3_address=$($STRIDE_MAIN_CMD keys show claimer3 -a)

echo -e "\n>>> Funding admin accounts..."
$STRIDE_MAIN_CMD tx bank send val1 $distributor_address 5000000000ustrd --from val1 -y | TRIM_TX
sleep 3
$STRIDE_MAIN_CMD tx bank send val1 $allocator_address 1ustrd --from val1 -y | TRIM_TX
sleep 3
$STRIDE_MAIN_CMD tx bank send val1 $linker_address 1ustrd --from val1 -y | TRIM_TX
sleep 3

echo -e "\n>>> Funding claimer accounts..."
$STRIDE_MAIN_CMD tx bank send val1 $claimer_1_address 1ustrd --from val1 -y | TRIM_TX
sleep 3
$STRIDE_MAIN_CMD tx bank send val1 $claimer_2_address 1ustrd --from val1 -y | TRIM_TX
sleep 3
$STRIDE_MAIN_CMD tx bank send val1 $claimer_3_address 1ustrd --from val1 -y | TRIM_TX
sleep 3

echo -e "\n>>> Creating airdrop..."
$STRIDE_MAIN_CMD tx airdrop create-airdrop $AIRDROP_NAME \
  	--distribution-start-date  $(current_time_with_minute_offset 1) \
	--distribution-end-date    $(current_time_with_minute_offset 4) \
	--clawback-date            $(current_time_with_minute_offset 5) \
	--claim-type-deadline-date $(current_time_with_minute_offset 3) \
	--early-claim-penalty      0.5 \
	--distributor-address      $distributor_address \
	--allocator-address        $allocator_address \
	--linker-address           $linker_address \
	--from admin -y | TRIM_TX
sleep 3

echo -e "\n>>> Adding allocations..."
$STRIDE_MAIN_CMD tx airdrop add-allocations $AIRDROP_NAME ${SCRIPT_DIR}/aidan_allocations.csv \
    --from allocator -y --gas 1000000 | TRIM_TX 
sleep 3

echo -e "\n>>> Airdrops:"
$STRIDE_MAIN_CMD q airdrop airdrops

echo -e "\n>>> Allocations:"
$STRIDE_MAIN_CMD q airdrop all-allocations $AIRDROP_NAME | head -n 10 
echo "..."


# ERROR CASES
if false; then
    echo -e "\n>>> Testing ERROR cases"
    # NOTE: aidan_allocations.csv must have 3 allocations per user
    # USING
    # echo -e "\n>>> Creating airdrop..."
    # $STRIDE_MAIN_CMD tx airdrop create-airdrop $AIRDROP_NAME \
    #   	--distribution-start-date  $(current_time_with_minute_offset 1) \
    # 	--distribution-end-date    $(current_time_with_minute_offset 3) \
    # 	--clawback-date            $(current_time_with_minute_offset 4) \
    # 	--claim-type-deadline-date $(current_time_with_minute_offset 2) \
    # 	--early-claim-penalty      0.5 \
    # 	--distributor-address      $distributor_address \
    # 	--allocator-address        $allocator_address \
    # 	--linker-address           $linker_address \
    # 	--from admin -y | TRIM_TX
    # sleep 3
    echo -e "\n>>> Claiming daily before the airdrop started"
    $STRIDE_MAIN_CMD tx airdrop claim-daily $AIRDROP_NAME --from claimer1 -y --gas 1000000 | TRIM_TX
    sleep 5

    # AIRDROP NOT STARTED
    echo -e "\n>>> Claiming early before the airdrop started"
    $STRIDE_MAIN_CMD tx airdrop claim-early $AIRDROP_NAME --from claimer1 -y --gas 1000000 | TRIM_TX
    sleep 60

    # AIRDROP STARTED
    echo -e "\n>>> claimer1 claims early (expected success)"
    $STRIDE_MAIN_CMD tx airdrop claim-early $AIRDROP_NAME --from claimer1 -y --gas 1000000 | TRIM_TX
    sleep 5
    echo -e "\n>>> Claiming daily after already claiming early (expected failure)"
    $STRIDE_MAIN_CMD tx airdrop claim-daily $AIRDROP_NAME --from claimer1 -y --gas 1000000 | TRIM_TX
    sleep 5
    echo -e "\n>>> Claiming early after already claiming early (expected failure)"
    $STRIDE_MAIN_CMD tx airdrop claim-early $AIRDROP_NAME --from claimer1 -y --gas 1000000 | TRIM_TX
    sleep 5

    echo -e "\n>>> claimer2 claims daily (expected success)"
    $STRIDE_MAIN_CMD tx airdrop claim-early $AIRDROP_NAME --from claimer2 -y --gas 1000000 | TRIM_TX
    sleep 5
    echo -e "\n>>> Claiming daily after already claiming all rewards daily (expected failure)"
    $STRIDE_MAIN_CMD tx airdrop claim-early $AIRDROP_NAME --from claimer2 -y --gas 1000000 | TRIM_TX
    sleep 45

    # AFTER DECISION DATE
    echo -e "\n>>> Claiming early after the decision date"
    $STRIDE_MAIN_CMD tx airdrop claim-early $AIRDROP_NAME --from claimer3 -y --gas 1000000 | TRIM_TX
    sleep 5

    # AFTER AIRDROP ENDS ENTIRELY
    echo -e "\n>>> Claiming daily after the clawback date"
    $STRIDE_MAIN_CMD tx airdrop claim-daily $AIRDROP_NAME --from claimer1 -y --gas 1000000 | TRIM_TX
    sleep 5

    # ALL CHECKED MANUALLY - OK
fi

# SUCCESS CASES
# NOTE: aidan_allocations.csv must have 4 allocations per user
if false; then
    echo -e "\n>>> Testing SUCCESS cases ROUND 1"
    # ROUND 1
    # USING
    # $STRIDE_MAIN_CMD tx airdrop create-airdrop $AIRDROP_NAME \
    #   	--distribution-start-date  $(current_time_with_minute_offset 1) \
    # 	--distribution-end-date    $(current_time_with_minute_offset 4) \
    # 	--clawback-date            $(current_time_with_minute_offset 5) \
    # 	--claim-type-deadline-date $(current_time_with_minute_offset 3) \
    # 	--early-claim-penalty      0.5 \
    # 	--distributor-address      $distributor_address \
    # 	--allocator-address        $allocator_address \
    # 	--linker-address           $linker_address \
    # 	--from admin -y | TRIM_TX
    # sleep 3
    # Claiming daily each day
    # -claimer1 claims daily on days 1,2,3
    # Claiming daily with some days skipped
    # -claimer2 claims on days 1,3
    # Claiming daily on the last day of distribution
    # -claimer3 claims on day 3
    # Claiming daily on the last day before clawback
    # -claimer3 claims on day 3
    sleep 60
    # DAY 1 - AIRDROP STARTS
    echo -e "\n>>> claimer1 claims daily (expected success)"
    $STRIDE_MAIN_CMD tx airdrop claim-daily $AIRDROP_NAME --from claimer1 -y --gas 1000000 | TRIM_TX
    sleep 5
    echo -e "\n>>> claimer2 claims daily (expected success)"
    $STRIDE_MAIN_CMD tx airdrop claim-daily $AIRDROP_NAME --from claimer2 -y --gas 1000000 | TRIM_TX
    sleep 5
    sleep 55
    # DAY 2 - AIRDROP CONTINUES
    echo -e "\n>>> claimer1 claims daily (expected success)"
    $STRIDE_MAIN_CMD tx airdrop claim-daily $AIRDROP_NAME --from claimer1 -y --gas 1000000 | TRIM_TX
    sleep 5
    sleep 55
    # DAY 3 - AIRDROP CONTINUES - EARLY CLAIM DEADLINE
    echo -e "\n>>> claimer1 claims daily (expected success)"
    $STRIDE_MAIN_CMD tx airdrop claim-daily $AIRDROP_NAME --from claimer1 -y --gas 1000000 | TRIM_TX
    sleep 5
    echo -e "\n>>> claimer2 claims daily (expected success)"
    $STRIDE_MAIN_CMD tx airdrop claim-daily $AIRDROP_NAME --from claimer2 -y --gas 1000000 | TRIM_TX
    sleep 5
    echo -e "\n>>> claimer3 claims daily (expected success)"
    $STRIDE_MAIN_CMD tx airdrop claim-daily $AIRDROP_NAME --from claimer3 -y --gas 1000000 | TRIM_TX
    sleep 5
    # DAY 4 - AIRDROP END
    # DAY 5 - AIRDROP CLAWBACK
fi


if false; then
    echo -e "\n>>> Testing SUCCESS cases ROUND 2"
    # ROUND 2
    # Claiming early immediately from the start
    # -claimer1 claims early on day 1
    # Claiming early after claiming daily for the first few days
    # -claimer2 claims daily on day 1, early on day 2
    # Linking then claiming daily
    # -claimer3 links on day 1, then claims daily
    sleep 60
    # DAY 1 - AIRDROP STARTS
    echo -e "\n>>> claimer1 claims daily (expected success)"
    $STRIDE_MAIN_CMD tx airdrop claim-early $AIRDROP_NAME --from claimer1 -y --gas 1000000 | TRIM_TX
    sleep 5
    echo -e "\n>>> claimer2 claims daily (expected success)"
    $STRIDE_MAIN_CMD tx airdrop claim-daily $AIRDROP_NAME --from claimer2 -y --gas 1000000 | TRIM_TX
    sleep 5
    echo -e "\n>>> claimer3 links (expected success)"
    $STRIDE_MAIN_CMD tx airdrop link-addresses $AIRDROP_NAME stride13k0vj64yr3dxq4e24v5s2ptqmnxmyl7xn5pz7q dym1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhr6zxkau --from linker -y --gas 1000000 | TRIM_TX
    sleep 5
    sleep 50
    # DAY 2 - AIRDROP CONTINUES
    echo -e "\n>>> claimer2 claims daily (expected success)"
    $STRIDE_MAIN_CMD tx airdrop claim-early $AIRDROP_NAME --from claimer2 -y --gas 1000000 | TRIM_TX
    sleep 5
    echo -e "\n>>> claimer3 claims daily (expected success)"
    $STRIDE_MAIN_CMD tx airdrop claim-daily $AIRDROP_NAME --from claimer3 -y --gas 1000000 | TRIM_TX
    sleep 5
    sleep 50
    # DAY 3 - AIRDROP CONTINUES - EARLY CLAIM DEADLINE
    # DAY 4 - AIRDROP END
    # DAY 5 - AIRDROP CLAWBACK
fi


if true; then
    echo -e "\n>>> Testing SUCCESS cases ROUND 3"
    # ROUND 3
    # Linking then claiming early
    # -claimer1 links then claims early, same day
    # Claiming daily then linking then claiming daily again
    # -claimer2 claims daily on day 1, links on day 2, claims daily on day 2
    # Claiming daily then linking then claiming early
    # -claimer3 claims daily on day 1, links on day 2, claims early on day 2
    sleep 60
    # DAY 1 - AIRDROP STARTS
    echo -e "\n>>> claimer1 links (expected success)"
    $STRIDE_MAIN_CMD tx airdrop link-addresses $AIRDROP_NAME stride1qtzlx93h8xlmej42pjqyez6yp9nscfgxsmtt59 dym1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhr6zxkau --from linker -y --gas 1000000 | TRIM_TX
    sleep 5
    echo -e "\n>>> claimer1 claims early (expected success)"
    $STRIDE_MAIN_CMD tx airdrop claim-early $AIRDROP_NAME --from claimer1 -y --gas 1000000 | TRIM_TX
    sleep 5

    echo -e "\n>>> claimer2 claims daily (expected success)"
    $STRIDE_MAIN_CMD tx airdrop claim-daily $AIRDROP_NAME --from claimer2 -y --gas 1000000 | TRIM_TX
    sleep 5
    echo -e "\n>>> claimer2 links (expected success)"
    $STRIDE_MAIN_CMD tx airdrop link-addresses $AIRDROP_NAME stride183g7tx3u4lmtwv7ph9fpq862e6dyapamexywru dym1qwuhp7hkesdtpx0gawewfx3dufww2p34lpcnca --from linker -y --gas 1000000 | TRIM_TX
    sleep 5

    echo -e "\n>>> claimer3 claims daily (expected success)"
    $STRIDE_MAIN_CMD tx airdrop claim-daily $AIRDROP_NAME --from claimer3 -y --gas 1000000 | TRIM_TX
    sleep 5

    sleep 35


    # DAY 2 - AIRDROP CONTINUES
    echo -e "\n>>> claimer2 claims daily (expected success)"
    $STRIDE_MAIN_CMD tx airdrop claim-early $AIRDROP_NAME --from claimer2 -y --gas 1000000 | TRIM_TX
    sleep 5

    cho -e "\n>>> claimer3 links (expected success)"
    $STRIDE_MAIN_CMD tx airdrop link-addresses $AIRDROP_NAME stride13k0vj64yr3dxq4e24v5s2ptqmnxmyl7xn5pz7q dym1lwqdqfpj4d6r98ahmrdw066t4cmvg5sz6jsns6 --from linker -y --gas 1000000 | TRIM_TX
    sleep 5
    echo -e "\n>>> claimer3 claims early (expected success)"
    $STRIDE_MAIN_CMD tx airdrop claim-early $AIRDROP_NAME --from claimer3 -y --gas 1000000 | TRIM_TX
    sleep 5
    sleep 50

    # DAY 3 - AIRDROP CONTINUES - EARLY CLAIM DEADLINE
    # DAY 4 - AIRDROP END
    # DAY 5 - AIRDROP CLAWBACK
fi










