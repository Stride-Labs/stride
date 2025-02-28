#!/bin/bash

set -eu

if [[ ! -f  peer_1_id.txt || ! -f  peer_2_id.txt ]]; then
    echo "Please start the node first before updating peers to make sure"
    echo "there is a peer_1_id.txt and peer_2_id.txt"
    exit 1
fi

STRIDE1_CONFIG=${HOME}/.stride1/config/config.toml
STRIDE2_CONFIG=${HOME}/.stride2/config/config.toml

STRIDE1_NODE_ID=$(cat ${HOME}/.stride1/node_id.txt)
STRIDE2_NODE_ID=$(cat ${HOME}/.stride2/node_id.txt)

STRIDE1_PORT=26656
STRIDE2_PORT=26658

STRIDE1_PEER=${STRIDE1_NODE_ID}@stride1:${STRIDE1_PORT}
STRIDE2_PEER=${STRIDE2_NODE_ID}@stride2:${STRIDE2_PORT}

sed -i "s|^persistent_peers = .*|persistent_peers = \"${STRIDE2_PEER}\"|" $STRIDE1_CONFIG
sed -i "s|^p2p.persistent_peers = .*|p2p.persistent_peers = \"${STRIDE1_PEER}\"|" $STRIDE2_CONFIG