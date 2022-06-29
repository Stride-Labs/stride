### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source ${SCRIPT_DIR}/../../account_vars.sh

# 
# GAIA validators 
# stridevaloper16vlrvd7lsfqg8q7kyxcyar9v7nt0h99phg85n9
# stridevaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrgpwsqm
# stridevaloper1ld5ewfgc3crml46n806km7djtr788vqdwpm0s3

# add some validators 
$STRIDE1_EXEC tx stakeibc add-validator gaia 

