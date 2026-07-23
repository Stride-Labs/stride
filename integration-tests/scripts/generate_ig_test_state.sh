#!/bin/bash
# generate_ig_test_state.sh
# One-time helper for the POA migration integration-tests run.
# Produces: 3 priv_validator_key.json blobs (cons keys), their lowercase
# hex consensus addresses, and operator bech32 addresses derived from
# val1/val2/val3 mnemonics. Outputs a single JSON document on stdout
# for the implementer to inject into keys.json, validators.json, and
# PoaValidatorSet.
#
# Requires: strided binary on PATH (any recent version that knows
# `strided init` and `strided keys`).

set -eu

WORKDIR=$(mktemp -d)
trap "rm -rf $WORKDIR" EXIT

KEYS_FILE="${KEYS_FILE:-integration-tests/network/configs/keys.json}"
NUM_VALIDATORS=3

# Sanity-check that keys.json has at least 3 validator entries.
existing_count=$(jq '.validators | length' "$KEYS_FILE")
if [[ "$existing_count" -lt "$NUM_VALIDATORS" ]]; then
    echo "ERROR: keys.json has only $existing_count validators; need $NUM_VALIDATORS." >&2
    exit 1
fi

# Generate 3 cons keys using `strided init`. Write priv_validator_key.json
# verbatim — same shape strided expects on disk.
declare -a CONS_KEYS_JSON
declare -a HEX_ADDRS

for ((i=0; i<NUM_VALIDATORS; i++)); do
    home="${WORKDIR}/cons-${i}"
    strided init "tmp-${i}" --chain-id ig-tests --home "$home" >/dev/null 2>&1
    cons_key=$(cat "${home}/config/priv_validator_key.json")
    hex_addr=$(jq -r '.address' <<<"$cons_key" | tr 'A-Z' 'a-z')

    CONS_KEYS_JSON[$i]="$cons_key"
    HEX_ADDRS[$i]="$hex_addr"
done

# Derive operator bech32 from each validator mnemonic (val1, val2, val3).
declare -a OPERATORS
for ((i=0; i<NUM_VALIDATORS; i++)); do
    mnemonic=$(jq -r ".validators[$i].mnemonic" "$KEYS_FILE")
    name="ig-tests-val-$i"
    keyring="${WORKDIR}/keyring-${i}"
    echo "$mnemonic" | strided keys add "$name" \
        --recover --keyring-backend test --home "$keyring" >/dev/null 2>&1
    operator=$(strided keys show "$name" -a \
        --keyring-backend test --home "$keyring")
    OPERATORS[$i]="$operator"
done

# Emit a single JSON document with the cons keys, hex addrs, and operators.
jq -n \
  --argjson cons_keys "$(printf '%s\n' "${CONS_KEYS_JSON[@]}" | jq -s .)" \
  --argjson hex_addrs "$(printf '%s\n' "${HEX_ADDRS[@]}" | jq -R . | jq -s .)" \
  --argjson operators "$(printf '%s\n' "${OPERATORS[@]}" | jq -R . | jq -s .)" \
  '{
     cons_keys: $cons_keys,
     hex_addrs: $hex_addrs,
     operators: $operators
   }'
