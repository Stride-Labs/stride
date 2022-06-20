#!/bin/bash

sleep 180

echo "Restoring ICQ Accounts"
interchain-queries keys restore test "ICQ_STRIDE_KEY" --chain stride-testnet
interchain-queries keys restore test "ICQ_GAIA_KEY" --chain gaia-testnet

$ICQ_RUN keys restore test "$ICQ_STRIDE_KEY" --chain stride-testnet
$ICQ_RUN keys restore test "$ICQ_GAIA_KEY" --chain gaia-testnet

echo "\nLaunch interchainquery relayer"
docker-compose up --force-recreate -d icq

interchain queries run