SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

echo '\n\nInitializing gaia...'

# import dependencies
source ${SCRIPT_DIR}/testnet_vars.sh $1

# cleanup any stale state
rm -rf $STATE/gaia*
mkdir $STATE/gaia

node_name=GAIA
$GAIA_CMD init test --chain-id $GAIA_CHAIN --overwrite 2> /dev/null

sed -i -E 's|"stake"|"uatom"|g' "${GAIA_FOLDER}/config/genesis.json"
sed -i -E 's|"full"|"validator"|g' "${GAIA_FOLDER}/config/config.toml"

echo "\n\n================== GAIA ==================" >> $STATE/keys.txt
$GAIA_CMD keys add gval --keyring-backend=test >> $STATE/keys.txt 2>&1

# get validator address
VAL_ADDR=$($GAIA_CMD keys show gval --keyring-backend test -a) > /dev/null

# add money for this validator account
$GAIA_CMD add-genesis-account ${VAL_ADDR} $GAIA_TOKENS
yes | $GAIA_CMD gentx gval $GAIA_STAKE_TOKENS --chain-id $GAIA_CHAIN --keyring-backend test 2> /dev/null

# now we grab the relevant node id
GAIA_NODE=$($GAIA_CMD tendermint show-node-id)@$GAIA_ENDPOINT:$PORT_ID
echo $GAIA_NODE
# Make Hermes account on gaia
$GAIA_CMD keys add rly2 --keyring-backend=test >> $STATE/keys.txt 2>&1
RLY_ADDRESS_2=$($GAIA_CMD keys show rly2 --keyring-backend test -a)

# Give relayer account token balance
$GAIA_CMD add-genesis-account ${RLY_ADDRESS_2} $GAIA_TOKENS

# Give icq account token balance
$GAIA_CMD add-genesis-account ${ICQ_ADDRESS_GAIA} $VAL_TOKENS
$GAIA_CMD collect-gentxs 2> /dev/null


# add small changes to config.toml
# use blind address (not loopback) to allow incoming connections from outside networks for local debugging
sed -i -E "s|127.0.0.1|0.0.0.0|g" "${GAIA_FOLDER}/config/config.toml"
sed -i -E "s|minimum-gas-prices = \"\"|minimum-gas-prices = \"0uatom\"|g" "${GAIA_FOLDER}/config/app.toml"
# allow CORS and API endpoints for block explorer
sed -i -E 's|enable = false|enable = true|g' "${GAIA_FOLDER}/config/app.toml"
sed -i -E 's|unsafe-cors = false|unsafe-cors = true|g' "${GAIA_FOLDER}/config/app.toml"


## add the message types ICA should allow to the host chain
ALLOW_MESSAGES='\"/cosmos.bank.v1beta1.MsgSend\", \"/cosmos.bank.v1beta1.MsgMultiSend\", \"/cosmos.staking.v1beta1.MsgDelegate\", \"/cosmos.staking.v1beta1.MsgUndelegate\", \"/cosmos.staking.v1beta1.MsgRedeemTokensforShares\", \"/cosmos.staking.v1beta1.MsgTokenizeShares\", \"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward\", \"/cosmos.distribution.v1beta1.MsgSetWithdrawAddress\", \"/ibc.applications.transfer.v1.MsgTransfer\"'
sed -i -E "s|\"allow_messages\": \[\]|\"allow_messages\": \[${ALLOW_MESSAGES}\]|g" "${GAIA_FOLDER}/config/genesis.json"
