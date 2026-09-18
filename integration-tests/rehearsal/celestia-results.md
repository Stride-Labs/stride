# v34 localstride rehearsal — Celestia / staketia / Cosmos Hub reconciliation (2026-09-18)

Branch under test: `v34-celestia-hub-reconciliation` at 6383ac42f (PR #1527), replayed against
real mainnet state with `localstride` (single-validator in-place testnet from the polkachu
snapshot at height 40374360, synced to 40376476 with v33.0.0).

## Procedure

1. `make setup-localstride-node` from `stride_40374360.tar.lz4`; sync with the v33.0.0 binary
   to 40376476; `STAGE=before make localstride-state-export` (1.39 GB).
2. `UPGRADE_NAME=v34 make upgrade-localstride` using **v33.0.0 plus the one-line
   `NewUncachedContext` fix from #1526** in `InitStrideAppForTestnet`. The tagged v33.0.0
   binary writes the triggered upgrade plan into a CheckTx cache that is discarded, so with it
   the node never halts. The branch binary cannot be used for this step either: the upgrade
   module panics with `BINARY UPDATED BEFORE TRIGGER` when it already carries the handler for
   a future plan.
3. Node halted at 40376486 with `UPGRADE "v34" NEEDED`. Restarted with the branch binary built
   with the local-only EndBlocker workaround (`ValidatorUpdates = nil`, see memory
   `localstride-poa-validator-changes`) so the POA swap does not crash a one-validator testnet.
4. Handler ran at 40376486, node kept producing blocks; stopped at ~40376498;
   `STAGE=after make localstride-state-export` (1.37 GB); modules extracted with
   `localstride/scratch/extract_modules.py`; checked with `celestia-verify-exports.py`.

## Handler log (relevant lines)

- `v34: celestia TotalDelegations adjusted by 15439858896 to 744230333512`
- `v34: celestia reconciled: booked 15439858896 utia of unacknowledged stake and removed the same amount from 298 delegation records`
- `v34: removed celestia deposit record 130237 (DELEGATION_IN_PROGRESS, 24492014 utia) and re-queued its leftover 9159189 as record 130319` (the split path)
- 9,342 `removed delegate callback ...` lines (active channel: false for all — every in-progress
  packet had already timed out, channel-862 was CLOSED with zero commitments, exactly the state
  the pinned-channel guard expects)
- `v34: staketia remaining delegated balance adjusted by -40076742843: 196064563213 -> 155987820370`
- `v34: cosmoshub-4 LSM deposit cosmosvaloper1jlr62.../114571 closed: 10999999 uatom booked to validator ...`
- `v34: injective-1 TotalDelegations adjusted by 200476671093651205896`
- POA swap, 12 slash-query resets, 30 stuck-ICQ deletions, gov params: all logged; `Upgrade v34 complete`
- No `NOT applied` line: every guard passed on real state.

## Export diff (`celestia-verify-exports.py`, all PASS)

| Check | Before → After |
|---|---|
| Celestia: all 89 table validators moved by exactly their delta | 0 mismatches; non-table validators untouched |
| Celestia: TotalDelegations | 728,790,474,616 → 744,230,333,512 (+15,439,858,896) |
| Celestia: open deposit records (queue + in progress) | 15,472,322,001 → 32,463,105 (−15,439,858,896 exactly) |
| Celestia: rate numerator (open records + TotalDelegations) | 744,262,796,617 → 744,262,796,617 (unchanged) |
| Celestia: delegate callbacks referencing deleted records | 0 left; `delegation_changes_in_progress` all 0 |
| Staketia: remaining_delegated_balance | 196,064,563,213 → 155,987,820,370 (−40,076,742,843 exactly) |
| Hub: stranded LSM deposit | present → gone; stakewithus 210,122,211,460 → 210,133,211,459; TotalDelegations +10,999,999 |
| Injective: delta table | applied in full, 0 mismatches |

Informational: 10,225 delegate callbacks on long-dead Celestia channels still reference records
that were already gone before the upgrade (pre-existing orphans, deliberately out of scope). The
Injective pending undelegation is not part of the stakeibc genesis export, so the export diff
cannot see it; the handler log shows it queued.

## Whole-export module comparison (`compare_modules.py`)

35 modules: 22 identical, 13 changed, 0 added/removed. Changed and why:
`stakeibc` (validator ledgers, TotalDelegations), `records` (298 celestia records − 1 split
re-queue, Hub LSM deposit), `staketia` (remaining balance), `icacallbacks` (−9,342 entries,
−19.4 MB), `interchainquery` (30 ICQs deleted), `poa` (validator swap), `gov` (quorum/period),
`auth`/`bank`/`distribution`/`staking`/`epochs`/`ratelimit` (12 blocks of normal block
production on the testnet). Nothing unexpected.

## Caveats

- The after export's CometBFT validator set does not reflect the POA swap (EndBlocker workaround).
- The snapshot post-dates channel-862's close, so the rehearsal exercised the closed-channel,
  zero-commitment path that mainnet is in today; the in-flight-callback decrement path is
  covered by unit tests and the mainnet export suite instead.

## Redemption rate, full formula recomputed from the exports

`(deposit records in transfer + undelegated deposit records + tokenized LSM deposits + TotalDelegations) / stToken supply`

| Host zone | Before | After | Δ |
|---|---|---|---|
| celestia | 1.176363576943712022 | 1.176363576943712022 | 0 (undelegated −15,439,858,896, TotalDelegations +15,439,858,896) |
| cosmoshub-4 | 1.996678110362767551 | 1.996678110362767551 | 0 (tokenized −10,999,999, TotalDelegations +10,999,999) |
| injective-1 | 1.542251109150432507 | 1.556613003698647561 | +0.0144 (+0.93%), by design of #1526: the 200.48 INJ was real stake in no bucket; the spec's runbook requires `max_inner_redemption_rate ≥ 1.575` on upgrade day (1.56871 today, 0.012 headroom) |

Stored `redemption_rate` fields are untouched by the handler (they refresh at the day epoch).
