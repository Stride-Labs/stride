#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/vars.sh

BUILDDIR="$2"

# build docker images and local binaries
<<<<<<< HEAD
while getopts sghir flag; do
    case "${flag}" in
        s) echo "Building Stride Docker...  ";
=======
while getopts sghi flag; do
    case "${flag}" in
        s) printf '%s' "Building Stride Docker...  ";
>>>>>>> 03e31a61 (setup relayers)
           docker build --tag stridezone:stride -f Dockerfile.stride . ;

           printf '%s' "Building Stride Locally...  ";
           go build -mod=readonly -trimpath -o $BUILDDIR ./... ;
           echo "Done" ;;

<<<<<<< HEAD
        g) echo "Building Gaia Docker...    ";
=======
        g) printf '%s' "Building Gaia Docker...    ";
>>>>>>> 03e31a61 (setup relayers)
           docker build --tag stridezone:gaia -f Dockerfile.gaia . ;

           printf '%s' "Building Gaia Locally...   ";
           cd deps/gaia ; 
           go build -mod=readonly -trimpath -o $BUILDDIR ./... 2>&1 | grep -v -E "deprecated|keychain" || true ;
           cd ../.. ;
           echo "Done" ;;

<<<<<<< HEAD
        h) echo "Building Hermes Docker... ";
=======
        h) printf '%s' "Building Hermes Docker... ";
>>>>>>> 03e31a61 (setup relayers)
           docker build --tag stridezone:hermes -f Dockerfile.hermes . ;

           printf '%s' "Building Hermes Locally... ";
           cd deps/hermes; 
           cargo build --release --target-dir $BUILDDIR/hermes; 
           cd ../..
           echo "Done" ;;

<<<<<<< HEAD
        i) echo "Building ICQ Docker...    ";
=======
        i) printf '%s' "Building ICQ Docker...    ";
>>>>>>> 03e31a61 (setup relayers)
           docker build --tag stridezone:interchain-queries -f Dockerfile.icq . ;

           printf '%s' "Building ICQ Locally...    ";
           cd deps/interchain-queries; 
           go build -mod=readonly -trimpath -o $BUILDDIR ./... 2>&1 | grep -v -E "deprecated|keychain" || true; 
           cd ../..
<<<<<<< HEAD
           echo "Done" ;;         

        r) echo "Building Relayer Docker...    ";
           docker build --tag stridezone:relayer -f Dockerfile.relayer . ;

           printf '%s' "Building Relayer Locally...    ";
           cd deps/relayer; 
           go build -mod=readonly -trimpath -o $BUILDDIR ./... 2>&1 | grep -v -E "deprecated|keychain" || true; 
           cd ../..
           echo "Done" ;;     
=======
           echo "Done" ;;           
>>>>>>> 03e31a61 (setup relayers)
    esac
done
