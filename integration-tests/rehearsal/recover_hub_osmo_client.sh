#!/bin/bash
# Recovers only the hub's client of osmosis (07-tendermint-1) via a substitute + gov MsgRecoverClient
source "$(dirname "$0")/lib.sh"
eval "$(sed -n '/^new_clients()/,/^}/p;/^recover_msg()/,/^}/p;/^submit_proposal()/,/^}/p;/^vote()/,/^}/p;/^proposal_passed()/,/^}/p;/^client_active()/,/^}/p' "$REHEARSAL_DIR/recover_clients.sh")"
HUB_GOV=cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn
log "## Client recovery: hub 07-tendermint-1 (osmosis) expired during the phase-5 relayer pause"
read -r H_SUB O_SUB <<<"$(new_clients relayer-cosmoshub-osmosis cosmoshub osmosis cosmoshub-osmosis)"
log "substitutes: hub[$H_SUB] osmosis[$O_SUB] (osmosis one unused)"
P=$(submit_proposal gaiad cosmoshub-validator-0 uatom 20000000 "[$(recover_msg $HUB_GOV 07-tendermint-1 "$H_SUB")]" "recover osmosis client")
log "hub proposal #$P"; vote cosmoshub-validator-0 gaiad cosmoshub-test-1 uatom "$P" val1
wait_until 120 "hub recovery passed" proposal_passed gaiad "$P"
checkpoint "hub osmosis-client active" client_active gaiad 07-tendermint-1
