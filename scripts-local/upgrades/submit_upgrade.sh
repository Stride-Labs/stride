#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

STRIDED="$SCRIPT_DIR/binaries/strided1 --home $SCRIPT_DIR/../state/stride"

printf "\nPROPOSAL\n"
$STRIDED tx gov submit-proposal software-upgrade $PROPOSAL_NAME \
    --title $PROPOSAL_NAME --description "version 2 description" \
    --upgrade-height $UPGRADE_HEIGHT --from val1 -y

sleep 5
printf "\nPROPOSAL CONFIRMATION\n"
$STRIDED query gov proposals

sleep 5 
printf "\nDEPOSIT\n"
$STRIDED tx gov deposit 1 10000001ustrd --from val1 -y

sleep 5
printf "\nDEPOSIT CONFIRMATION\n"
$STRIDED query gov deposits 1

sleep 5
printf "\nVOTE\n"
$STRIDED tx gov vote 1 yes --from val1 -y

sleep 5
printf "\nVOTE CONFIRMATION\n"
$STRIDED query gov tally 1
