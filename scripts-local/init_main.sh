#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh

# First argument is build flags
BUILDDIR="$2"
CACHE="${3:-false}"

# check and install dependencies
if [ "$CACHE" != "true" ]; then
    echo "\nChecking dependencies... "
    DEPENDENCIES="jq"
    deps=0
    for name in $DEPENDENCIES
    do
        [[ $(type $name 2>/dev/null) ]] || { echo "\n    * $name is required to run this script;";deps=1; }
    done
    [[ $deps -ne 1 ]] && echo "OK\n" || { echo "\nInstall the missing dependencies and rerun this script...\n"; exit 1; }

    # Clear existing state
    rm -rf $STATE ~/.hermes/keys ~/.icq/keys
else
    echo "Rebuilding while restoring cached state..."
fi

while getopts sghi flag; do
    case "${flag}" in
        s) printf '%s' "Building Stride... ";
           go build -mod=readonly -trimpath -o $BUILDDIR ./...;
           echo "Done" ;;
        g) printf '%s' "Building Gaia...   ";
           cd deps/gaia; 
           go build -mod=readonly -trimpath -o $BUILDDIR ./... 2>&1 | grep -v -E "deprecated|keychain" || true; 
           cd ../..;
           echo "Done" ;;
        h) printf '%s' "Building Hermes... ";
           cd deps/hermes; 
           cargo build --release --target-dir $BUILDDIR/hermes; 
           cd ../..
           echo "Done" ;;
        i) printf '%s' "Building ICQ...    ";
           cd deps/interchain-queries; 
           go build -mod=readonly -trimpath -o $BUILDDIR ./... 2>&1 | grep -v -E "deprecated|keychain" || true; 
           cd ../..
           echo "Done" ;;
    esac
done

# Initialize the state for stride/gaia and relayers
if [ "$CACHE" != "true" ]; then
    sh ${SCRIPT_DIR}/init_stride.sh
    sh ${SCRIPT_DIR}/init_gaia.sh
    sh ${SCRIPT_DIR}/init_relayers.sh
fi
