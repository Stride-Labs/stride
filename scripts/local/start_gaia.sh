SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

mkdir -p $SCRIPT_DIR/logs

nohup gaiad start --home $SCRIPT_DIR/state/gaia | sed -r -u "s/\x1B\[([0-9]{1,3}(;[0-9]{1,2})?)?[mGK]//g" > $SCRIPT_DIR/logs/gaia.log 2>&1 &
echo $! >> $SCRIPT_DIR/pids.txt