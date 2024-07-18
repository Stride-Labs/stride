#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

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

current_time_with_minute_offset() {
	minutes=$1
	if [[ "$OS" == "Darwin" ]]; then
		date -u -v+"${minutes}M" +"%Y-%m-%dT%H:%M:%S"
	else
		date -u +"%Y-%m-%dT%H:%M:%S" -d "$minutes minutes"
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
$STRIDE_MAIN_CMD tx bank send val1 $claimer_1_address 1ustrd --from val1 -y | TRIM_TX
sleep 3
$STRIDE_MAIN_CMD tx bank send val1 $claimer_2_address 1ustrd --from val1 -y | TRIM_TX
sleep 3
$STRIDE_MAIN_CMD tx bank send val1 $claimer_3_address 1ustrd --from val1 -y | TRIM_TX
sleep 3

echo -e "\n>>> Creating airdrop..."
$STRIDE_MAIN_CMD tx airdrop create-airdrop $AIRDROP_NAME \
  	--distribution-start-date  $(current_time_with_minute_offset 1) \
	--distribution-end-date    $(current_time_with_minute_offset 20) \
	--clawback-date            $(current_time_with_minute_offset 30) \
	--claim-type-deadline-date $(current_time_with_minute_offset 10) \
	--early-claim-penalty      0.5 \
	--distributor-address      $distributor_address \
	--allocator-address        $allocator_address \
	--linker-address           $linker_address \
	--from admin -y | TRIM_TX
sleep 3

echo -e "\n>>> Adding allocations..."
$STRIDE_MAIN_CMD tx airdrop add-allocations $AIRDROP_NAME ${SCRIPT_DIR}/allocations.csv \
    --from allocator -y --gas 1000000 | TRIM_TX 
sleep 3

echo -e "\n>>> Airdrops:"
$STRIDE_MAIN_CMD q airdrop airdrops

echo -e "\n>>> Allocations:"
$STRIDE_MAIN_CMD q airdrop all-allocations $AIRDROP_NAME | head -n 10 
echo "..."