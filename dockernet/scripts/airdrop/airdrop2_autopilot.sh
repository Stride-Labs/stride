### AIRDROP TESTING FLOW
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

### AIRDROP TESTING FLOW Pt 2 (AUTOPILOT)

# This script tests airdrop claiming via autopilot
# The claim is initiated by sending an IBC transfer with the stride address in the memo
# Gaia is used for this test with ibc v3 - and the memo is included in the receiver field of the transfer

# To run:
#   1. Start the network with `make start-docker`
#   2. Run this script with `bash dockernet/scripts/airdrop/airdrop2_autopilot.sh`

# NOTE: First, store the keys using the following mnemonics
echo "Registering accounts..."
# distributor address: stride1z835j3j65nqr6ng257q0xkkc9gta72gf48txwl
# distributor mnemonic: barrel salmon half click confirm crunch sense defy salute process cart fiscal sport clump weasel render private manage picture spell wreck hill frozen before
echo "barrel salmon half click confirm crunch sense defy salute process cart fiscal sport clump weasel render private manage picture spell wreck hill frozen before" | \
    $STRIDE_MAIN_CMD keys add distributor-test --recover

## AIRDROP SETUP
echo "Funding accounts..."
# Transfer uatom from gaia to stride, so that we can liquid stake later
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr 1000000uatom --from ${GAIA_VAL_PREFIX}1 -y | TRIM_TX
sleep 5
# Fund the distributor account
$STRIDE_MAIN_CMD tx bank send val1 stride1z835j3j65nqr6ng257q0xkkc9gta72gf48txwl 600000ustrd --from val1 -y | TRIM_TX
sleep 5
# Create the airdrop, so that the airdrop account can claim tokens
$STRIDE_MAIN_CMD tx claim create-airdrop gaia GAIA ustrd $(date +%s) 40000000 true --from distributor-test -y | TRIM_TX
sleep 5

## Test airdrop flow for chains who have non-standard coin types (not type 118). 
#       For example Evmos is using coin type 60, while Stride uses 118. Therefore, we can't map Evmos <> Stride addresses, because the one-way mapping works like this
#           seed phrase  ----> Evmos address (e.g. evmos123z469cfejeusvk87ufrs5520wmdxmmlc7qzuw)
#                        ----> Stride address (e.g. stride19uvw0azm9u0k6vqe4e22cga6kteskdqq3ulj6q)
#           and there is no function that can map between the two addresses.

#         evmos airdrop-test address: cosmos16lmf7t0jhaatan6vnxlgv47h2wf0k5lnhvye5h (rly2)
#            to test, we don't need to use evmos, just an address from a different mnemonic (can come from a coin_type 118 chain) 
#            here we choose to use an osmosis address with a new menmonic since we don't have an Evmos binary set up

echo -e "\n>>> Testing airdrop for coin types != 118, ibc-go version 3..."
# Transfer uatom from gaia to stride, so that we can liquid stake later
$GAIA_MAIN_CMD tx bank send cosmos1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrgl2scj cosmos16lmf7t0jhaatan6vnxlgv47h2wf0k5lnhvye5h 1uatom --from ${GAIA_VAL_PREFIX}1 -y | TRIM_TX

#     setup: set an airdrop allocation for the mechanically converted stride address, converted using utils.ConvertAddressToStrideAddress()
#        mechanically-converted stride address: stride16lmf7t0jhaatan6vnxlgv47h2wf0k5ln58y9qm
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations gaia stride16lmf7t0jhaatan6vnxlgv47h2wf0k5ln58y9qm 1 --from distributor-test -y | TRIM_TX
sleep 5

#     1. Overwrite incorrectly-derived stride address associated with an airdrop account with the proper Stride address (e.g. stride1abc...xyz)
#         a. query the claims module to verify that the airdrop-eligible address is as expected
echo "> initial claim record [should show one record]:"
$STRIDE_MAIN_CMD q claim claim-record gaia stride16lmf7t0jhaatan6vnxlgv47h2wf0k5ln58y9qm

#         b. ibc-transfer from Osmo to Stride to change the airdrop account to stride1jrmtt5c6z8h5yrrwml488qnm7p3vxrrml2kgvl
#              Memo: {
#                "autopilot": {
#                     "claim": {
#                         "stride_address": "stride1jrmtt5c6z8h5yrrwml488qnm7p3vxrrml2kgvl",
#                      },
#                 },
#              }
#              Receiver: "xxx"
echo -e ">>> Claiming airdrop via IBC transfer..."
memo='{"autopilot": {"receiver": "stride1jrmtt5c6z8h5yrrwml488qnm7p3vxrrml2kgvl","claim": { "stride_address": "stride1jrmtt5c6z8h5yrrwml488qnm7p3vxrrml2kgvl" } }}'
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 "$memo" 1uatom --from rly2 -y | TRIM_TX
sleep 15
#         c. query the claims module 
#           - to verify nothing is eligible from the old address anymore stride16lmf7t0jhaatan6vnxlgv47h2wf0k5ln58y9qm
#           - to get the updated airdrop-eligible address's eligible amount from stride1jrmtt5c6z8h5yrrwml488qnm7p3vxrrml2kgvl
echo -e "\n>>> Querying the claims module to verify that the airdrop-eligible address is as expected"
echo "> Previously eligible account, should no longer return any records:"
$STRIDE_MAIN_CMD q claim claim-record gaia stride16lmf7t0jhaatan6vnxlgv47h2wf0k5ln58y9qm
echo "> New eligible account, should show 1 record:"
$STRIDE_MAIN_CMD q claim claim-record gaia stride1jrmtt5c6z8h5yrrwml488qnm7p3vxrrml2kgvl

        # liquid stake as a task to increase eligibility, re-check eligibliity 
echo -e "\n>>> Liquid staking..."
$STRIDE_MAIN_CMD tx stakeibc liquid-stake 1 $ATOM_DENOM --from rly3 -y | TRIM_TX
sleep 5
echo "> After liquid staking there should be one action complete"
$STRIDE_MAIN_CMD q claim claim-record gaia stride1jrmtt5c6z8h5yrrwml488qnm7p3vxrrml2kgvl | grep claim_record -A 4

        # d. claim the airdrop from this address
echo -e "\n>>> Claiming the airdrop from the new address"
$STRIDE_MAIN_CMD tx claim claim-free-amount --from rly3 -y | TRIM_TX
sleep 5

echo "> After claiming, there should be two action complete"
$STRIDE_MAIN_CMD q claim claim-record gaia stride1jrmtt5c6z8h5yrrwml488qnm7p3vxrrml2kgvl | grep claim_record -A 4

        # e. verify funds are vesting
echo "> Verifying vesting record [expected: 120000ustrd]:"
$STRIDE_MAIN_CMD q claim user-vestings stride1jrmtt5c6z8h5yrrwml488qnm7p3vxrrml2kgvl | grep spendable_coins -A 2
