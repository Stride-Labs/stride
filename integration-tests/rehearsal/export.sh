#!/bin/bash
# usage: export.sh <out.json>. Exports Stride state from stride-validator-3 without stopping the node for long.
# strided is SIGSTOPped just long enough to copy config/ and data/ inside the pod (a consistent snapshot of the frozen
# databases), then SIGCONTed. The export runs against the copy: goleveldb's LOCK file is removed from the copy first,
# since the copy of a held lock would otherwise make the export refuse to open the database.
source "$(dirname "$0")/lib.sh"
OUT=${1:?usage: export.sh <out.json>}
POD=stride-validator-3
EXPORT_HOME=/tmp/exphome
# The node process only ("strided start"): a bare pgrep can match a concurrent "strided q ..." and leave the node running
FIND_PID='pid=$(ps -o pid,args | grep "bin/strided start" | grep -v grep | awk "{print \$1}" | head -1)'

# The pod's own DAEMON_HOME wins; the fallback is the default layout of the test network
HOME_DIR=$($KX exec $POD -c validator -- sh -c 'echo $DAEMON_HOME' 2>/dev/null)
HOME_DIR=${HOME_DIR:-/home/validator/.stride}

if stride_is_v35; then BIN=$NEW_BIN; else BIN=strided; fi

# SIGSTOP, copy, SIGCONT in one exec session: the node is thawed even if the copy fails, and never stays frozen
# longer than the cp
$KX exec $POD -c validator -- sh -c "
  rm -rf $EXPORT_HOME && mkdir -p $EXPORT_HOME
  $FIND_PID
  [ -n \"\$pid\" ] || exit 1
  kill -STOP \$pid
  status=0
  mkdir -p $EXPORT_HOME/data && cp -a $HOME_DIR/config $EXPORT_HOME/ && cp -a $HOME_DIR/data/application.db $EXPORT_HOME/data/ && { [ -d $HOME_DIR/data/wasm ] && cp -a $HOME_DIR/data/wasm $EXPORT_HOME/data/ || true; } || status=\$?
  kill -CONT \$pid
  exit \$status"

$KX exec $POD -c validator -- sh -c "find $EXPORT_HOME/data -name LOCK -exec rm -f {} +"
# SDK 0.50+ export writes nothing to stdout unless --output-document is given
$KX exec $POD -c validator -- sh -c "$BIN export --home $EXPORT_HOME --output-document $EXPORT_HOME/out.json >/dev/null 2>&1 && cat $EXPORT_HOME/out.json" > "$OUT" || true
$KX exec $POD -c validator -- sh -c "rm -rf $EXPORT_HOME"

jq -e '.app_state.bank.supply' "$OUT" >/dev/null || {
  log "export FAILED for $OUT (empty or invalid export output)"; exit 1; }
log "export written: $OUT ($(wc -c <"$OUT") bytes, binary $BIN)"
