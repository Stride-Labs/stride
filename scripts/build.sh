#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/vars.sh

BUILDDIR="$2"

build_local_and_docker() {
   module="$1"
   folder="$2"
   title=$(printf "$module" | awk '{ print toupper($0) }')

   echo "Building $title Docker...  "
   docker build --tag stridezone:$module -f Dockerfile.$module . 

   printf '%s' "Building $title Locally...  "
   cwd=$PWD
   cd $folder
   go build -mod=readonly -trimpath -o $BUILDDIR ./... 2>&1 | grep -v -E "deprecated|keychain" || true ;
   cd $cwd
   echo "Done" 
}

# build docker images and local binaries
while getopts sgojhir flag; do
   case "${flag}" in
      s) build_local_and_docker stride . ;;
      g) build_local_and_docker gaia deps/gaia ;;
      j) build_local_and_docker juno deps/juno ;;
      o) build_local_and_docker osmo deps/osmosis ;;
      i) build_local_and_docker icq deps/interchain-queries ;;
      r) build_local_and_docker relayer deps/relayer ;;  
      h) echo "Building Hermes Docker... ";
         docker build --tag stridezone:hermes -f Dockerfile.hermes . ;

         printf '%s' "Building Hermes Locally... ";
         cd deps/hermes; 
         cargo build --release --target-dir $BUILDDIR/hermes; 
         cd ../..
         echo "Done" ;;
   esac
done