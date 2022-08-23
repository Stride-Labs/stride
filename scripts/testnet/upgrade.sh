#!/bin/bash

set -eu 

UPGRADE_NAME="$1"
UPGRADE_COMMIT_HASH="$2"

COSMOVISOR_HOME=/stride/.stride/cosmovisor/upgrades

echo "Upgrade Name: $UPGRADE_NAME"
printf "Upgrade Commit Hash: $UPGRADE_COMMIT_HASH\n\n"

while true; do
    read -p "Continue? [y/n]" yn
    case $yn in
        [Yy]* ) echo ""; break;;
        [Nn]* ) exit ;;
        * ) printf "Please answer yes or no.\n";;
    esac
done

UPGRADE_DIR=${COSMOVISOR_HOME}/${UPGRADE_NAME}/bin

mkdir -p $UPGRADE_DIR
git clone https://github.com/Stride-Labs/stride.git
cd stride
git checkout $UPGRADE_COMMIT_HASH 
env GOOS=linux GOARCH=amd64 go build -mod=readonly -trimpath -o ${UPGRADE_DIR}/ ./... 
cp ${UPGRADE_DIR}/strided /usr/local/bin/strided 
cd .. 
rm -rf stride go