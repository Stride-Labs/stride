#!/bin/bash

set -eu 

UPGRADE_NAME="$1"
STRIDE_COMMIT_HASH="$2"
COSMOVISOR_HOME=/stride/.stride/cosmovisor/upgrades

echo "Upgrade Name: $UPGRADE_NAME"
printf "Stride Commit Hash: $STRIDE_COMMIT_HASH\n\n"

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
git checkout $STRIDE_COMMIT_HASH 
env GOOS=linux GOARCH=amd64 go build -mod=readonly -trimpath -o ${UPGRADE_DIR}/ ./... 
cd .. 
rm -rf stride go