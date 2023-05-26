#!/bin/bash

if [ -z "$UPGRADE_OLD_VERSION" ]; then
    echo "`UPGRADE_OLD_VERSION` is not specified. Please provide the old tag or commit hash using 'make UPGRADE_OLD_VERSION=<tag> upgrade-build-old-binary'"
    exit 1
fi
current_branch=$(git rev-parse --abbrev-ref HEAD)

git checkout $UPGRADE_OLD_VERSION
bash ${DOCKERNET_HOME}/build.sh -s ${BUILDDIR}
mkdir -p ${DOCKERNET_HOME}/upgrades/binaries
rm -f ${DOCKERNET_HOME}/upgrades/binaries/strided1
cp ${BUILDDIR}/strided ${DOCKERNET_HOME}/upgrades/binaries/strided1
git checkout $current_branch