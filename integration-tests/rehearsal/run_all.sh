#!/bin/bash
# Runs the whole rehearsal in order, stopping at the first phase that exits non-zero. START_AT=<name> resumes.
cd "$(dirname "$0")"
PHASES=(phase0_preflight seed phase1_upgrade phase2_day0 phase3_redemptions phase4_drain phase5_transfers phase6_staketia phase7_pools phase8_sweep phase9_halt)
started=${START_AT:+0}; started=${started:-1}
for p in "${PHASES[@]}"; do
  [[ $started == 1 || $p == "$START_AT" ]] || continue; started=1
  echo "=== $p start $(date -u +%T)"
  bash "$p.sh"; rc=$?
  echo "=== $p exit=$rc $(date -u +%T)"
  [[ $rc == 0 ]] || exit $rc
done
echo "=== ALL PHASES DONE"
