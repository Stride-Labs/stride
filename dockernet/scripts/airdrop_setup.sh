### AIRDROP TESTING FLOW
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../config.sh

## INSTRUCTIONS
## - Start the network with `make start-docker`
## - Run this script with `bash dockernet/scripts/airdrop_setup.sh`

## NOTE: First, store the keys using the following mnemonics
## distributor address: stride1z835j3j65nqr6ng257q0xkkc9gta72gf48txwl
## distributor mnemonic: barrel salmon half click confirm crunch sense defy salute process cart fiscal sport clump weasel render private manage picture spell wreck hill frozen before
echo "barrel salmon half click confirm crunch sense defy salute process cart fiscal sport clump weasel render private manage picture spell wreck hill frozen before" | \
    $STRIDE_MAIN_CMD keys add distributor-test --recover

## airdrop-test address: stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr
## airdrop claimer mnemonic: royal auction state december october hip monster hotel south help bulk supreme history give deliver pigeon license gold carpet rabbit raw wool fatigue donate
echo "royal auction state december october hip monster hotel south help bulk supreme history give deliver pigeon license gold carpet rabbit raw wool fatigue donate" | \
    $STRIDE_MAIN_CMD keys add airdrop-test --recover

## AIRDROP SETUP
## Transfer uinj from injective to stride, so that we can liquid stake later
echo "Funding accounts..."
$INJECTIVE_MAIN_CMD tx ibc-transfer transfer transfer channel-0 stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr 1000000uinj --from ${INJECTIVE_VAL_PREFIX}1 -y --keyring-backend test
sleep 10
## Fund the distributor account
echo "Funding distributor account..."
$STRIDE_MAIN_CMD tx bank send val1 stride1z835j3j65nqr6ng257q0xkkc9gta72gf48txwl 600000ustrd --from val1 -y --keyring-backend test
sleep 10
## Fund the airdrop account
echo "Funding airdrop account..."
$STRIDE_MAIN_CMD tx bank send val1 stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr 1000000000ustrd --from val1 -y --keyring-backend test
sleep 10
## Create the airdrop, so that the airdrop account can claim tokens
echo "Creating airdrop..."
$STRIDE_MAIN_CMD tx claim create-airdrop stride1 1666792900 1000000 ustrd --from distributor-test -y --keyring-backend test
sleep 10
## Set airdrop allocations
echo "Setting airdrop allocations..."
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations stride1 stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr 1 --from distributor-test -y --keyring-backend test
sleep 10

## Check that the address is claimable
echo "Checking that the user has a claimable airdrop... expecting weight: '1.000000000000000000'..."
$STRIDE_MAIN_CMD  q claim claim-record stride1  stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr

