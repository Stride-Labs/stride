#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

path_to_x="$SCRIPT_DIR/../x"
for d in $(find $path_to_x -maxdepth 4 -type d)
do
  echo "staticcheck $d"
  staticcheck $d
done

path_to_app="$SCRIPT_DIR/../app"
for d in $(find $path_to_app -maxdepth 4 -type d)
do
  echo "staticcheck $d"
  staticcheck $d
done