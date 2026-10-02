#!/bin/bash
# Phase 0 pre-flight. Runs twice: right after the network starts (channels, ICA host params, REST, versions)
# and again after the seed (host-zone dependent lines: constant proofs and withdraw addresses).
source "$(dirname "$0")/lib.sh"
log "## Phase 0: pre-flight"

# Small named functions instead of `sh -c` strings: checkpoint runs them in this shell, so the lib wrappers are in scope
channel_open() { # chain-wrapper channel
  "$1" q ibc channel end transfer "$2" -o json | jq -e '.channel.state == "STATE_OPEN"'
}
channel_counterparty_chain() { # chain-wrapper channel expected-chain-id
  "$1" q ibc channel client-state transfer "$2" -o json \
    | jq -e --arg chain "$3" '.identified_client_state.client_state.chain_id == $chain'
}
ica_host_allows() { # chain-wrapper msg-type-url
  "$1" q interchain-accounts host params -o json \
    | jq -e --arg msg "$2" '.params.allow_messages | (index($msg) != null) or (index("*") != null)'
}
rest_reachable() { # host
  curl -s -m 10 "https://$1-api.internal.stridenet.co/cosmos/base/tendermint/v1beta1/node_info" \
    | jq -e '.default_node_info.network'
}
tx_ok() { # chain-wrapper tx-args... (broadcasts, then waits for the tx to land with code 0)
  local chain=$1; shift
  local hash
  hash=$("$chain" "$@" | tx_hash) && [[ -n "$hash" && "$hash" != null ]] && wait_tx "$chain" "$hash"
}
host_zone_exists() { strided_old q stakeibc show-host-zone "$1" >/dev/null 2>&1; }
withdraw_address() { # zone-chain-id host-chain-wrapper
  local delegation_ica
  delegation_ica=$(strided_old q stakeibc show-host-zone "$1" -o json | jq -er '.host_zone.delegation_ica_address')
  # distribution queries are autocli-generated in SDK 0.50+: `delegator-withdraw-address [delegator-address]`
  "$2" q distribution delegator-withdraw-address "$delegation_ica" -o json
}
withdraw_addresses() {
  withdraw_address cosmoshub-test-1 gaiad && withdraw_address osmosis-test-1 osmosisd
}

# Channels, client chain ids and ICA host allow-lists (no host zones needed)
checkpoint "stride channel-0 -> cosmoshub" channel_open strided_old channel-0
checkpoint "stride channel-0 client is cosmoshub" channel_counterparty_chain strided_old channel-0 cosmoshub-test-1
checkpoint "stride channel-1 -> osmosis" channel_counterparty_chain strided_old channel-1 osmosis-test-1
checkpoint "hub channel-1 -> osmosis" channel_counterparty_chain gaiad channel-1 osmosis-test-1
checkpoint "hub ica host allows MsgTransfer" ica_host_allows gaiad /ibc.applications.transfer.v1.MsgTransfer
checkpoint "osmo ica host allows MsgSend" ica_host_allows osmosisd /cosmos.bank.v1beta1.MsgSend
for host in stride cosmoshub osmosis; do
  checkpoint "REST $host reachable" rest_reachable "$host"
done

# Host zones exist only after the seed, so everything below is skipped on the first (post-start) run
if host_zone_exists cosmoshub-test-1; then
  # seed.sh adds vault-ms after phase 0's first run; add it here (idempotent) before the vault spend
  osmosisd keys add vault-ms --multisig "$ADMIN_MS_MEMBERS" --multisig-threshold 2 --keyring-backend test || true

  # Prove the constants: a test transfer in and a signed spend out of the vault and the sweep operator
  checkpoint "vault receives" tx_ok osmosisd tx bank send user1 "$VAULT_MS_OSMO" 1000000uosmo $OSMO_TX
  checkpoint "vault spends (multisig)" ms_tx osmosisd vault-ms "$ADMIN_MS_MEMBERS" -- bank send "$VAULT_MS_OSMO" "$USER1_OSMO" 1uosmo
  checkpoint "sweep operator spends" tx_ok strided_old tx bank send sweep-operator "$USER1_STRIDE" 1ustrd $STRIDE_TX
  log_cmd "withdraw addresses" withdraw_addresses
else
  log "host zones not seeded yet: skipping vault, sweep operator and withdraw-address checks"
fi

log "gaia $(gaiad version 2>&1 | tail -1), osmosis $(osmosisd version 2>&1 | tail -1), strided $(strided_old version 2>&1 | tail -1)"
