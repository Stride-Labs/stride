#!/bin/bash
BASE_SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

run_bats=$BASE_SCRIPT_DIR/bats/bin/bats

# run test files
$run_bats $BASE_SCRIPT_DIR/basic_tests.bats
