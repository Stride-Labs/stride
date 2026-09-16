# v34 POA validator swap — k8s rehearsal results (2026-09-17)

Branch `v34-poa-rehearsal` (throwaway, off #1526 @ `349108784`). Network: v33.0.0 → v34 (branch),
five Stride nodes with only val1–val3 in the genesis POA set (`values.yaml: poaGenesisValidators: 3`,
power 274523 to mirror mainnet); no host chains, no relayers. Driver: `integration-tests/rehearsal/poa.sh`.

The handler's constants were swapped for the test network only: `IncomingValidators` = val4/val5
with the consensus pubkeys read from the running pods (`measure`), `OutgoingMonikers` = val2/val3,
and `utils.PoaValidatorSet` gained val4/val5 payout entries. `SwapPoaValidators` itself is the PR's
code, unmodified.

## Outcome: the swap verified end to end

| Step | Evidence |
|---|---|
| Pre-upgrade | POA set `val1 val2 val3` at 274523 each, CometBFT set size 3, val4/val5 synced with voting power 0, blocks signed by exactly val1–3 |
| Upgrade (expedited gov, h=394) | `v34: adding POA validator val4`, `v34: adding POA validator val5`, `v34: removing POA validator val2`, `v34: removing POA validator val3`, `Upgrade v34 complete`; the Injective and slash-query steps no-op'd on the missing host zones as designed |
| POA state after | active set `val1 val4 val5` at 274523 each; val2/val3 entries remain in the store with power 0 (`UpdateValidator` sets power 0, it does not delete — the query omits the zero from JSON) |
| CometBFT after | validator set size 3; `/status` voting power 274523 on val1/val4/val5 and 0 on val2/val3; all five nodes still synced (val2/val3 keep running as full nodes) |
| Block signing | five consecutive commits (h=461…490) each signed by exactly val1, val4, val5 — never by val2 or val3 |
| Gov params | voting period 120h after the upgrade (`UpdateGovParams`) |

No chain halt or consensus stall at the swap: the outgoing pair and the incoming pair changed in the
same block, and the incoming nodes (already peered and synced) began signing immediately.

## Notes for the real upgrade

- The incoming validators' nodes must be running, synced and peered **before** the upgrade height —
  the swap takes effect one block after the handler and the post-swap set immediately needs their
  votes toward the 2/3 threshold. Confirm on upgrade day that both nodes are up and that
  `strided tendermint show-validator` on each matches its `IncomingValidators` pubkey.
- Removed validators stay in POA state at power 0 (harmless; they can be deleted later with
  `poa update-validators` if a clean list is wanted).

## Harness additions (this branch only)

- `values.yaml` / `templates/validator.yaml` / `scripts/init-chain.sh`: `poaGenesisValidators`
  (nodes beyond it run with generated keys but start outside the POA set) and `poaValidatorPower`.
- `rehearsal/poa.sh`: `measure` pastes the pods' consensus pubkeys into `constants.go`;
  `build-and-swap`, `upgrade` (laptop-side expedited proposal) and `verify` as in the Injective
  rehearsal, plus a `status` dump. Zero POA power is absent from the query JSON — treat null as 0.
