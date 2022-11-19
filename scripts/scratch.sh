### IBC TRANSFER
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/vars.sh

# stride airdrop: stride19gvy00yat402jep7w6zdhssu4u09utkrflux5n
# oyster clump trial gossip canvas wool else express seat youth advance blood risk expand thing swamp soda mom attract ankle spot amused among gallery
# other address: stride1ju2eh2qskqgff83apn3u5y5exnh8lwht8t0aps
# mobile april mouse rocket poem return medal large puzzle come fashion dutch group fly surge pride loop innocent coyote record problem lab mother symptom

# stride distributor: stride1cpvl8yf848karqauyhr5jzw6d9n9lnuuu974ev
# gaia distributor: stride1fmh0ysk5nt9y2cj8hddms5ffj2dhys55xkkjwz

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


# ------------------------------------------------------------------------------------------------
# FUND ACCOUNTS
# ## IBC ATOM from GAIA to STRIDE
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 stride19gvy00yat402jep7w6zdhssu4u09utkrflux5n 1000000uatom --from ${GAIA_VAL_PREFIX}1 -y 
sleep 10
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 stride1ju2eh2qskqgff83apn3u5y5exnh8lwht8t0aps 1000000uatom --from ${GAIA_VAL_PREFIX}1 -y 
sleep 3

# send funds to airdrop addresses
$STRIDE_MAIN_CMD tx bank send val1 stride19gvy00yat402jep7w6zdhssu4u09utkrflux5n 3000ustrd --from val1 --chain-id STRIDE -y --keyring-backend test
sleep 3
$STRIDE_MAIN_CMD tx bank send val1 stride1ju2eh2qskqgff83apn3u5y5exnh8lwht8t0aps 3000ustrd --from val1 --chain-id STRIDE -y --keyring-backend test
sleep 3
$STRIDE_MAIN_CMD tx bank send val1 stride1cpvl8yf848karqauyhr5jzw6d9n9lnuuu974ev 3000ustrd --from val1 --chain-id STRIDE -y --keyring-backend test
sleep 3
$STRIDE_MAIN_CMD tx bank send val1 stride1fmh0ysk5nt9y2cj8hddms5ffj2dhys55xkkjwz 3000ustrd --from val1 --chain-id STRIDE -y --keyring-backend test
sleep 3
exit

# Check balances
$STRIDE_MAIN_CMD q bank balances stride19gvy00yat402jep7w6zdhssu4u09utkrflux5n
$STRIDE_MAIN_CMD q bank balances stride1ju2eh2qskqgff83apn3u5y5exnh8lwht8t0aps
$STRIDE_MAIN_CMD q bank balances stride1cpvl8yf848karqauyhr5jzw6d9n9lnuuu974ev
$STRIDE_MAIN_CMD q bank balances stride1fmh0ysk5nt9y2cj8hddms5ffj2dhys55xkkjwz
exit

# build/strided q tx C8DF243D9E44BCDF76421FEC2FB7AF9179637441A2E1448880840082586F8AAE --node http://localhost:26657
# exit

# ------------------------------------------------------------------------------------------------
# CLAIM AIRDROP
# Claim free portion
build/strided tx claim claim-free-amount --from airdrop-2 --node http://localhost:26657 --keyring-backend test --chain-id STRIDE --gas 300000
exit

# Check balances
$STRIDE_MAIN_CMD q bank balances stride1ju2eh2qskqgff83apn3u5y5exnh8lwht8t0aps
exit


# Stake
build/strided tx staking delegate stridevaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrgpwsqm 1ustrd --from airdrop-2 --node http://localhost:26657 --chain-id STRIDE --gas 600000
sleep 3
# Liquid stake
build/strided tx stakeibc liquid-stake 100 uatom --from airdrop-2 --node http://localhost:26657 --chain-id STRIDE --gas 600000
exit

# Check balances
$STRIDE_MAIN_CMD q bank balances stride1ju2eh2qskqgff83apn3u5y5exnh8lwht8t0aps
exit


# build/strided q bank balances stride16ea8j8mxvcy29w3jxuhvkculr4rg56mgkcwp6d --node http://localhost:26657


# build/strided q staking validators --node http://localhost:26657

# build/strided q bank balances stride1cpvl8yf848karqauyhr5jzw6d9n9lnuuu974ev --node http://localhost:26657
# build/strided q bank balances stride16ea8j8mxvcy29w3jxuhvkculr4rg56mgkcwp6d --node http://localhost:26657
# build/strided q tx 748F78633A641021F06F379288E4D45119F014B300B620F9877F4E2F10257FD5 --node http://localhost:26657

