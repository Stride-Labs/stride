### AIRDROP TESTING FLOW
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/vars.sh

# First, start the network with `make start-docker`
# Then, run this script with `bash scripts/airdrop.sh`

# NOTE: First, store the keys using the following mnemonics
# distributor-test address: stride1z835j3j65nqr6ng257q0xkkc9gta72gf48txwl
echo "barrel salmon half click confirm crunch sense defy salute process cart fiscal sport clump weasel render private manage picture spell wreck hill frozen before" | \
    $STRIDE_MAIN_CMD keys add distributor-test --recover
# distributor-test-2 address: stride1nhll7ze2xzdxqhmc7mrc3ck7nz8nr367lg2689
echo "noise frozen tell snack wear hybrid wedding finish avocado prepare logic warrior ribbon law purity token theory first clown spread mother payment cycle dash" | \
    $STRIDE_MAIN_CMD keys add distributor-test-2 --recover
# distributor-test-3 address: stride1ff8krgfpjwj92gavkrayn5nvd9z58n3cllef3p
echo "shock main path weird hair note brisk elephant lyrics donkey company sea celery clog credit prize knife perfect inch banana ozone patient subject describe" | \
    $STRIDE_MAIN_CMD keys add distributor-test-3 --recover

# airdrop-test address: stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr
# airdrop claimer mnemonic: royal auction state december october hip monster hotel south help bulk supreme history give deliver pigeon license gold carpet rabbit raw wool fatigue donate
echo "royal auction state december october hip monster hotel south help bulk supreme history give deliver pigeon license gold carpet rabbit raw wool fatigue donate" | \
    $STRIDE_MAIN_CMD keys add airdrop-test --recover

## AIRDROP SETUP
echo "Funding accounts..."
# Transfer uatom from gaia to stride, so that we can liquid stake later
$GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr 1000000uatom --from ${GAIA_VAL_PREFIX}1 -y 
sleep 5
# Fund distributor-test
$STRIDE_MAIN_CMD tx bank send val1 stride1z835j3j65nqr6ng257q0xkkc9gta72gf48txwl 100000000ustrd --from val1 -y
sleep 5
# Fund distributor-test-2
$STRIDE_MAIN_CMD tx bank send val1 stride1nhll7ze2xzdxqhmc7mrc3ck7nz8nr367lg2689 200000000ustrd --from val1 -y
sleep 5
# Fund distributor-test-3
$STRIDE_MAIN_CMD tx bank send val1 stride1ff8krgfpjwj92gavkrayn5nvd9z58n3cllef3p 300000000ustrd --from val1 -y
sleep 5
# Fund the airdrop account
$STRIDE_MAIN_CMD tx bank send val1 stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr 1000000000ustrd --from val1 -y
sleep 5

# Create the STRIDE airdrop, so that the airdrop account can claim tokens
# create-airdrop [identifier] [start] [duration] [denom]
$STRIDE_MAIN_CMD tx claim create-airdrop stride 1666792900 40000000 ustrd --from distributor-test -y
sleep 5
# Set airdrop allocations
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations stride stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr 1 --from distributor-test -y
sleep 5
# Create the OSMOSIS airdrop
$STRIDE_MAIN_CMD tx claim create-airdrop osmosis 1666792900 50000000 ustrd --from distributor-test-2 -y
sleep 5
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations osmosis stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr 1 --from distributor-test-2 -y
sleep 5
# Create the JUNO airdrop
$STRIDE_MAIN_CMD tx claim create-airdrop juno 1666792900 600000000 ustrd --from distributor-test-3 -y
sleep 5
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations juno stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr 1 --from distributor-test-3 -y
sleep 5
OPTIONAL: Fund ledger
LEDGER=""
$STRIDE_MAIN_CMD tx bank send val1 $(LEDGER) 600000ustrd --from val1 -y
sleep 5
$STRIDE_MAIN_CMD tx claim set-airdrop-allocations stride $(LEDGER) 1 --from distributor-test -y
sleep 5

exit
# AIRDROP CLAIMS
# Check balances before claims
echo "Initial balance before claim:"
$STRIDE_MAIN_CMD query bank balances stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr
# NOTE: You can claim here using the CLI, or from the frontend!
# Claim 20% of the free tokens
echo "Claiming fee amount..."
$STRIDE_MAIN_CMD tx claim claim-free-amount --from airdrop-test --gas 500000
sleep 5
echo "Balance after claim:" 
$STRIDE_MAIN_CMD query bank balances stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr
# Stake, to claim another 20%
echo "Staking..."
$STRIDE_MAIN_CMD tx staking delegate stridevaloper1nnurja9zt97huqvsfuartetyjx63tc5zrj5x9f 100ustrd --from airdrop-test --gas 500000
sleep 5
echo "Balance after stake:" 
$STRIDE_MAIN_CMD query bank balances stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr
# Liquid stake, to claim the final 60% of tokens
echo "Liquid staking..."
$STRIDE_MAIN_CMD tx stakeibc liquid-stake 1000 uatom --from airdrop-test --gas 500000
sleep 5
echo "Balance after liquid stake:" 
$STRIDE_MAIN_CMD query bank balances stride1nf6v2paty9m22l3ecm7dpakq2c92ueyununayr