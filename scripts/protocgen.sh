#!/usr/bin/env bash

set -eo pipefail

generate_protos() {
  package="$1"
  proto_dirs=$(find $package -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
  for dir in $proto_dirs; do
    for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
      if grep go_package "$file" &>/dev/null; then
        buf generate --template buf.gen.gogo.yaml "$file"
      fi
    done
  done
}

echo "Generating gogo proto code"
cd proto

generate_protos "./stride"
generate_protos "./cosmos"
generate_protos "./cosmwasm"

cd ..

# move proto files to the right places
#
# Note: Proto files are suffixed with the current binary version.
cp -r github.com/Stride-Labs/stride/v14/* ./
rm -rf github.com
