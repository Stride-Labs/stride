#!/bin/bash

set -euo pipefail

VERSION_REGEX='v[0-9]{1,2}$'
PACKAGE_PREFIX="github.com/Stride-Labs/stride"

# Validate script parameters
if [ -z "$OLD_VERSION" ]; then
    echo "OLD_VERSION must be set (e.g. v8). Exiting..."
    exit 1
fi

if [ -z "$NEW_VERSION" ]; then
    echo "NEW_VERSION must be set (e.g. v9). Exiting..."
    exit 1
fi

if ! echo $OLD_VERSION | grep -Eq $VERSION_REGEX; then 
    echo "OLD_VERSION must be of form v{major} (e.g. v8). Exiting..."
    exit 1
fi 

if ! echo $NEW_VERSION | grep -Eq $VERSION_REGEX; then 
    echo "NEW_VERSION must be of form v{major} (e.g. v9). Exiting..."
    exit 1
fi 

if [ "$(basename "$PWD")" != "stride" ]; then
    echo "The script must be run from the project home directory. Exiting..."
    exit 1
fi

# Update package name
echo ">>> Updating package name..."
update_version() {
    file=$1
    sed -i "s|$PACKAGE_PREFIX/$OLD_VERSION|$PACKAGE_PREFIX/$NEW_VERSION|g" $file
}

for parent_directory in "app" "cmd" "proto" "testutil" "third_party" "utils" "x"; do
    for file in $(find $parent_directory -type f \( -name "*.go" -o -name "*.proto" \)); do 
        echo "Updating version in $file"
        update_version $file
    done
done

update_version go.mod
update_version ./scripts/protocgen.sh

echo ">>> Committing changes..."

git add .
git commit -m "updated package from $OLD_VERSION -> $NEW_VERSION"

# Re-generate protos
echo ">>> Rebuilding protos..."

make proto-all

git add .
git commit -m 'generated protos'

echo "Done"