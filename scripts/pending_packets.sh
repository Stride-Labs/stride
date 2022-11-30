#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/vars.sh

echo ">>> Transferring from GAIA"
bash scripts/test-util/1.sh
sleep 5

echo ">>> Liquid Stake"
bash scripts/test-util/2.sh | TRIM_TX
sleep 60
bash scripts/test-util/2.sh | TRIM_TX
sleep 60

echo ">>> Disable ICAs"
sed -i -E 's|rule: ""|rule: "allowlist"|g' scripts/state/relayer-gaia/config/config.yaml
sed -i -E 's|channel-list: \[\]|channel-list: \[channel-0\]|g' scripts/state/relayer-gaia/config/config.yaml
docker rm -f stride-relayer-gaia-1
sleep 2
printf "\n\nRESTARTING\n\n" >> scripts/logs/relayer-gaia.log
docker-compose up -d relayer-gaia

echo ">>> Liquid Stake Again"
bash scripts/test-util/2.sh | TRIM_TX
sleep 60
bash scripts/test-util/2.sh | TRIM_TX
sleep 60
bash scripts/test-util/2.sh | TRIM_TX
sleep 60

echo ">>> Re-enable ICAs"
sed -i -E 's|rule: "allowlist"|rule: ""|g' scripts/state/relayer-gaia/config/config.yaml
sed -i -E 's|channel-list: \[channel-0\]|channel-list: \[\]|g' scripts/state/relayer-gaia/config/config.yaml
docker rm -f stride-relayer-gaia-1
sleep 2
printf "\n\nRESTARTING\n\n" >> scripts/logs/relayer-gaia.log
docker-compose up -d relayer-gaia
docker-compose logs -f relayer-gaia | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" >> scripts/logs/relayer-gaia.log 2>&1 &
sleep 10

echo ">>> Restore channels"
build/strided --home scripts/state/stride1 tx stakeibc restore-interchain-account GAIA WITHDRAWAL --from val1 --gas 400000 -y | TRIM_TX
sleep 5
build/strided --home scripts/state/stride1 tx stakeibc restore-interchain-account GAIA DELEGATION --from val1 --gas 400000 -y | TRIM_TX