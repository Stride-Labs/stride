#!/bin/sh
set -e
set -o pipefail

STRIDE_HOME=$HOME/.stride
CONFIG_FOLDER=$STRIDE_HOME/config

DEFAULT_MNEMONIC="deer gaze swear marine one perfect hero twice turkey symbol mushroom hub escape accident prevent rifle horse arena secret endless panel equal rely payment"
DEFAULT_CHAIN_ID="localstride"
DEFAULT_MONIKER="val"

# Override default values with environment variables
MNEMONIC=${MNEMONIC:-$DEFAULT_MNEMONIC}
CHAIN_ID=${CHAIN_ID:-$DEFAULT_CHAIN_ID}
MONIKER=${MONIKER:-$DEFAULT_MONIKER}

install_prerequisites () {
    sudo apk add -q --no-cache \
        python3 \
        py3-pip
}

edit_config () {

    # Remove seeds
    dasel put string -f $CONFIG_FOLDER/config.toml '.p2p.seeds' ''

    # Disable fast_sync
    dasel put bool -f $CONFIG_FOLDER/config.toml '.fast_sync' 'false'

    # Expose the rpc
    dasel put string -f $CONFIG_FOLDER/config.toml '.rpc.laddr' "tcp://0.0.0.0:26657"
}

if [[ ! -d $CONFIG_FOLDER ]]
then

    install_prerequisites

    echo "Chain ID: $CHAIN_ID"
    echo "Moniker:  $MONIKER"
    echo "MNEMONIC: $MNEMONIC"
    echo "STRIDE_HOME: $STRIDE_HOME"

    echo $MNEMONIC | strided init localstride -o --chain-id=$CHAIN_ID --home $STRIDE_HOME
    echo $MNEMONIC | strided keys add val --recover --keyring-backend test

    ACCOUNT_PUBKEY=$(strided keys show --keyring-backend test val --pubkey | dasel -r json '.key' --plain)
    ACCOUNT_ADDRESS=$(strided keys show -a --keyring-backend test val --bech acc)

    VALIDATOR_PUBKEY_JSON=$(strided tendermint show-validator --home $STRIDE_HOME)
    VALIDATOR_PUBKEY=$(echo $VALIDATOR_PUBKEY_JSON | dasel -r json '.key' --plain)
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

    edit_config
fi

strided start --home $STRIDE_HOME --x-crisis-skip-assert-invariants
