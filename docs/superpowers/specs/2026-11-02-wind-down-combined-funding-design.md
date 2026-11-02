# Wind-down dashboard: combined funding txs

## Goal

On 2026-11-10 every remaining zone (Hub, dYdX, Injective, Juno, Band, Terra, Sommelier, Saga) moves its ICA balances
to the Osmosis vault and funds its pools. Instead of one signing round per ICA and per zone, the Multisig tab offers
six combined txs, each a pure bundle of the existing per-zone txs (no change to any message):

| # | Tx | Chain | Members (existing per-zone txs) |
|---|----|-------|----------------------------------|
| 1 | Test transfer, all zones | Stride | every zone's ICA `test` txs |
| 2 | Full transfer, Hub | Stride | cosmoshub-4's ICA `rest` txs |
| 3 | Full transfer, all other zones | Stride | every other zone's ICA `rest` txs |
| 4 | Test join, all zones | Osmosis | every zone's `test joins` tx |
| 5 | Full join + close, Hub | Osmosis | cosmoshub-4's `close` tx |
| 6 | Full join + close, all other zones | Osmosis | every other zone's `close` tx |

## Membership and readiness

- Member candidates come from the existing sets' txs (`_ica_transfer_txs` test/rest txs, `_pool_txs` test-joins/close).
  Foreign-denom transfer txs are never members.
- Skipped (not members): done txs (`done=True`, e.g. Celestia, Osmosis, haqq), transfer txs whose reason is the
  "nothing to send" reason (an ICA below the test amount), and the `_empty_pool_tx` of a zone with no pool to fund
  only if its reason is "no pool to fund" — any other empty-pool reason (snapshot missing) holds.
- Hold: the combined tx is ready only when every member is ready. Otherwise it is not ready and its reason names the
  first unready member and its reason (`<member title>: <reason>`).
- No members at all: a done-style card ("nothing left to send") with no commands.
- Each card's `members` chips list the member tx titles.

## Generate command

1. `rm -f <stem>.unsigned.json`, then each member's existing generate text verbatim (it writes its own per-zone
   unsigned file exactly as today; a close member still runs its fresh `pool_amounts.py` read).
2. Merge, only if every member file is non-empty: `test -s f1 && test -s f2 ... && jq -s '<merge>' f1 f2 ... > <stem>.unsigned.json`.
   Files are listed explicitly in member order (message order = member order).
3. The jq merge concatenates `body.messages`, keeps the first file's body otherwise, clears signatures, and sets
   `auth_info.fee.gas_limit` and `auth_info.fee.amount[0].amount` to the combined gas and fee.

Combined gas = sum of the members' gas, capped at `COMBINED_GAS_CAP = 25_000_000` (Osmosis nodes reject a tx above
30M gas wanted by default, `osmosis-mempool.max-gas-wanted-per-tx`; real use measured ~50–125k per pool message, the
24-message TIA close used 1.18M). Combined fee = combined gas × the chain's gas price (Stride 0.005ustrd, Osmosis
0.1uosmo), computed in Python and baked into the jq expression. The generate label states message count, gas and fee.

Sign/multisign/broadcast commands: the standard `_commands` for the chain (STRIDE_TOOLS / OSMOSIS_TOOLS), stems
`/tmp/wind-down/combined-<n>-<slug>`.

## Placement

A new `TxSet` `combined-funding` ("Combined funding: every remaining zone in six txs"), placed right before
`ica-transfers`, step_id `transfers`. Notes: the order (1 → settle → 2,3 → settle → 4 → settle → 5,6), the hold rule,
the gas cap, and that the per-zone sets below stay as the fallback. `chain_id` on each card: "all zones",
"cosmoshub-4" or "all zones except cosmoshub-4". The per-zone sets are unchanged.

## Testing

- Membership: done/nothing-to-send/foreign txs excluded; Hub vs others split.
- Hold: one unready member makes the combined tx not ready with that member's reason.
- Generate text contains each member's generate text verbatim, the `test -s` guard, and the merge with the summed
  (and capped) gas and fee.
- No members → done-style card.

## Build plan

One chunk.

**Chunk 1: combined funding set** — Depends on: none.
- Files: `multisig.py`, `test_multisig.py` (and `static/multisig.js` only if a new card field is needed — avoid it;
  reuse `members`, `done`, `reason`).
- Produce `_combined_funding_set(funds_data, pools_data) -> TxSet`, inserted in `tx_sets` before `_ica_transfers_set`.
  Build it from the same per-zone MultisigTx lists the other sets use (call `_ica_transfer_txs` / `_pool_txs` once per
  zone; do not duplicate their logic). To know a member's gas, expose it without parsing shell text: add an optional
  `gas: int | None` field to `MultisigTx` set by `_stride_tx` callers (`TRANSFER_GAS`) and `_bundle_pool_tx`
  (`parts[0].gas`); also keep the member's unsigned path (`<file_stem>.unsigned.json`) available (the first entry of
  `files` or a new `unsigned` field — pick one, keep it simple).
- Distinguish skip vs hold by comparing against the existing reason builders (`_nothing_to_send_reason` text pattern
  is fragile; prefer a small `skip: bool` computed where the reason is built, or a constant prefix).
- Tests as listed above; run the full dashboard suite (`python3 -m pytest scripts/wind-down/dashboard -q`).
