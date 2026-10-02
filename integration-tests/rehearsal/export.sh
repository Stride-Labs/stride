#!/bin/bash
# usage: export.sh <out.json>. Exports Stride state from stride-validator-3 while that validator is frozen.
# Default: SIGSTOP the node, export, SIGCONT. If export fails on the database lock held by the frozen
# process, rerun with EXPORT_KILL=1: the node is killed and the export runs in the same exec
# session; the container then restarts the node (cosmovisor) and this script waits for the pod.
source "$(dirname "$0")/lib.sh"
OUT=${1:?usage: export.sh <out.json>}
POD=stride-validator-3
HOME_DIR=/home/validator/.stride
FIND_PID='pid=$(pgrep -x strided || ps -o pid,comm | awk '"'"'$2=="strided"{print $1}'"'"' | head -1)'

if stride_is_v35; then BIN=$NEW_BIN; else BIN=strided; fi
EXPORT_CMD="$BIN export --home $HOME_DIR 2>/dev/null"

export_frozen() {
  $KX exec $POD -c validator -- sh -c "$FIND_PID; kill -STOP \$pid"
  sleep 3

  # A failed export must still thaw the node, so its status is checked after the SIGCONT
  $KX exec $POD -c validator -- sh -c "$EXPORT_CMD" > "$OUT" || true
  $KX exec $POD -c validator -- sh -c "$FIND_PID; kill -CONT \$pid"
}

export_killed() {
  # kill and export in one session so the export does not race the restarted node for the lock
  $KX exec $POD -c validator -- sh -c "$FIND_PID; kill \$pid; sleep 3; $EXPORT_CMD" > "$OUT" || true
  $KX wait --for=condition=Ready pod/$POD --timeout=180s
}

if [[ "${EXPORT_KILL:-0}" == 1 ]]; then export_killed; else export_frozen; fi

jq -e '.app_state.bank.supply' "$OUT" >/dev/null || {
  log "export FAILED for $OUT (a locked database shows as empty output: rerun with EXPORT_KILL=1)"; exit 1; }
log "export written: $OUT ($(wc -c <"$OUT") bytes, binary $BIN)"
