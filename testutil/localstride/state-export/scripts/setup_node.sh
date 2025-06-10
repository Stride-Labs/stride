#!/bin/bash

POLKACHU_SNAPSHOT_API="https://polkachu.com/api/v2/chain_snapshots/stride/"
STRIDE_HOME=${HOME}/.stride

if [[ -d $STRIDE_HOME ]]; then
  echo "$STRIDE_HOME directory exists. Please rename before proceeding"
  exit 1
fi

SNAPSHOT_URL=$(curl -s -H "x-polkachu: stride" $POLKACHU_SNAPSHOT_API | jq -r '.snapshot.url')

strided init localstride --chain-id stride-1 --overwrite
curl -L https://raw.githubusercontent.com/Stride-Labs/mainnet/main/mainnet/genesis.json -o ~/.stride/config/genesis.json
sed -i -E 's|seeds = ".*"|seeds = "ade4d8bc8cbe014af6ebdf3cb7b1e9ad36f412c0@seeds.polkachu.com:12256"|g' ~/.stride/config/config.toml

echo "Downloading pruned state from $SNAPSHOT_URL..."
curl -L -f $SNAPSHOT_URL -o ${STRIDE_HOME}/pruned_state.tar.lz4
lz4 -c -d ${STRIDE_HOME}/pruned_state.tar.lz4 | tar -x -C $STRIDE_HOME
rm -f ${STRIDE_HOME}/pruned_state.tar.lz4
