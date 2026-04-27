#!/usr/bin/env bash
#
# fetch_validator_monikers.sh
#
# Builds the v33 validator metadata JSON by:
# 1. Reading the spreadsheet (validator name + Hub valoper + Hub valcons + Stride payment addr).
# 2. Fetching the live Stride active validator set (CometBFT level) via Stride RPC.
# 3. Fetching the Hub's ICS provider key-assignment pairs (provider valcons → consumer pubkey)
#    via gRPC.
# 4. Joining: spreadsheet row → Hub valcons → consumer pubkey → Stride hex cons addr.
#
# Output (stdout): flat JSON object mapping lowercase hex consensus address
# to validator moniker. Maps 1:1 to the ValidatorMonikers map in constants.go.
# {
#   "94532a622413b9ef925fb078805f2e4ced675018": "Polkachu",
#   "56b2473d1218a34d8e2c5350abd839a8243f78b1": "L5",
#   ...
# }
#
# Refuses to emit output if any spreadsheet row fails to resolve, or if the live
# validator count differs from the spreadsheet row count.
#
# Requirements: curl, jq, grpcurl

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SPREADSHEET="${SCRIPT_DIR}/spreadsheet.csv"

STRIDE_RPC="${STRIDE_RPC:-https://stride-rpc.polkachu.com}"
HUB_GRPC="${HUB_GRPC:-cosmos-grpc.publicnode.com:443}"
STRIDE_CONSUMER_ID="${STRIDE_CONSUMER_ID:-1}"

if [[ ! -f "$SPREADSHEET" ]]; then
  echo "ERROR: spreadsheet not found at $SPREADSHEET" >&2
  exit 1
fi

# 1. Parse spreadsheet (skip header). Build JSON array of rows.
SPREADSHEET_JSON=$(jq -Rn '
  [inputs | split(",") | select(.[0] != "moniker" and length == 4) | {
    moniker: .[0],
    hub_valoper: .[1],
    hub_valcons: .[2],
    stride_payment_addr: .[3]
  }]
' "$SPREADSHEET")

ROW_COUNT=$(echo "$SPREADSHEET_JSON" | jq 'length')
if [[ "$ROW_COUNT" -ne 8 ]]; then
  echo "ERROR: spreadsheet has $ROW_COUNT rows (expected 8)" >&2
  exit 1
fi

# 2. Fetch active block-producing validators from Stride RPC.
STRIDE_VALIDATORS=$(curl -sS --max-time 15 "${STRIDE_RPC}/validators?per_page=100" \
  | jq '[.result.validators[] | {hex_cons_addr: .address, pubkey_b64: .pub_key.value, power: (.voting_power | tonumber)}]')

ACTIVE_COUNT=$(echo "$STRIDE_VALIDATORS" | jq 'length')
if [[ "$ACTIVE_COUNT" -ne 8 ]]; then
  echo "ERROR: Stride RPC reports $ACTIVE_COUNT active validators (expected 8)" >&2
  exit 1
fi

# 3. Fetch Hub ICS provider pairs (provider valcons → consumer pubkey).
PAIRS_RAW=$(grpcurl -d "{\"consumer_id\": \"${STRIDE_CONSUMER_ID}\"}" \
  "${HUB_GRPC}" interchain_security.ccv.provider.v1.Query/QueryAllPairsValConsAddrByConsumer 2>&1)

# grpcurl outputs camelCase; normalize.
PAIRS=$(echo "$PAIRS_RAW" | jq '[.pairValConAddr[] | {provider_address: .providerAddress, consumer_pubkey_b64: .consumerKey.ed25519}]')

PAIR_COUNT=$(echo "$PAIRS" | jq 'length')
if [[ "$PAIR_COUNT" -lt 8 ]]; then
  echo "ERROR: Hub gRPC returned $PAIR_COUNT pairs (expected at least 8)" >&2
  exit 1
fi

# 4. Join: spreadsheet → pair (by hub_valcons) → stride validator (by pubkey).
RESULT=$(jq -n \
  --argjson sheet "$SPREADSHEET_JSON" \
  --argjson pairs "$PAIRS" \
  --argjson stride "$STRIDE_VALIDATORS" '
  $sheet | map(
    . as $row |
    ($pairs | map(select(.provider_address == $row.hub_valcons)) | first) as $pair |
    if $pair == null then
      $row + {_ERROR: "no Hub→consumer key pair for hub_valcons \($row.hub_valcons)"}
    else
      ($stride | map(select(.pubkey_b64 == $pair.consumer_pubkey_b64)) | first) as $sv |
      if $sv == null then
        $row + {_ERROR: "Hub returned consumer key \($pair.consumer_pubkey_b64) but no live Stride validator has it"}
      else
        {
          moniker: $row.moniker,
          hex_cons_addr: ($sv.hex_cons_addr | ascii_downcase)
        }
      end
    end
  )
')

# 5. Refuse to emit output if any row failed.
ERROR_COUNT=$(echo "$RESULT" | jq '[.[] | select(._ERROR)] | length')
if [[ "$ERROR_COUNT" -ne 0 ]]; then
  echo "ERROR: $ERROR_COUNT spreadsheet rows could not be resolved:" >&2
  echo "$RESULT" | jq '.[] | select(._ERROR)' >&2
  exit 1
fi

# 6. Sanity check: every Stride active validator should be claimed by exactly one row.
CLAIMED_HEX=$(echo "$RESULT" | jq -r '.[] | .hex_cons_addr' | sort)
ACTIVE_HEX=$(echo "$STRIDE_VALIDATORS" | jq -r '.[] | .hex_cons_addr | ascii_downcase' | sort)
if [[ "$CLAIMED_HEX" != "$ACTIVE_HEX" ]]; then
  echo "ERROR: spreadsheet maps to a different set of validators than the live active set." >&2
  echo "Spreadsheet → these hex addresses:" >&2
  echo "$CLAIMED_HEX" >&2
  echo "Live active set:" >&2
  echo "$ACTIVE_HEX" >&2
  exit 1
fi

# Collapse to flat {hex: moniker} object — 1:1 with the Go ValidatorMonikers map.
echo "$RESULT" | jq 'map({(.hex_cons_addr): .moniker}) | add'
