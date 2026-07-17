#!/usr/bin/env bash
#
# trim_export.sh
#
# Trims a full `strided export` mainnet JSON down to only the modules touched
# by the v33 upgrade handler (and its assertions). Used to keep the committed
# fixture (testdata/mainnet_export.json.gz) reasonably small (under GitHub's
# 100 MB per-file limit). A raw export at current mainnet height is ~3 GB
# uncompressed; trimmed output is normally well under 50 MB compressed.
#
# Two passes:
#   1. Module-level: drop every app_state.<module> not in KEEP_MODULES.
#   2. Sub-key: empty out app_state.ccvconsumer.height_to_valset_update_id —
#      a per-block bookkeeping array (~30M entries / 1.4 GB on a chain that
#      has been a consumer for 2 years). The v33 handler only reads the
#      8-entry provider.initial_val_set, so this field is dead weight.
#
# Usage:
#   ./app/upgrades/v33/scripts/trim_export.sh app/upgrades/v33/testdata/raw_export.json app/upgrades/v33/testdata/mainnet_export.json.gz
#
# Modules KEPT in app_state (everything else is stripped):
#   ccvconsumer    snapshot source — the validator set we migrate to POA
#   bank           ICS module account balances swept to the community pool
#   auth           module account registry (resolves bank addrs)
#   distribution   community pool — sweep target + assertion
#   staking        bond denom + params
#   gov            assertion target — proposals must survive untouched
#   stakeibc       assertion target — host zones must survive untouched
#   slashing       keeper stays wired post-v33 (manager removal only)
#   evidence       keeper stays wired post-v33 (manager removal only)
#   mint           referenced by fee/reward routing
#   upgrade        RunMigrations reads consensus versions from here
#   params         legacy param subspaces
#   capability     needed for IBC-adjacent module wiring at boot
#   consensus      block params
#
# Top-level fields (chain_id, genesis_time, initial_height, validators,
# consensus_params, app_hash) are passed through unchanged. Stripped modules
# are removed entirely; the test backfills them via DefaultGenesis() if any
# code path requires them.

set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: $0 <input.json> <output.json.gz>" >&2
  exit 1
fi

INPUT="$1"
OUTPUT="$2"

if [[ ! -f "$INPUT" ]]; then
  echo "ERROR: input file not found: $INPUT" >&2
  exit 1
fi

KEEP_MODULES=(
  ccvconsumer
  bank
  auth
  distribution
  staking
  gov
  stakeibc
  slashing
  evidence
  mint
  upgrade
  params
  capability
  consensus
)

KEEP_JSON=$(printf '%s\n' "${KEEP_MODULES[@]}" | jq -R . | jq -s .)

INPUT_SIZE=$(wc -c < "$INPUT" | tr -d ' ')
echo "Input:  $INPUT  ($(numfmt --to=iec --suffix=B "$INPUT_SIZE" 2>/dev/null || echo "$INPUT_SIZE B"))"
echo "Keeping app_state.{$(IFS=,; echo "${KEEP_MODULES[*]}")}"

# Module-level filter: drop everything outside the keep-list.
# Then the second-level scrub: ccvconsumer.height_to_valset_update_id is a
# per-block array (~30M entries / ~1.4 GB on a 2-year-old consumer chain).
# The v33 handler only reads provider.initial_val_set, so this field is dead
# weight — stripping it shrinks the trimmed fixture from ~160 MB to <50 MB.
JQ_FILTER='
  .app_state |= with_entries(select(.key as $k | $keep | index($k)))
  | if .app_state.ccvconsumer? then
      .app_state.ccvconsumer.height_to_valset_update_id = []
    else . end
'

if [[ "$OUTPUT" == *.gz ]]; then
  jq --argjson keep "$KEEP_JSON" "$JQ_FILTER" "$INPUT" | gzip -9 > "$OUTPUT"
  PRESENT=$(gunzip -c "$OUTPUT" | jq -r '.app_state | keys | .[]')
else
  jq --argjson keep "$KEEP_JSON" "$JQ_FILTER" "$INPUT" > "$OUTPUT"
  PRESENT=$(jq -r '.app_state | keys | .[]' "$OUTPUT")
fi

OUTPUT_SIZE=$(wc -c < "$OUTPUT" | tr -d ' ')
echo "Output: $OUTPUT ($(numfmt --to=iec --suffix=B "$OUTPUT_SIZE" 2>/dev/null || echo "$OUTPUT_SIZE B"))"
echo
echo "Modules retained in app_state:"
echo "$PRESENT" | sed 's/^/  /'

# Warn (not fail) if any KEEP_MODULES entry was missing from the input —
# the input may be from a chain that doesn't have all modules we list.
MISSING=()
for m in "${KEEP_MODULES[@]}"; do
  if ! echo "$PRESENT" | grep -qx "$m"; then
    MISSING+=("$m")
  fi
done

if [[ ${#MISSING[@]} -gt 0 ]]; then
  echo
  echo "WARNING: these modules were requested but missing from input app_state:" >&2
  printf '  %s\n' "${MISSING[@]}" >&2
  echo "(probably fine — input may be from an older chain or have different module names)" >&2
fi
