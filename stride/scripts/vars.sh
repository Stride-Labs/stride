#!/bin/bash

DEPENDENCIES="jq"

# check and install dependencies
echo -en "\nChecking dependencies... "
for name in $DEPENDENCIES
do
    [[ $(type $name 2>/dev/null) ]] || { echo -en "\n    * $name is required to run this script;";deps=1; }
done
[[ $deps -ne 1 ]] && echo -e "OK\n" || { echo -e "\nInstall the missing dependencies and rerun this script...\n"; exit 1; }

# define vars
STATE=state
STRIDE_CHAINS=(STRIDE_1 STRIDE_2 STRIDE_3)
main_chain=${STRIDE_CHAINS[0]}

ST1_RUN="docker-compose --ansi never run -T stride1 strided"
ST2_RUN="docker-compose --ansi never run -T stride2 strided"
ST3_RUN="docker-compose --ansi never run -T stride3 strided"

VAL_ACCTS=(val1 val2 val3)

V1="close soup mirror crew erode defy knock trigger gather eyebrow tent farm gym gloom base lemon sleep weekend rich forget diagram hurt prize fly"
V2="timber vacant teach wedding disease fashion place merge poet produce promote renew sunny industry enforce heavy inch three call sustain deal flee athlete intact"
V3="enjoy dignity rule multiply kitchen arrange flight rocket kingdom domain motion fire wage viable enough comic cry motor memory fancy dish sing border among"
VKEYS=("$V1" "$V2" "$V3")
BASE_RUN=strided
