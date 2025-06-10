#!/bin/bash

set -eu

PEER_PORT=26656

STRIDE1_CONFIG=/home/stride/.stride1/config/config.toml
STRIDE2_CONFIG=/home/stride/.stride2/config/config.toml

STRIDE1_NODE_ID_FILE=/home/stride/.stride1/node_id.txt
STRIDE2_NODE_ID_FILE=/home/stride/.stride2/node_id.txt

if [[ ! -f $STRIDE1_NODE_ID_FILE || ! -f $STRIDE2_NODE_ID_FILE ]]; then 
    echo "Node ID files do not exist, please start nodes first"
    exit 1
fi

STRIDE1_NODE_ID=$(cat $STRIDE1_NODE_ID_FILE)
STRIDE2_NODE_ID=$(cat $STRIDE2_NODE_ID_FILE)

echo "Node ID #1: $STRIDE1_NODE_ID"
echo "Node ID #2: $STRIDE2_NODE_ID"

STRIDE1_PEER=${STRIDE1_NODE_ID}@stride1:${PEER_PORT}
STRIDE2_PEER=${STRIDE2_NODE_ID}@stride2:${PEER_PORT}

sudo sed -i "s|^  persistent_peers = .*|  persistent_peers = \"${STRIDE2_PEER}\"|" $STRIDE1_CONFIG
sudo sed -i "s|^  persistent_peers = .*|  persistent_peers = \"${STRIDE1_PEER}\"|" $STRIDE2_CONFIG
