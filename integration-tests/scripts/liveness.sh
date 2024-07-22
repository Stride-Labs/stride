set -e

BINARY=strided
STRIDE_HOME=/home/stride/.stride

# If chain hasn't been initialized yet, exit immediately
if [ ! -d $STRIDE_HOME/config ]; then
    echo "READINESS CHECK FAILED - Chain has not been initialized yet."
    exit 1
fi

# Check that the node is running
if ! $($BINARY status &> /dev/null); then
    echo "READINESS CHECK FAILED - Chain is down"
    exit 1
fi

# Then check if the node is synced according to it's status query
CATCHING_UP=$($BINARY status 2>&1 | jq ".SyncInfo.catching_up")
if [[ "$CATCHING_UP" != "false" ]]; then
    echo "READINESS CHECK FAILED - Node is still syncing"
    exit 1
fi