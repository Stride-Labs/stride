#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# First argument is build flags
BUILDDIR="$2"

while getopts sghijo flag; do
    case "${flag}" in
        s) printf '%s' "Building Stride... ";
           go build -mod=readonly -trimpath -tags "ledger" -ldflags '-X "github.com/cosmos/cosmos-sdk/version.BuildTags=ledger" -w -s' -o $BUILDDIR ./...; # having a build with ledger is convenient
           mkdir -p $BUILDDIR/stride2
           go build -mod=readonly -trimpath -o $BUILDDIR/stride2 ./... 2>&1 | grep -v -E "deprecated|keychain" || true; 
           mkdir -p $BUILDDIR/stride3
           go build -mod=readonly -trimpath -o $BUILDDIR/stride3 ./... 2>&1 | grep -v -E "deprecated|keychain" || true;
           mkdir -p $BUILDDIR/stride4
           go build -mod=readonly -trimpath -o $BUILDDIR/stride4 ./... 2>&1 | grep -v -E "deprecated|keychain" || true;
           mkdir -p $BUILDDIR/stride5
           go build -mod=readonly -trimpath -o $BUILDDIR/stride5 ./... 2>&1 | grep -v -E "deprecated|keychain" || true; 
           echo "Done" ;;
        g) printf '%s' "Building Gaia...   ";
           cd deps/gaia; 
           go build -mod=readonly -trimpath -o $BUILDDIR ./... 2>&1 | grep -v -E "deprecated|keychain" || true; 
           mkdir -p $BUILDDIR/gaia2
           go build -mod=readonly -trimpath -o $BUILDDIR/gaia2 ./... 2>&1 | grep -v -E "deprecated|keychain" || true; 
           mkdir -p $BUILDDIR/gaia3
           go build -mod=readonly -trimpath -o $BUILDDIR/gaia3 ./... 2>&1 | grep -v -E "deprecated|keychain" || true; 
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
        j) printf '%s' "Building Juno...   ";
           cd deps/juno; 
           go build -mod=readonly -trimpath -o $BUILDDIR ./... 2>&1 | grep -v -E "deprecated|keychain" || true; 
           cd ../..
           echo "Done" ;;
        o) printf '%s' "Building Osmosis...   ";
           cd deps/osmosis; 
           go build -mod=readonly -trimpath -o $BUILDDIR ./... 2>&1 | grep -v -E "deprecated|keychain" || true; 
           cd ../..
           echo "Done" ;;
    esac
done