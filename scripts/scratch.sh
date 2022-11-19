### IBC TRANSFER
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/vars.sh

# stride airdrop: stride1thl8e7smew8q7jrz8at4f64wrjjl8mwan3nc4l
# other address: stride12cv6uvcrn0eca8l6lszhyn50hh9dayfg5sukcr
# stride distributor: stride1thl8e7smew8q7jrz8at4f64wrjjl8mwan3nc4l
# gaia distributor: stride104az7rd5yh3p8qn4ary8n3xcwquuwgee4vnvvc

# airdrop test

# prep
# - large weights for addresses
# - 2 airdrops
# - 2 addresses

# - setup upgrade with binary

# airdrop

# 1) make start-docker && sleep 5 && bash scripts/upgrades/submit_upgrade.sh && bash scripts/scratch.sh
# 2) scratch - fund accounts with ustrd and uatom
# 3) scratch - claim free portion (check balances)
# 4) scratch - stake (check balances)
# 5) scratch - liquid stake (check balances)


## IBC ATOM from GAIA to STRIDE
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 stride1thl8e7smew8q7jrz8at4f64wrjjl8mwan3nc4l 1000000uatom --from ${GAIA_VAL_PREFIX}1 -y 
sleep 10
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 stride12cv6uvcrn0eca8l6lszhyn50hh9dayfg5sukcr 1000000uatom --from ${GAIA_VAL_PREFIX}1 -y 
sleep 3

# send funds to airdrop addresses
$STRIDE_MAIN_CMD tx bank send val1 stride1thl8e7smew8q7jrz8at4f64wrjjl8mwan3nc4l 3000ustrd --from val1 --chain-id STRIDE -y --keyring-backend test
sleep 3
$STRIDE_MAIN_CMD tx bank send val1 stride12cv6uvcrn0eca8l6lszhyn50hh9dayfg5sukcr 3000ustrd --from val1 --chain-id STRIDE -y --keyring-backend test
sleep 3
$STRIDE_MAIN_CMD tx bank send val1 stride1thl8e7smew8q7jrz8at4f64wrjjl8mwan3nc4l 3000ustrd --from val1 --chain-id STRIDE -y --keyring-backend test
sleep 3
$STRIDE_MAIN_CMD tx bank send val1 stride104az7rd5yh3p8qn4ary8n3xcwquuwgee4vnvvc 3000ustrd --from val1 --chain-id STRIDE -y --keyring-backend test
sleep 3
exit

# Check balances
$STRIDE_MAIN_CMD q bank balances stride1thl8e7smew8q7jrz8at4f64wrjjl8mwan3nc4l
exit

# Claim free portion
build/strided tx claim claim-free-amount stride --from airdrop --home scripts/state/stride1 --node http://localhost:26657
# Check balances
$STRIDE_MAIN_CMD q bank balances stride1thl8e7smew8q7jrz8at4f64wrjjl8mwan3nc4l
exit

# Stake
build/strided tx staking delegate stridevaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrgpwsqm 1ustrd --from airdrop --home scripts/state/stride1 --node http://localhost:26657
sleep 3
# Liquid stake
build/strided tx stakeibc liquid-stake 100 ustrd --from airdrop --home scripts/state/stride1 --node http://localhost:26657
# Check balances
$STRIDE_MAIN_CMD q bank balances stride1thl8e7smew8q7jrz8at4f64wrjjl8mwan3nc4l
exit


# build/strided q bank balances stride16ea8j8mxvcy29w3jxuhvkculr4rg56mgkcwp6d --node http://localhost:26657


# build/strided q staking validators --node http://localhost:26657

# build/strided q bank balances stride1thl8e7smew8q7jrz8at4f64wrjjl8mwan3nc4l --node http://localhost:26657
# build/strided q bank balances stride16ea8j8mxvcy29w3jxuhvkculr4rg56mgkcwp6d --node http://localhost:26657
# build/strided q tx 748F78633A641021F06F379288E4D45119F014B300B620F9877F4E2F10257FD5 --node http://localhost:26657

