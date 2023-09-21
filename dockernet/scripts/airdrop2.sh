### AIRDROP TESTING FLOW
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

# CLEANUP if running tests twice, clear out and re-fund accounts
$STRIDE_MAIN_CMD keys delete distributor-test -y &> /dev/null || true 
$GAIA_MAIN_CMD keys delete hosttest -y &> /dev/null || true 
$STRIDE_MAIN_CMD keys delete airdrop-test -y &> /dev/null || true 
$OSMO_MAIN_CMD keys delete host-address-test -y &> /dev/null || true 

# First, start the network with `make start-docker`
# Then, run this script with `bash dockernet/scripts/airdrop.sh`

# NOTE: First, store the keys using the following mnemonics
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
$STRIDE_MAIN_CMD tx claim create-airdrop stride 1666792900 40000000 ustrd --from distributor-test -y | TRIM_TX
sleep 5

## Test airdrop flow for chains who have non-standard coin types (not type 118). 
#       For example Evmos is using coin type 60, while Stride uses 118. Therefore, we can't map Evmos <> Stride addresses, because the one-way mapping works like this
#           seed phrase  ----> Evmos address (e.g. evmos123z469cfejeusvk87ufrs5520wmdxmmlc7qzuw)
#                        ----> Stride address (e.g. stride19uvw0azm9u0k6vqe4e22cga6kteskdqq3ulj6q)
#           and there is no function that can map between the two addresses.

#         evmos airdrop-test address: cosmos16lmf7t0jhaatan6vnxlgv47h2wf0k5lnhvye5h (rly2)
#            to test, we don't need to use evmos, just an address from a different mnemonic (can come from a coin_type 118 chain) 
#            here we choose to use an osmosis address with a new menmonic since we don't have an Evmos binary set up

echo ">>>Testing airdrop for coin types != 118..."
echo ">>>Testing for ibc-go version 3"
# Transfer uatom from gaia to stride, so that we can liquid stake later
$GAIA_MAIN_CMD tx bank send cosmos1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrgl2scj cosmos16lmf7t0jhaatan6vnxlgv47h2wf0k5lnhvye5h 1uatom --from ${GAIA_VAL_PREFIX}1 -y | TRIM_TX

#     setup: set an airdrop allocation for the mechanically converted stride address, converted using utils.ConvertAddressToStrideAddress()
#        mechanically-converted stride address: stride16lmf7t0jhaatan6vnxlgv47h2wf0k5ln58y9qm
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations stride stride16lmf7t0jhaatan6vnxlgv47h2wf0k5ln58y9qm 1 --from distributor-test -y | TRIM_TX
sleep 5

#     1. Overwrite incorrectly-derived stride address associated with an airdrop account with the proper Stride address (e.g. stride1abc...xyz)
#         a. query the claims module to verify that the airdrop-eligible address is as expected
$STRIDE_MAIN_CMD q claim claim-record stride stride16lmf7t0jhaatan6vnxlgv47h2wf0k5ln58y9qm

#         b. ibc-transfer from Osmo to Stride to change the airdrop account to stride1jrmtt5c6z8h5yrrwml488qnm7p3vxrrml2kgvl
#              Memo: {
#                "autopilot": {
#                     "stakeibc": {
#                       "stride_address": "stride1jrmtt5c6z8h5yrrwml488qnm7p3vxrrml2kgvl",
#                       },
#                         "claim": {
#                         }
#                    },
#                }
#               Receiver: "xxx"
memo='{"autopilot": {"receiver": "stride1jrmtt5c6z8h5yrrwml488qnm7p3vxrrml2kgvl","claim": { "stride_address": "stride1jrmtt5c6z8h5yrrwml488qnm7p3vxrrml2kgvl", "airdrop_id": "stride" } }}'
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 "$memo" 1uatom --from rly2 -y | TRIM_TX
echo ">>> Waiting for 15 seconds to allow the IBC transfer to complete..."
sleep 15
#         c. query the claims module 
#           - to verify nothing is eligible from the old address anymore stride16lmf7t0jhaatan6vnxlgv47h2wf0k5ln58y9qm
#           - to get the updated airdrop-eligible address's eligible amount from stride1jrmtt5c6z8h5yrrwml488qnm7p3vxrrml2kgvl
echo ">>> Querying the claims module to verify that the airdrop-eligible address is as expected"
echo "> previously eligible account, now should have 0:"
$STRIDE_MAIN_CMD q claim claim-record stride stride16lmf7t0jhaatan6vnxlgv47h2wf0k5ln58y9qm
echo "> new eligible account, now should have 1:"
$STRIDE_MAIN_CMD q claim claim-record stride stride1jrmtt5c6z8h5yrrwml488qnm7p3vxrrml2kgvl

        # liquid stake as a task to increase eligibility, re-check eligibliity 
$STRIDE_MAIN_CMD tx stakeibc liquid-stake 1 $ATOM_DENOM --from rly3 -y | TRIM_TX
sleep 5
echo "> after liquid staking eligiblity should be higher"
$STRIDE_MAIN_CMD q claim claim-record stride stride1jrmtt5c6z8h5yrrwml488qnm7p3vxrrml2kgvl

        # d. claim the airdrop from this address
echo "> Claiming the airdrop from the new address"
$STRIDE_MAIN_CMD tx claim claim-free-amount --from rly3 -y | TRIM_TX
sleep 5
        # e. verify funds are vesting
echo "> Verifying funds are vesting, should be 1."
$STRIDE_MAIN_CMD q claim user-vestings stride1jrmtt5c6z8h5yrrwml488qnm7p3vxrrml2kgvl