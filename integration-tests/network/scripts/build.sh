#!/bin/bash
set -eu

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
STRIDE_HOME=$SCRIPT_DIR/../../..

CHAIN=$1

PLATFORM=linux/amd64
GCR_REPO=gcr.io/stride-nodes/integration-tests
ADMINS_FILE=${STRIDE_HOME}/utils/admins.go
KEYS_FILE=${STRIDE_HOME}/integration-tests/network/configs/keys.json
DOCKERFILES=${STRIDE_HOME}/integration-tests/dockerfiles

# Builds and pushes a docker image to GCR
build_and_push_docker() {
    dockerfile_suffix=$1
    context=$2
    image_name=$3
    upgrade_name=${4:-}  # Optional

    local_tag=stride-tests:$dockerfile_suffix
    global_tag=$GCR_REPO/${image_name}

    echo "Building docker image: $dockerfile_suffix"
    dockerfile=${DOCKERFILES}/Dockerfile.$dockerfile_suffix
    if [[ "$upgrade_name" != "" ]]; then
        docker buildx build --platform $PLATFORM --build-arg upgrade_name=$upgrade_name --tag $local_tag -f $dockerfile $context
    else
        docker buildx build --platform $PLATFORM --tag $local_tag -f $dockerfile $context
    fi

	echo "Pushing image to GCR: $global_tag"
	docker tag $local_tag $global_tag
	docker push $global_tag
}

main() {
    # For stride, we have to update the admin address
    if [[ "$CHAIN" == "stride" ]]; then
        # Trap SIGINT (Control + C) to cleanup admins file
        trap 'echo "Interrupt received, cleaning up..."; git checkout -- $ADMINS_FILE && rm -f ${ADMINS_FILE}-E; exit' INT

        # If an upgrade old version was specified, build that old dockerfile
        if [  "${UPGRADE_OLD_VERSION:-}" != "" ]; then
            echo "UPGRADE ENABLED: Building old version..."
            if ! git diff-index --quiet HEAD --; then
                echo "ERROR: There are uncommitted changes. Please commit all changes in the current branch before proceeding with this script."
                exit 1
            fi

            current_branch=$(git rev-parse --abbrev-ref HEAD)

            git checkout $UPGRADE_OLD_VERSION
            docker buildx build --platform linux/amd64 --tag core:stride-upgrade-old ..
            git checkout $current_branch
        fi

        # Update the admin address and build the main dockerfile in the repo root
        admin_address=$(jq -r '.admin.address' $KEYS_FILE)
        sed -i -E "s|stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh|$admin_address|g" $ADMINS_FILE 

        docker buildx build --platform linux/amd64 --tag core:stride ..

        # Then build and push the integration test dockerfile
        if [[ "${UPGRADE_OLD_VERSION:-}" != "" ]]; then
            upgrade_name=$(ls ${STRIDE_HOME}/app/upgrades | sort -V | tail -1)
            echo "Setting up upgrade test from $UPGRADE_OLD_VERSION to $upgrade_name"
            build_and_push_docker stride-upgrade . chains/stride:latest $upgrade_name
        else
            build_and_push_docker stride . chains/stride:latest
        fi

        # Cleanup the admins file
        git checkout -- $ADMINS_FILE && rm -f ${ADMINS_FILE}-E
    else
        echo "ERROR: Chain not supported"
        exit 1
    fi
}

main
