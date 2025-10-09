#!/bin/bash
set -eu

STRIDE_HOME=${HOME}/.stride-localstride

prompt_continue() {
    operation=$1
    while true; do
        read -p "$operation? [y/n]" yn
        case $yn in
            [Yy]* ) echo ""; break;;
            [Nn]* ) exit ;;
            * ) printf "Please answer yes or no.\n";;
        esac
    done 
}

if [ -e $STRIDE_HOME ]; then
    prompt_continue "Clear ${STRIDE_HOME}"
    rm -rf ${STRIDE_HOME}
fi

if [ -e ${STRIDE_HOME}-backup ]; then
    prompt_continue "Clear ${STRIDE_HOME}-backup"
    rm -rf ${STRIDE_HOME}-backup
fi

prompt_continue "Initialize chain"

strided init local --chain-id stride-1 --overwrite --home ${STRIDE_HOME}
curl -L https://raw.githubusercontent.com/Stride-Labs/mainnet/main/mainnet/genesis.json -o ${STRIDE_HOME}/config/genesis.json
sed -i -E 's|seeds = ".*"|seeds = "ade4d8bc8cbe014af6ebdf3cb7b1e9ad36f412c0@seeds.polkachu.com:12256"|g' ${STRIDE_HOME}/config/config.toml

mnemonic="deer gaze swear marine one perfect hero twice turkey symbol mushroom hub escape accident prevent rifle horse arena secret endless panel equal rely payment" 
echo $mnemonic | strided keys add val --recover --keyring-backend test

snapshot_file=$(ls -t ${HOME}/Downloads/stride_*.tar.lz4 | head -n 1)
echo -e "\nSnapshot File: $snapshot_file"
prompt_continue "Continue"

lz4 -c -d $snapshot_file | tar -x -C ${STRIDE_HOME}
