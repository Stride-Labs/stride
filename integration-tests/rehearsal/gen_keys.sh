#!/bin/bash
# Generates every rehearsal key once and derives the addresses the branch hard-codes.
# Uses the locally installed strided only for key derivation (any version works).
set -euo pipefail
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KEYS=$DIR/keys.json
OUT=$DIR/addresses.env
HOME_TMP=$(mktemp -d)
K="strided keys --keyring-backend test --home $HOME_TMP"

if [[ -f $KEYS ]]; then echo "$KEYS exists; delete it to regenerate"; exit 1; fi

new_key() { # name -> json {name, mnemonic}
  local mnemonic; mnemonic=$(strided keys mnemonic 2>/dev/null)
  echo "$mnemonic" | $K add "$1" --recover >/dev/null 2>&1
  jq -n --arg n "$1" --arg m "$mnemonic" '{name:$n, mnemonic:$m}'
}
addr() { $K show "$1" -a 2>/dev/null; }
to_prefix() { # stride-addr prefix -> re-encoded
  python3 - "$1" "$2" <<'EOF'
import sys; sys.path.insert(0, "scripts/wind-down")
import bech32_ref
_, raw = bech32_ref.decode(sys.argv[1]); print(bech32_ref.encode(sys.argv[2], raw))
EOF
}
ibc_denom() { echo "ibc/$(printf '%s' "$1" | shasum -a 256 | cut -d' ' -f1 | tr a-f A-F)"; }

cd "$DIR/../.."   # repo root so scripts/wind-down is importable

m=$(jq -n "[$(new_key m1),$(new_key m2),$(new_key m3)]")
d=$(jq -n "[$(new_key d1),$(new_key d2),$(new_key d3)]")
sweep=$(new_key sweep-operator)
deposit=$(new_key st-deposit); redemption=$(new_key st-redemption); claim=$(new_key st-claim)
safe=$(new_key st-safe); operator=$(new_key st-operator); reward=$(new_key st-reward)
hbase=$(new_key holder-base); hvest=$(new_key holder-vesting)

$K add admin-ms --multisig m1,m2,m3 --multisig-threshold 2 >/dev/null 2>&1
$K add hub-ms   --multisig d1,d2,d3 --multisig-threshold 2 >/dev/null 2>&1

jq -n --argjson m "$m" --argjson d "$d" --argjson sweep "$sweep" \
  --argjson deposit "$deposit" --argjson redemption "$redemption" --argjson claim "$claim" \
  --argjson safe "$safe" --argjson operator "$operator" --argjson reward "$reward" \
  --argjson hbase "$hbase" --argjson hvest "$hvest" \
  '{multisig_members:$m, hub_multisig_members:$d, sweep_operator:$sweep,
    staketia:{deposit:$deposit, redemption:$redemption, claim:$claim, safe:$safe, operator:$operator, reward:$reward},
    holders:{base:$hbase, vesting:$hvest}}' > "$KEYS"

user1=$(jq -r '.users[0].mnemonic' integration-tests/network/configs/keys.json)
echo "$user1" | $K add user1 --recover >/dev/null 2>&1

admin_ms=$(addr admin-ms); hub_ms=$(addr hub-ms)
cat > "$OUT" <<EOF
ADMIN_MS_STRIDE=$admin_ms
VAULT_MS_OSMO=$(to_prefix $admin_ms osmo)
ADMIN_MS_MEMBERS=m1,m2,m3
HUB_MS_COSMOS=$(to_prefix $hub_ms cosmos)
HUB_MS_MEMBERS=d1,d2,d3
SWEEP_OPERATOR=$(addr sweep-operator)
STAKETIA_DEPOSIT=$(addr st-deposit)
STAKETIA_REDEMPTION=$(addr st-redemption)
STAKETIA_CLAIM=$(addr st-claim)
STAKETIA_SAFE=$(addr st-safe)
STAKETIA_OPERATOR_STRIDE=$(addr st-operator)
STAKETIA_OPERATOR_HUB=$(to_prefix $(addr st-operator) cosmos)
HUB_REWARD=$(to_prefix $(addr st-reward) cosmos)
HOLDER_BASE=$(addr holder-base)
HOLDER_BASE_OSMO=$(to_prefix $(addr holder-base) osmo)
HOLDER_VESTING=$(addr holder-vesting)
HOLDER_VESTING_OSMO=$(to_prefix $(addr holder-vesting) osmo)
USER1_STRIDE=$(addr user1)
USER1_OSMO=$(to_prefix $(addr user1) osmo)
USER1_COSMOS=$(to_prefix $(addr user1) cosmos)
ATOM_ON_STRIDE=$(ibc_denom transfer/channel-0/uatom)
OSMO_ON_STRIDE=$(ibc_denom transfer/channel-1/uosmo)
EOF
rm -rf "$HOME_TMP"
cat "$OUT"
