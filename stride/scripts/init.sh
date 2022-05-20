# this file should be called from the `stride` folder
# e.g. `sh ./scripts/init.sh`

# define vars
STATE=state
STRIDE_CHAINS=(STRIDE_1 STRIDE_2 STRIDE_3)
VAL_ACCT=val
V1="close soup mirror crew erode defy knock trigger gather eyebrow tent farm gym gloom base lemon sleep weekend rich forget diagram hurt prize fly"
V2="timber vacant teach wedding disease fashion place merge poet produce promote renew sunny industry enforce heavy inch three call sustain deal flee athlete intact"
V3="enjoy dignity rule multiply kitchen arrange flight rocket kingdom domain motion fire wage viable enough comic cry motor memory fancy dish sing border among"
VKEYS=("$V1" "$V2" "$V3")
BASE_RUN=strided

# cleanup any stale state
rm -rf $STATE
docker-compose down

# first, we need to create some saved state, so that we can copy to docker files
for chain_name in ${STRIDE_CHAINS[@]}; do
    mkdir -p ./$STATE/$chain_name
done

# then, we initialize our chains 
echo 'Initializing chains...'
for i in "${!STRIDE_CHAINS[@]}"; do
    chain_name=${STRIDE_CHAINS[i]}
    vkey=${VKEYS[i]}
    echo "\t$chain_name"
    $BASE_RUN init test --chain-id $chain_name --overwrite --home "$STATE/$chain_name" 2> /dev/null
    sed -i -E 's|"stake"|"ustrd"|g' "${STATE}/${chain_name}/config/genesis.json"
    # add VALidator account 
    echo $vkey | $BASE_RUN keys add $VAL_ACCT --recover --keyring-backend=test --home "$STATE/$chain_name" > /dev/null
    # get validator address
    VAL_ADDR=$($BASE_RUN keys show $VAL_ACCT --keyring-backend test -a --home "$STATE/$chain_name")
    # add money for this validator account
    $BASE_RUN add-genesis-account ${VAL_ADDR} 500000000000ustrd --home "$STATE/$chain_name"
    # actually set this account as a validator
    yes | $BASE_RUN gentx val 1000000000ustrd --chain-id $chain_name --keyring-backend test --home "$STATE/$chain_name"
    # this is just annoying, but we need to move the validator tx
    # cp "./${STATE}/${chain_name}/config/gentx/*.json" "./${STATE}/${chain_name}/config/gentx/"
    # now we process these txs 
    $BASE_RUN collect-gentxs --home "$STATE/$chain_name" 2> /dev/null
done

# strided start --home state/STRIDE_1  # TESTING ONLY
# next we build our docker images
docker build --no-cache --pull --tag stridezone:stride -f Dockerfile.stride .  # builds from scratch
# docker build --tag stridezone:stride -f Dockerfile.stride .  # uses cache to speed things up

# finally we serve our docker images
sleep 5
docker-compose up -d stride1 stride2 stride3


