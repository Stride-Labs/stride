SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

mkdir -p $SCRIPT_DIR/logs

nohup strided start --home $SCRIPT_DIR/state/stride > $SCRIPT_DIR/logs/stride.log 2>&1 &
echo $! >> $SCRIPT_DIR/pids.txt