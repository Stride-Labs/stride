#!/bin/sh
set -e
set -o pipefail

CONFIG_FOLDER=/home/stride/config

MNEMONIC="deer gaze swear marine one perfect hero twice turkey symbol mushroom hub escape accident prevent rifle horse arena secret endless panel equal rely payment"
CHAIN_ID="localstride"
MONIKER="val"


install_prerequisites () {
    sudo apk add -q --no-cache \
        python3 \
        py3-pip
}

edit_config () {

    # Remove seeds
    dasel put -t string -f $CONFIG_FOLDER/config.toml -s '.p2p.seeds' -v ''

    # Disable fast_sync
    dasel put -t bool -f $CONFIG_FOLDER/config.toml -s '.fast_sync' -v 'false'

    # Expose the rpc
    dasel put -t string -f $CONFIG_FOLDER/config.toml -s '.rpc.laddr' -v "tcp://0.0.0.0:26657"

    # Update the local client chain ID 
    dasel put -t string -f $CONFIG_FOLDER/client.toml -s '.chain-id' -v 'localstride'

    # Update the local client keyring backend
    dasel put -t string -f $CONFIG_FOLDER/client.toml -s '.keyring-backend' -v 'test'
}

if [[ ! -d $CONFIG_FOLDER ]]
then

    install_prerequisites

    echo "Chain ID: $CHAIN_ID"
    echo "Moniker:  $MONIKER"
    echo "MNEMONIC: $MNEMONIC"
    echo "STRIDE_HOME: $STRIDE_HOME"

    strided init localstride -o --chain-id=$CHAIN_ID --home $STRIDE_HOME

    echo $MNEMONIC | strided keys add val --recover --keyring-backend test --home $STRIDE_HOME

    ACCOUNT_PUBKEY=$(strided keys show --keyring-backend test val --pubkey --home $STRIDE_HOME | jq -r '.key')
    ACCOUNT_ADDRESS=$(strided keys show -a --keyring-backend test val --bech acc --home $STRIDE_HOME)

    VALIDATOR_PUBKEY_JSON=$(strided tendermint show-validator --home $STRIDE_HOME)
    VALIDATOR_PUBKEY=$(echo $VALIDATOR_PUBKEY_JSON | jq -r '.key')
    VALIDATOR_HEX_ADDRESS=$(strided debug pubkey $VALIDATOR_PUBKEY_JSON 2>&1 --home $STRIDE_HOME | grep Address | cut -d " " -f 2)
    VALIDATOR_ACCOUNT_ADDRESS=$(strided debug addr $VALIDATOR_HEX_ADDRESS 2>&1  --home $STRIDE_HOME | grep Acc | cut -d " " -f 3)
    VALIDATOR_OPERATOR_ADDRESS=$(strided debug addr $VALIDATOR_HEX_ADDRESS 2>&1  --home $STRIDE_HOME | grep Val | cut -d " " -f 3)
    VALIDATOR_CONSENSUS_ADDRESS=$(strided tendermint show-address --home $STRIDE_HOME)

    python3 -u testnetify.py \
    -i /home/stride/state_export.json \
    -o $CONFIG_FOLDER/genesis.json \
    -c $CHAIN_ID \
    --validator-hex-address $VALIDATOR_HEX_ADDRESS \
    --validator-operator-address $VALIDATOR_OPERATOR_ADDRESS \
    --validator-consensus-address $VALIDATOR_CONSENSUS_ADDRESS \
    --validator-pubkey $VALIDATOR_PUBKEY \
    --account-pubkey $ACCOUNT_PUBKEY \
    --account-address $ACCOUNT_ADDRESS

else 
    edit_config

    strided tendermint show-node-id --home $STRIDE_HOME > $STRIDE_HOME/node_id.txt

    strided start --home $STRIDE_HOME --x-crisis-skip-assert-invariants --reject-config-defaults
fi

