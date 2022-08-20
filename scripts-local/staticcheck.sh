#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

GO_VERSION=1.18

path_to_x="$SCRIPT_DIR/../x"
for d in $(find $path_to_x -maxdepth 4 -type d)
do
  echo "staticcheck $d"
  staticcheck -go $GO_VERSION $d
done

path_to_app="$SCRIPT_DIR/../app"
for d in $(find $path_to_app -maxdepth 4 -type d)
do
  echo "staticcheck $d"
  staticcheck -go $GO_VERSION $d
done