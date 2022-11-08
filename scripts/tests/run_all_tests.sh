#!/bin/bash
BASE_SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# run test files
bash $BASE_SCRIPT_DIR/gaia_tests.bash
bash $BASE_SCRIPT_DIR/juno_tests.bash
bash $BASE_SCRIPT_DIR/osmo_tests.bash
bash $BASE_SCRIPT_DIR/stars_tests.bash

