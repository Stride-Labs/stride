#!/bin/bash

set -eu 

echo "Checking executable dependencies... ";
DEPENDENCIES="jq bats"
deps=0
for name in ${DEPENDENCIES}
do
   [[ $(type $name 2>/dev/null) ]] || { echo "\n    * $name is required to run this script;";deps=1; }
done
[[ $deps -ne 1 ]] && echo "OK\n" || { echo "\nInstall the missing dependencies and rerun this script...\n"; }

# add jq
if ! type "jq" > /dev/null; then
  brew install jq
fi

echo "Checking module dependencies... ";
MODULES=("gaia" "hermes" "interchain-queries")
deps=0
for module in ${MODULES}; 
do
   [ "$(ls -A ./deps/${module})" ] || { echo "\n    * $module is required to run this script;";deps=1; }
done
[[ $deps -ne 1 ]] && echo "OK\n" || { echo "\nInstall the dependency modules with \"git submodule update --init\"...\n"; }


