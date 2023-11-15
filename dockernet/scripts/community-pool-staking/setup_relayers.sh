#!/bin/bash
set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

for path in "dydx-noble" "noble-osmo" "osmo-dydx"; do
    relayer_logs=${LOGS}/relayer-${path}.log
    relayer_config=$STATE/relayer-${path}/config
    relayer_exec="$DOCKER_COMPOSE run --rm relayer-$path"

    mkdir -p $relayer_config
    chmod -R 777 $STATE/relayer-${path}
    cp ${DOCKERNET_HOME}/config/relayer_config_dydx_noble.yaml $relayer_config/config.yaml

    IFS='-' read -r zone_1 zone_2 <<< "$path"

    ZONE_1=$(printf "$zone_1" | awk '{ print toupper($0) }')
    ZONE_2=$(printf "$zone_2" | awk '{ print toupper($0) }')

    mnemonic_1=$(GET_VAR_VALUE RELAYER_${ZONE_1}_${ZONE_2}_MNEMONIC)
    mnemonic_2=$(GET_VAR_VALUE RELAYER_${ZONE_2}_${ZONE_1}_MNEMONIC)

    cmd_1=$(GET_VAR_VALUE      ${ZONE_1}_MAIN_CMD)
    val_acct_1=$(GET_VAR_VALUE ${ZONE_1}_VAL_PREFIX)1
    denom_1=$(GET_VAR_VALUE    ${ZONE_1}_DENOM)

    cmd_2=$(GET_VAR_VALUE      ${ZONE_2}_MAIN_CMD)
    val_acct_2=$(GET_VAR_VALUE ${ZONE_2}_VAL_PREFIX)1
    denom_2=$(GET_VAR_VALUE    ${ZONE_2}_DENOM)

    echo "${ZONE_1} <> ${ZONE_2} - Adding relayer keys..."
    relayer_address_1=$($relayer_exec rly keys restore $zone_1 $zone_1 "$mnemonic_1")
    relayer_address_2=$($relayer_exec rly keys restore $zone_2 $zone_2 "$mnemonic_2")
    echo $relayer_address_1
    echo $relayer_address_2
    echo "Done"

    # Ignore noble when funding since the relayers are funded at genesis
    echo "${ZONE_1} <> ${ZONE_2} - Funding relayers..."
    if [ "$ZONE_1" != "NOBLE" ]; then
        $cmd_1 tx bank send $($cmd_1 keys show -a $val_acct_1) $relayer_address_1 10000000${denom_1} --from ${val_acct_1} -y | TRIM_TX
    fi
    if [ "$ZONE_2" != "NOBLE" ]; then
        $cmd_2 tx bank send $($cmd_2 keys show -a $val_acct_2) $relayer_address_2 10000000${denom_2} --from ${val_acct_2} -y | TRIM_TX
    fi
    sleep 5

    printf "${ZONE_1} <> ${ZONE_2} - Creating client, connection, and transfer channel..." | tee -a $relayer_logs
    $relayer_exec rly transact link ${path} >> $relayer_logs 2>&1
    echo "Done"

    $DOCKER_COMPOSE up -d relayer-${path}
    SAVE_DOCKER_LOGS relayer-${path} $relayer_logs
done
