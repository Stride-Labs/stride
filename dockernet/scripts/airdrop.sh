### AIRDROP TESTING FLOW
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../config.sh

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

# airdrop-test address: stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr
# airdrop claimer mnemonic: royal auction state december october hip monster hotel south help bulk supreme history give deliver pigeon license gold carpet rabbit raw wool fatigue donate
echo "royal auction state december october hip monster hotel south help bulk supreme history give deliver pigeon license gold carpet rabbit raw wool fatigue donate" | \
    $STRIDE_MAIN_CMD keys add airdrop-test --recover

## AIRDROP SETUP
echo "Funding accounts..."
# Transfer uatom from gaia to stride, so that we can liquid stake later
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr 1000000uatom --from ${GAIA_VAL_PREFIX}1 -y 
sleep 5
# Fund the distributor account
$STRIDE_MAIN_CMD tx bank send val1 stride1z835j3j65nqr6ng257q0xkkc9gta72gf48txwl 600000ustrd --from val1 -y
sleep 5
# Fund the airdrop account
$STRIDE_MAIN_CMD tx bank send val1 stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr 1000000000ustrd --from val1 -y
sleep 5
# Create the airdrop, so that the airdrop account can claim tokens
$STRIDE_MAIN_CMD tx claim create-airdrop stride 1666792900 40000000 ustrd --from distributor-test -y
sleep 5
# Set airdrop allocations
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations stride stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr 1 --from distributor-test -y
sleep 5

# AIRDROP CLAIMS
# Check balances before claims
echo "Initial balance before claim:"
$STRIDE_MAIN_CMD query bank balances stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr
# NOTE: You can claim here using the CLI, or from the frontend!
# Claim 20% of the free tokens
echo "Claiming fee amount..."
$STRIDE_MAIN_CMD tx claim claim-free-amount --from airdrop-test --gas 400000 -y
sleep 5
echo "Balance after claim:" 
$STRIDE_MAIN_CMD query bank balances stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr
# Stake, to claim another 20%
echo "Staking..."
$STRIDE_MAIN_CMD tx staking delegate stridevaloper1nnurja9zt97huqvsfuartetyjx63tc5zrj5x9f 100ustrd --from airdrop-test --gas 400000 -y
sleep 5
echo "Balance after stake:" 
$STRIDE_MAIN_CMD query bank balances stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr
# Liquid stake, to claim the final 60% of tokens
echo "Liquid staking..."
$STRIDE_MAIN_CMD tx stakeibc liquid-stake 1000 uatom --from airdrop-test --gas 400000 -y
sleep 5
echo "Balance after liquid stake:" 
$STRIDE_MAIN_CMD query bank balances stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr



## Test airdrop flow for chains who have non-standard coin types (not type 118). 
#       For example Evmos is using coin type 60, while Stride uses 118. Therefore, we can't map Evmos <> Stride addresses, because the one-way mapping works like this
#           seed phrase  ----> Evmos address (e.g. evmos123z469cfejeusvk87ufrs5520wmdxmmlc7qzuw)
#                        ----> Stride address (e.g. stride19uvw0azm9u0k6vqe4e22cga6kteskdqq3ulj6q)
#           and there is no function that can map between the two addresses.

#         evmos airdrop-test address: cosmos16lmf7t0jhaatan6vnxlgv47h2wf0k5lnhvye5h (rly2)
#            to test, we don't need to use evmos, just an address from a different mnemonic (can come from a coin_type 118 chain) 
#            here we choose to use an osmosis address with a new menmonic since we don't have an Evmos binary set up

echo "Testing airdrop for coin types != 118..."

# Transfer uatom from gaia to stride, so that we can liquid stake later
$GAIA_MAIN_CMD tx bank send cosmos1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrgl2scj cosmos16lmf7t0jhaatan6vnxlgv47h2wf0k5lnhvye5h 1uatom --from ${GAIA_VAL_PREFIX}1 -y 

#     setup: set an airdrop allocation for the mechanically converted stride address, converted using utils.ConvertAddressToStrideAddress()
#        mechanically-converted stride address: stride18y9zdh00fr2t6uw20anr6e89svqmfddgxfsxkh
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations stride stride16lmf7t0jhaatan6vnxlgv47h2wf0k5ln58y9qm 1 --from distributor-test -y
sleep 5

#     1. Overwrite incorrectly-derived stride address associated with an airdrop account with the proper Stride address (e.g. stride1abc...xyz)
#         a. query the claims module to verify that the airdrop-eligible address is as expected
$STRIDE_MAIN_CMD q claim claim-record stride stride16lmf7t0jhaatan6vnxlgv47h2wf0k5ln58y9qm

#         b. ibc-transfer from Osmo to Stride to change the airdrop account to stride1qz677nj82mszxjuh4mzy52zv5md5qrgg60pxpc
#              Memo: {
#                "autopilot": {
#                     "stakeibc": {
#                       "stride_address": "stride1qz677nj82mszxjuh4mzy52zv5md5qrgg60pxpc",
#                       },
#                         "claim": {
#                         }
#                    },
#                }
#               Receiver: "xxx"
memo='{"autopilot": {"receiver": "stride1qz677nj82mszxjuh4mzy52zv5md5qrgg60pxpc","claim": { "stride_address": "stride1qz677nj82mszxjuh4mzy52zv5md5qrgg60pxpc", "airdrop_id": "stride" } }}'
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 "$memo" 1uatom --from rly2 -y 

#         c. query the claims module 
#           - to verify nothing is eligible from the old address anymore stride18y9zdh00fr2t6uw20anr6e89svqmfddgxfsxkh
#           - to get the updated airdrop-eligible address's eligible amount from stride1qz677nj82mszxjuh4mzy52zv5md5qrgg60pxpc
$STRIDE_MAIN_CMD q claim claim-record stride stride16lmf7t0jhaatan6vnxlgv47h2wf0k5ln58y9qm
$STRIDE_MAIN_CMD q claim claim-record stride stride1qz677nj82mszxjuh4mzy52zv5md5qrgg60pxpc

        # d. claim the airdrop from this address
# $STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1qz677nj82mszxjuh4mzy52zv5md5qrgg60pxpc

        # e. verify the vesting account is created for stride1qz677nj82mszxjuh4mzy52zv5md5qrgg60pxpc
# $STRIDE_MAIN_CMD q auth account stride1qz677nj82mszxjuh4mzy52zv5md5qrgg60pxpc



### Test airdrop reset and multiple claims flow
    #   The Stride airdrop occurs in batches. We need to test three batches. 

    # SETUP
    # 1. Create a new airdrop that rolls into its next batch in just 30 seconds
    #    - include the add'l param that makes each batch 30 seconds long (after the first batch) 
    # 2. Set the airdrop allocations

# Create the airdrop, so that the airdrop account can claim tokens
# $STRIDE_MAIN_CMD tx claim create-airdrop stride2 $(date +%s) 30 ustrd --from distributor-test -y
# sleep 5
# Set airdrop allocations
# $STRIDE_MAIN_CMD tx claim set-airdrop-allocations stride2 stride1kd3z076usuqytj9rdfqnqaj9sdyx9aq5j2lqs5 1 --from distributor-test -y
# sleep 5

    # BATCH 1
    # 3. Claim the airdrop
# $STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kd3z076usuqytj9rdfqnqaj9sdyx9aq5j2lqs5

    # 4. Check that the claim worked
# $STRIDE_MAIN_CMD q bank balances stride1kd3z076usuqytj9rdfqnqaj9sdyx9aq5j2lqs5

    # 5. Query to check airdrop vesting account was created (w/ correct amount)
# $STRIDE_MAIN_CMD q auth account stride1kd3z076usuqytj9rdfqnqaj9sdyx9aq5j2lqs5


    # BATCH 2
    # 6. Wait 30 seconds
# sleep 30
    # 7. Claim the airdrop
# $STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kd3z076usuqytj9rdfqnqaj9sdyx9aq5j2lqs5

    # 8. Check that the claim worked
# $STRIDE_MAIN_CMD q bank balances stride1kd3z076usuqytj9rdfqnqaj9sdyx9aq5j2lqs5

    # 9. Query to check airdrop vesting account was created (w/ correct amount)
# $STRIDE_MAIN_CMD q auth account stride1kd3z076usuqytj9rdfqnqaj9sdyx9aq5j2lqs5

    # BATCH 3
    # 10. Wait 30 seconds
# sleep 30
    # 11. Claim the airdrop
# $STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kd3z076usuqytj9rdfqnqaj9sdyx9aq5j2lqs5

    # 12. Check that the claim worked
# $STRIDE_MAIN_CMD q bank balances stride1kd3z076usuqytj9rdfqnqaj9sdyx9aq5j2lqs5

    # 13. Query to check airdrop vesting account was created (w/ correct amount)
# $STRIDE_MAIN_CMD q auth account stride1kd3z076usuqytj9rdfqnqaj9sdyx9aq5j2lqs5



### Test staggered airdrops

# create airdrop 1 with a 60 day start window, 60 sec reset, claim, sleep 35
# $STRIDE_MAIN_CMD tx claim create-airdrop airdrop1 $(date +%s) 60 ustrd --from distributor-test -y
# sleep 5
# $STRIDE_MAIN_CMD tx claim set-airdrop-allocations airdrop1 stride1kd3z076usuqytj9rdfqnqaj9sdyx9aq5j2lqs5 1 --from distributor-test -y
# sleep 5
# $STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kd3z076usuqytj9rdfqnqaj9sdyx9aq5j2lqs5
# sleep 35

# # create airdrop 2 with a 60 day start window, 60 sec reset, claim, sleep 35
# $STRIDE_MAIN_CMD tx claim create-airdrop airdrop1 $(date +%s) 60 stuatom --from distributor-test -y
# sleep 5
# $STRIDE_MAIN_CMD tx claim set-airdrop-allocations airdrop1 stride1kd3z076usuqytj9rdfqnqaj9sdyx9aq5j2lqs5 1 --from distributor-test -y
# sleep 5
# $STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kd3z076usuqytj9rdfqnqaj9sdyx9aq5j2lqs5
# sleep 35

# # airdrop 1 resets
# $STRIDE_MAIN_CMD q bank balances stride1kd3z076usuqytj9rdfqnqaj9sdyx9aq5j2lqs5
# $STRIDE_MAIN_CMD tx claim claim-free-amount --from stride1kd3z076usuqytj9rdfqnqaj9sdyx9aq5j2lqs5
# $STRIDE_MAIN_CMD q bank balances stride1kd3z076usuqytj9rdfqnqaj9sdyx9aq5j2lqs5

