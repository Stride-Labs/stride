#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

# NOTE: Before running this script, you must remove the line from init_chain.sh that changes the airdrop period length!

# Options: fresh-start, midway-before-deadline, midway-after-deadline, distribution-ended
STAGE="fresh-start"

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

# Same mnemonic as claimer 3
LINKED_CLAIMER_ADDRESS="dym1np5x8s6lufkv8ghu8lzj5xtlgae5pwl8y8ne6x"

current_time_minus_day_offset() {
	days=$1
	if [[ "$OS" == "Darwin" ]]; then
		date -u -v-"${days}d" +"%Y-%m-%dT00:00:00"
	else
		date -u +"%Y-%m-%dT00:00:00" -d "$days days ago"
	fi
}
current_time_plus_day_offset() {
	days=$1
	if [[ "$OS" == "Darwin" ]]; then
		date -u -v+"${days}d" +"%Y-%m-%dT00:00:00"
	else
		date -u +"%Y-%m-%dT00:00:00" -d "$days days"
	fi
}

echo ">>> Creating admin accounts..."
echo $DISTRIBUTOR_MNEMONIC | $STRIDE_MAIN_CMD keys add distributor --recover -- | grep -E "address|name"
echo $ALLOCATOR_MNEMONIC | $STRIDE_MAIN_CMD keys add allocator --recover | grep -E "address|name"
echo $LINKER_MNEMONIC | $STRIDE_MAIN_CMD keys add linker --recover | grep -E "address|name"

echo -e "\n>>> Creating claimer accounts..."
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
$STRIDE_MAIN_CMD tx bank send val1 $claimer_1_address 10000000ustrd --from val1 -y | TRIM_TX
sleep 3
$STRIDE_MAIN_CMD tx bank send val1 $claimer_2_address 10000000ustrd --from val1 -y | TRIM_TX
sleep 3
$STRIDE_MAIN_CMD tx bank send val1 $claimer_3_address 10000000ustrd --from val1 -y | TRIM_TX
sleep 3

if [[ "$STAGE" == "fresh-start" ]]; then 
	start_date=$(current_time_plus_day_offset 0)
	end_date=$(current_time_plus_day_offset 149)
	clawback_date=$(current_time_plus_day_offset 160)
	deadline_date=$(current_time_plus_day_offset 30)
fi

if [[ "$STAGE" == "midway-before-deadline" ]]; then 
	start_date=$(current_time_minus_day_offset 10)
	end_date=$(current_time_plus_day_offset 139)
	clawback_date=$(current_time_plus_day_offset 150)
	deadline_date=$(current_time_plus_day_offset 20)
fi

if [[ "$STAGE" == "midway-after-deadline" ]]; then 
	start_date=$(current_time_minus_day_offset 40)
	end_date=$(current_time_plus_day_offset 109)
	clawback_date=$(current_time_plus_day_offset 120)
	deadline_date=$(current_time_minus_day_offset 10)
fi

if [[ "$STAGE" == "distribution-ended" ]]; then 
	start_date=$(current_time_minus_day_offset 155)
	end_date=$(current_time_minus_day_offset 6)
	clawback_date=$(current_time_plus_day_offset 25)
	deadline_date=$(current_time_minus_day_offset 125)
fi

echo -e "\n>>> Creating airdrop..."
$STRIDE_MAIN_CMD tx airdrop create-airdrop $AIRDROP_NAME \
  	--distribution-start-date  $start_date \
	--distribution-end-date    $end_date \
	--clawback-date            $clawback_date \
	--claim-type-deadline-date $deadline_date \
	--early-claim-penalty      0.5 \
	--distributor-address      $distributor_address \
	--allocator-address        $allocator_address \
	--linker-address           $linker_address \
	--from admin -y | TRIM_TX
sleep 3

echo -e "\n>>> Adding allocations..."
$STRIDE_MAIN_CMD tx airdrop add-allocations $AIRDROP_NAME ${SCRIPT_DIR}/allocations_prod.csv \
    --from allocator -y --gas 1000000 | TRIM_TX 
sleep 3

echo -e "\n>>> Airdrops:"
$STRIDE_MAIN_CMD q airdrop airdrops

echo -e "\n>>> Allocations:"
$STRIDE_MAIN_CMD q airdrop all-allocations $AIRDROP_NAME | head -n 10 
echo "..."