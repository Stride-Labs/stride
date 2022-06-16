#!/bin/bash
BASE_SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

$BASE_SCRIPT_DIR/bats/bin/bats $BASE_SCRIPT_DIR/basic_tests.bats
