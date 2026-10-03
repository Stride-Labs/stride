#!/bin/bash
# Recovers expired IBC light clients with substitute clients + gov MsgRecoverClient (the real ops procedure for an
# expired client). Creates substitutes with a scratch rly home in each relayer pod, so the running relayers keep their
# config (which names the subject clients, the ones the channels use) and resume updating them once recovered.
# Substitutes expire after the same 204s trusting period, so creation, proposals and votes run back to back.
#   usage: recover_clients.sh   (recovers: stride 07-tendermint-0 [hub], hub 07-tendermint-0 [stride],
#                                hub 07-tendermint-1 [osmosis], osmosis 07-tendermint-1 [hub])
source "$(dirname "$0")/lib.sh"
log "## Client recovery (expired clients after the hub OOM outage)"

STRIDE_GOV=stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl
HUB_GOV=cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn
OSMO_GOV=osmo10d07y265gmmuvt4z0w9aw880jnsr700jjeq4qp

# Creates a fresh client pair for a path in a scratch rly home; prints "<src-client> <dst-client>"
new_clients() { # relayer-deployment src-chain dst-chain path
  $KX exec deploy/$1 -- sh -c "
    rm -rf /tmp/rlyx && cp -r \$HOME/.relayer /tmp/rlyx
    rly --home /tmp/rlyx tx client $2 $3 $4 --override >/tmp/rlyx.log 2>&1 || { tail -5 /tmp/rlyx.log >&2; exit 1; }
    yq -r '.paths.\"$4\".src.client-id + \" \" + .paths.\"$4\".dst.client-id' /tmp/rlyx/config/config.yaml"
}

recover_msg() { # gov-address subject substitute
  jq -n --arg s "$1" --arg a "$2" --arg b "$3" \
    '{"@type": "/ibc.core.client.v1.MsgRecoverClient", subject_client_id: $a, substitute_client_id: $b, signer: $s}'
}

# Submits a proposal with the given messages from val1 on the chain's pod and returns the new proposal id
submit_proposal() { # chain-wrapper pod denom deposit messages-json title
  local proposal; proposal=$(jq -n --argjson m "$5" --arg d "$4$3" --arg t "$6" \
    '{messages: $m, metadata: "", deposit: $d, title: $t, summary: $t, expedited: false}')
  echo "$proposal" | $KX exec -i "$2" -c validator -- sh -c 'cat > /tmp/recover.json'
  local bin; case $1 in strided_new) bin=$NEW_BIN;; *) bin=$1;; esac
  local chainid; case $1 in strided_new) chainid=stride-test-1;; gaiad) chainid=cosmoshub-test-1;; osmosisd) chainid=osmosis-test-1;; esac
  local out hash
  out=$($KX exec "$2" -c validator -- $bin tx gov submit-proposal /tmp/recover.json --from val1 --keyring-backend test \
    --chain-id $chainid --gas 600000 --gas-prices 1$3 -y -o json 2>&1)
  hash=$(tx_hash <<<"$out"); wait_tx "$1" "$hash" >&2 || { log "proposal submit failed on $chainid: $(head -c 300 <<<"$out")"; return 1; }
  $1 q gov proposals -o json 2>/dev/null | grep -E '^\{' | jq -r '.proposals | max_by(.id | tonumber).id'
}

vote() { # pod binary chain-id denom proposal key
  $KX exec "$1" -c validator -- $2 tx gov vote "$5" yes --from "$6" --keyring-backend test --chain-id "$3" \
    --gas 200000 --gas-prices 1$4 -y -o json >/dev/null 2>&1
}

proposal_passed() { # chain-wrapper id
  [[ "$($1 q gov proposal "$2" -o json 2>/dev/null | grep -E '^\{' | jq -r '.proposal.status // .status')" == PROPOSAL_STATUS_PASSED ]]
}
client_active() { # chain-wrapper client
  [[ "$($1 q ibc client status "$2" -o json 2>/dev/null | grep -E '^\{' | jq -r .status)" == Active ]]
}

# 1. Substitutes: stride-cosmoshub path (src stride, dst hub) and cosmoshub-osmosis path (src hub, dst osmosis)
read -r S_SUB H_SUB_S <<<"$(new_clients relayer-stride-cosmoshub stride cosmoshub stride-cosmoshub)"
read -r H_SUB_O O_SUB <<<"$(new_clients relayer-cosmoshub-osmosis cosmoshub osmosis cosmoshub-osmosis)"
log "substitutes: stride[$S_SUB] hub[$H_SUB_S (stride), $H_SUB_O (osmosis)] osmosis[$O_SUB]"
[[ -n "$S_SUB" && -n "$H_SUB_S" && -n "$H_SUB_O" && -n "$O_SUB" ]] || { log "substitute creation failed"; exit 1; }

# 2. Proposals on the three chains, then votes in parallel
P_S=$(submit_proposal strided_new stride-validator-0 ustrd 20000000 \
  "[$(recover_msg $STRIDE_GOV 07-tendermint-0 "$S_SUB")]" "recover hub client")
P_H=$(submit_proposal gaiad cosmoshub-validator-0 uatom 20000000 \
  "[$(recover_msg $HUB_GOV 07-tendermint-0 "$H_SUB_S"), $(recover_msg $HUB_GOV 07-tendermint-1 "$H_SUB_O")]" "recover stride and osmosis clients")
P_O=$(submit_proposal osmosisd osmosis-validator-0 uosmo 20000000 \
  "[$(recover_msg $OSMO_GOV 07-tendermint-1 "$O_SUB")]" "recover hub client")
log "proposals: stride #$P_S hub #$P_H osmosis #$P_O"

for i in 0 1 2 3; do vote stride-validator-$i $NEW_BIN stride-test-1 ustrd "$P_S" val$((i + 1)) & done
vote cosmoshub-validator-0 gaiad cosmoshub-test-1 uatom "$P_H" val1 &
vote osmosis-validator-0 osmosisd osmosis-test-1 uosmo "$P_O" val1 &
wait

# 3. Passed and active
wait_until 120 "stride recovery passed" proposal_passed strided_new "$P_S"
wait_until 120 "hub recovery passed" proposal_passed gaiad "$P_H"
wait_until 120 "osmosis recovery passed" proposal_passed osmosisd "$P_O"
checkpoint "stride hub-client active"     client_active strided_new 07-tendermint-0
checkpoint "hub stride-client active"     client_active gaiad 07-tendermint-0
checkpoint "hub osmosis-client active"    client_active gaiad 07-tendermint-1
checkpoint "osmosis hub-client active"    client_active osmosisd 07-tendermint-1
