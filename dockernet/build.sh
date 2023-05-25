#!/bin/bash

set -eu
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/config.sh

BUILDDIR="$2"
mkdir -p $BUILDDIR

# Build the local binary and docker image
build_local_and_docker() {
   set +e

   module="$1"
   folder="$2"
   title=$(printf "$module" | awk '{ print toupper($0) }')

   printf '%s' "Building $title Locally...  "

   stride_home=$PWD
   cd $folder

   # Clear any previously build binaries, otherwise the binary can get corrupted
   if [[ "$module" == "stride" ]]; then
      rm -f build/strided
   else
      rm -f build/*
   fi

   # Many projects have a "check_version" in their makefile that prevents building
   # the binary if the machine's go version does not match exactly,
   # however, we can relax this constraint
   # The following command overrides the check_version using a temporary Makefile override
   BUILDDIR=$BUILDDIR make -f Makefile -f <(echo -e 'check_version: ;') build --silent 
   local_build_succeeded=${PIPESTATUS[0]}
   cd $stride_home

   # Some projects have a hard coded build directory, while others allow the passing of BUILDDIR
   # In the event that they have it hard coded, this will copy it into our build directory
   mv $folder/build/* $BUILDDIR/ > /dev/null 2>&1

   if [[ "$local_build_succeeded" == "0" ]]; then
      echo "Done" 
   else
      echo "Failed"
      return $local_build_succeeded
   fi

   echo "Building $title Docker...  "
   if [[ "$module" == "stride" ]]; then
      image=Dockerfile
   else
      image=dockernet/dockerfiles/Dockerfile.$module
   fi

   DOCKER_BUILDKIT=1 docker build --tag stridezone:$module -f $image . 
   docker_build_succeeded=${PIPESTATUS[0]}

   if [[ "$docker_build_succeeded" == "0" ]]; then
      echo "Done" 
   else
      echo "Failed"
   fi

   set -e
   return $docker_build_succeeded
}


ADMINS_FILE=${DOCKERNET_HOME}/../utils/admins.go
ADMINS_FILE_BACKUP=${DOCKERNET_HOME}/../utils/admins.go.main

replace_admin_address() {
   cp $ADMINS_FILE $ADMINS_FILE_BACKUP
   sed -i -E "s|stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh|$STRIDE_ADMIN_ADDRESS|g" $ADMINS_FILE
}

revert_admin_address() {
   mv $ADMINS_FILE_BACKUP $ADMINS_FILE
   rm -f ${ADMINS_FILE}-E
}


# build docker images and local binaries
while getopts sgojtehrn flag; do
   case "${flag}" in
      # For stride, we need to update the admin address to one that we have the seed phrase for
      s) replace_admin_address
         if (build_local_and_docker stride .) ; then
            revert_admin_address
         else
            revert_admin_address
            exit 1
         fi
         ;;
      g) build_local_and_docker gaia deps/gaia ;;
      j) build_local_and_docker juno deps/juno ;;
      o) build_local_and_docker osmo deps/osmosis ;;
      t) build_local_and_docker stars deps/stargaze ;;
      e) build_local_and_docker evmos deps/evmos ;;
      n) continue ;; # build_local_and_docker {new-host-zone} deps/{new-host-zone} ;;
      r) build_local_and_docker relayer deps/relayer ;;  
      h) echo "Building Hermes Docker... ";
         docker build --tag stridezone:hermes -f dockernet/dockerfiles/Dockerfile.hermes . ;

         printf '%s' "Building Hermes Locally... ";
         cd deps/hermes; 
         cargo build --release --target-dir $BUILDDIR/hermes; 
         cd ../..
         echo "Done" ;;
   esac
done