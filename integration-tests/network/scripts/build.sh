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

    local_tag=stride-tests:$dockerfile_suffix
    global_tag=$GCR_REPO/${image_name}

    echo "Building docker image: $dockerfile_suffix"
	docker buildx build --platform $PLATFORM --tag $local_tag -f ${DOCKERFILES}/Dockerfile.$dockerfile_suffix $context
	docker tag $local_tag $global_tag

	echo "Pushing image to GCR: $global_tag"
	docker push $global_tag
}

main() {
    # For stride, we have to update the admin address
    if [[ "$CHAIN" == "stride" ]]; then
        # Trap SIGINT (Control + C) to cleanup admins file
        trap 'echo "Interrupt received, cleaning up..."; git checkout -- $ADMINS_FILE && rm -f ${ADMINS_FILE}-E; exit' INT

        # Update the admin address
        admin_address=$(jq -r '.admin.address' $KEYS_FILE)
        sed -i -E "s|stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh|$admin_address|g" $ADMINS_FILE 

        # First build the main dockerfile in the repo root, then build the integration test specific file
        docker buildx build --platform linux/amd64 --tag core:stride ..
        build_and_push_docker stride . chains/stride:latest

        # Cleanup the admins file
        git checkout -- $ADMINS_FILE && rm -f ${ADMINS_FILE}-E
    else
        echo "ERROR: Chain not supported"
        exit 1
    fi
}

main
