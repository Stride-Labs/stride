#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# First argument is build flags
BUILDDIR="$2"

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

