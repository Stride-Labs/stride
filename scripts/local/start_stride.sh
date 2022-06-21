SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

mkdir -p $SCRIPT_DIR/logs

nohup strided start --home $SCRIPT_DIR/state/stride | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $SCRIPT_DIR/logs/stride.log 2>&1 &

# finalizing commit of block