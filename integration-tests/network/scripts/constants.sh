#!/bin/bash

SCRIPTS_DIR=scripts
CONFIG_DIR=configs

VALIDATOR_KEYS_DIR=validator-keys
NODE_KEYS_DIR=node-keys
NODE_IDS_DIR=node-ids
GENESIS_DIR=genesis
KEYS_FILE=${CONFIG_DIR}/keys.json

PEER_PORT=26656
RPC_PORT=26657

API_ENDPOINT=${API_ENDPOINT:-'http://api.integration.svc:8000'}
