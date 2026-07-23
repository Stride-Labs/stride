#!/bin/bash
# generate_rehearsal_state.sh
# One-time helper for the POA migration rehearsal.
# Produces: 8 priv_validator_key.json blobs (cons keys), their hex consensus
# addresses, and operator bech32 addresses derived from each validator
# mnemonic. Outputs everything to a single JSON document on stdout for the
# implementer to copy into keys.json, validators.json, and PoaValidatorSet.
#
# Requires: strided binary on PATH (any recent version that knows
# `strided init` and `strided keys`).

set -eu

WORKDIR=$(mktemp -d)
trap "rm -rf $WORKDIR" EXIT

KEYS_FILE="${KEYS_FILE:-integration-tests/network/configs/keys.json}"
NUM_VALIDATORS=8

# Ensure 8 validator entries exist in keys.json. If only 5 are present today,
# the script generates 3 fresh mnemonics and prints them so the implementer
# can append them to keys.json before re-running.
existing_count=$(jq '.validators | length' "$KEYS_FILE")
if [[ "$existing_count" -lt "$NUM_VALIDATORS" ]]; then
    echo "ERROR: keys.json has only $existing_count validators; need $NUM_VALIDATORS." >&2
    echo "Generating $((NUM_VALIDATORS - existing_count)) fresh mnemonic(s):" >&2
    for ((i=existing_count+1; i<=NUM_VALIDATORS; i++)); do
        mnemonic=$(strided keys mnemonic 2>/dev/null)
        echo "  val${i}: ${mnemonic}" >&2
    done
    echo "Append the above to keys.json::validators[] then re-run." >&2
    exit 1
fi

# Generate 8 cons keys using `strided init`. We write the priv_validator_key.json
# verbatim — same shape strided expects on disk.
declare -a CONS_KEYS_JSON
declare -a HEX_ADDRS

for ((i=0; i<NUM_VALIDATORS; i++)); do
    home="${WORKDIR}/cons-${i}"
    strided init "tmp-${i}" --chain-id rehearsal --home "$home" >/dev/null 2>&1
    cons_key=$(cat "${home}/config/priv_validator_key.json")
    hex_addr=$(jq -r '.address' <<<"$cons_key" | tr 'A-Z' 'a-z')

    CONS_KEYS_JSON[$i]="$cons_key"
    HEX_ADDRS[$i]="$hex_addr"
done

# Derive operator bech32 from each validator mnemonic.
declare -a OPERATORS
for ((i=0; i<NUM_VALIDATORS; i++)); do
    mnemonic=$(jq -r ".validators[$i].mnemonic" "$KEYS_FILE")
    name="rehearsal-val-$i"
    # Use a fresh keyring per loop to avoid name collisions across runs.
    keyring="${WORKDIR}/keyring-${i}"
    echo "$mnemonic" | strided keys add "$name" \
        --recover --keyring-backend test --home "$keyring" >/dev/null 2>&1
    operator=$(strided keys show "$name" -a \
        --keyring-backend test --home "$keyring")
    OPERATORS[$i]="$operator"
done

# Emit a single JSON document the implementer can paste into the relevant files.
jq -n \
  --argjson cons_keys "$(printf '%s\n' "${CONS_KEYS_JSON[@]}" | jq -s .)" \
  --argjson hex_addrs "$(printf '%s\n' "${HEX_ADDRS[@]}" | jq -R . | jq -s .)" \
  --argjson operators "$(printf '%s\n' "${OPERATORS[@]}" | jq -R . | jq -s .)" \
  '{
     cons_keys: $cons_keys,
     hex_addrs: $hex_addrs,
     operators: $operators
   }'
