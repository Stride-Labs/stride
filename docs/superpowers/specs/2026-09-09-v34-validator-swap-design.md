# v34: POA Validator Swap (Citadel.one, Cosmostation → cosmosrescue, Citizen Web3)

**Date:** 2026-09-09
**Status:** Approved

## §1. Goal and scope

Replace two validators in Stride's POA set via the v34 chain upgrade:

- **Out:** `Citadel.one`, `Cosmostation` (POA monikers, verified against mainnet POA state)
- **In:** `cosmosrescue`, `Citizen Web3`

Two places encode validator identity and must change together:

1. **POA on-chain state** (`x/poa`, `cosmos-sdk/enterprise/poa v1.0.0`) — who signs blocks and
   who receives the 15% fee/inflation slice distributed by the POA module.
2. **`utils/poa.go` → `PoaValidatorSet`** — the hardcoded list that
   `x/stakeibc/keeper/reward_allocation.go` iterates to pay out the stToken share of
   liquid-staking revenue.

Both are changed in the same upgrade so the signing set and the payout set never diverge.

**Non-goals (deferred):**

- ICS keeper/store cleanup promised for "v34" in the v33 design doc — slides to v35.
- Module-path bump `/v33 → /v34` — lands as a separate end-of-cycle "v34 version" PR,
  matching the v33 convention (see commit `6a39d565e`).
- Power normalization (274523 → 1). Deliberately **not** doing this: keeping current power
  means the 6 continuing validators are never touched, shrinking the emitted ABCI update
  set to 4 entries and structurally eliminating the duplicate-address hazard (§3).
- Relayer/ops coordination and validator node setup — release-checklist items, not code.

## §2. Constraints established by consensus-safety review

A three-way code review (POA keeper, CometBFT v0.39.3, SDK/ibc-go plumbing) established
the following. These are load-bearing constraints on the handler design:

1. **One power change per consensus address per block.** POA's `queuedUpdates` is an
   append-only transient `collections.Vec` with no dedup
   (`x/poa/keeper/keeper.go`, `validator.go:79-88`); CometBFT rejects a duplicate
   consensus address in one block's update set (`types/validator_set.go:444-447`) and a
   rejected set is a deterministic panic on every node **after** the block is WAL-fsync'd
   — an unrecoverable coordinated halt. The handler must touch each validator's power at
   most once.
2. **The incoming consensus keys must be ed25519.** Consensus params allow only ed25519;
   `keeper.CreateValidator` (unlike the msg server) does **not** validate pubkey type, so
   a wrong key type sails through the handler and halts the chain at EndBlock. The
   handler must guarantee the key type itself.
3. **`CreateValidator` must pass `checkpoint=true`** so pending unallocated fees are
   checkpointed before the new validator becomes eligible (genesis is the only
   `checkpoint=false` caller).
4. **Power-0 is the removal encoding.** The record is soft-deleted (retained at power 0);
   the removed validator's accrued fees stay withdrawable via its operator address.
   `GetAllValidators` will report 10 records (8 active + 2 at power 0) post-upgrade.
5. Everything else is safe without special handling: no CometBFT limit on per-block power
   change; proposer-priority rescale is clamped; fault tolerance stays 6-of-8; upgrade
   handler (PreBlocker) writes reach POA's EndBlock in the same block via the transient
   store; the ICS democracy-staking wrapper discards staking's validator updates so POA
   remains the sole update source; counterparty light clients verify against the trusted
   set's recorded powers (6/8 continuing signers ≈ 75% ≫ 1/3 needed). New powers take
   signing effect at upgrade height + 2.

## §3. Validator data

| | cosmosrescue | Citizen Web3 |
|---|---|---|
| POA moniker | `cosmosrescue` | `Citizen Web3` |
| Consensus pubkey (ed25519, base64) | `JEREY43D2nFKgcFQYAVUHcaKsaD14wmpFpGtXWrNK3c=` * | `rH8ddv5Ev2eIUTDx4x0ESGm7IFkskbdaSE30liWu02M=` * |
| Payout / operator address | **placeholder** | **placeholder** |
| Power | 274523 (matches current uniform set) | 274523 |

\* Candidate keys pulled from the validators' registered Stride staking (govenator)
records — right type, not yet confirmed as the keys their nodes actually run. Constants
ship as **placeholder sentinels** until each validator confirms (a) the consensus pubkey
via `strided tendermint show-validator` and (b) a fresh payout address. Neither candidate
key collides with any key in the current POA set.

Outgoing validators are identified by **moniker only** (`Citadel.one`, `Cosmostation`);
their consensus addresses are resolved from live POA state at upgrade time (approach A:
no hand-transcribed cons addresses to typo).

## §4. Single source of truth for addresses

`utils/poa.go` remains the registry of payout addresses:

- Replace the `Citadel.one` and `Cosmostation` entries with `cosmosrescue` and
  `Citizen Web3` (operator = payout address, placeholder until confirmed).
  Payout placeholders are **valid-bech32 deterministic burn addresses** (sha256-derived,
  no known private key), not a raw `"PLACEHOLDER"` string — `reward_allocation.go`
  and its tests call `sdk.MustAccAddressFromBech32` on every operator, so an
  unparseable placeholder would panic them. The handler detects them by exact
  equality (`utils.IsPlaceholderOperator`).
- **Delete the `HubAddress` field** from `PoaValidator` and all entries — dead since the
  ICS migration.
- The v34 handler **joins against `utils.PoaValidatorSet` by moniker** to obtain each
  incoming validator's operator address (the same join v33's
  `SnapshotValidatorsFromICS` used). The payout address therefore lives in exactly one
  place; `v34/constants.go` holds only the consensus pubkeys, monikers, outgoing
  monikers, and power.

## §5. Handler design (`app/upgrades/v34/`)

**`constants.go`:**

```go
const UpgradeName = "v34"

const PlaceholderSentinel = "PLACEHOLDER"

// Incoming: moniker (must match the utils.PoaValidatorSet entry) + base64 ed25519 pubkey.
var IncomingValidators = []IncomingValidator{
    {Moniker: "cosmosrescue", ConsPubKeyBase64: PlaceholderSentinel},
    {Moniker: "Citizen Web3", ConsPubKeyBase64: PlaceholderSentinel},
}

var OutgoingMonikers = []string{"Citadel.one", "Cosmostation"}

const ValidatorPower = int64(274523) // matches the current uniform POA set
```

**`upgrades.go` — `CreateUpgradeHandler(mm, configurator, cdc codec.Codec, poaKeeper *poakeeper.Keeper)`**
(POA keeper passed by pointer, mirroring v33 wiring in `app/upgrades.go`; `cdc` is
needed to unpack stored consensus-pubkey `Any`s when resolving outgoing validators'
consensus addresses — `GetAllValidators` returns values only, not store keys). After
`RunMigrations`, in order:

1. **Sentinel guard.** Error if any incoming pubkey is the sentinel, or if the joined
   operator address from `utils.PoaValidatorSet` is a placeholder / fails bech32
   validation. An unfilled release fails its dry-run tests, never mainnet.
2. **Resolve outgoing.** Walk POA validators; match each of `OutgoingMonikers` to exactly
   one consensus address. Error if missing or matched more than once.
3. **Create incoming (before removals).** For each incoming validator: decode base64 into
   an `ed25519.PubKey` (ed25519 by construction — closes constraint §2.2), wrap in
   `codectypes.Any`, join moniker → operator address from `utils.PoaValidatorSet`, then
   `poaKeeper.CreateValidator(ctx, consAddr, validator, true /* checkpoint */)` with
   `Power: ValidatorPower`.
4. **Zero outgoing.** For each resolved cons address:
   `poaKeeper.UpdateValidator(ctx, consAddr, poatypes.Validator{Power: 0})` — nil
   `Metadata`/`PubKey` so the stored record is preserved and the emitted removal update
   carries the stored pubkey.

Each consensus address is touched exactly once → exactly 4 ABCI entries in the upgrade
block (constraint §2.1 satisfied by construction). Any error aborts the upgrade handler,
which halts the upgrade loudly before commit — the correct failure mode.

**Error handling philosophy:** no silent fallbacks. Every lookup failure, decode failure,
or count mismatch is a returned error.

## §6. Tests

**Synthetic suite (`app/upgrades/v34/upgrades_test.go`):** seed the test app's POA store
with 8 validators (generated ed25519 keys, monikers matching `utils.PoaValidatorSet`'s
*pre-change* monikers for the outgoing pair), run the handler with test-filled constants,
assert:

- 8 active (power > 0) validators post-upgrade; total power unchanged (8 × 274523).
- Outgoing pair present at power 0 (records retained); incoming pair present at 274523
  with correct monikers and operator addresses.
- **`ReapValidatorUpdates()` returns exactly 4 entries with unique consensus addresses**
  — the direct regression guard against the chain-halt hazard.
- Failure cases: outgoing moniker absent → error; sentinel unfilled → error; incoming
  moniker missing from `utils.PoaValidatorSet` → error.

**Mainnet export suite (`app/upgrades/v34/mainnet_export_test.go`):** extend the v33
harness pattern — replay the handler against a real `strided export`, same assertions.
Skips gracefully when the fixture is absent; while placeholders are unfilled it fails,
doubling as the release gate. Release flow must run it with the fixture present before
tagging.

**v33 test freeze:** v33's `helpers_test.go` and `mainnet_export_test.go` currently join
against the live `utils.PoaValidatorSet` and would break when it changes. Give the v33
test package its own frozen copy of the old 8-entry set (v33 is historical; its tests
must not track a moving registry).

**No changes needed:** `x/stakeibc/keeper/reward_allocation_test.go` iterates the slice
dynamically.

## §7. Wiring and bookkeeping

- `app/upgrades.go`: register the v34 handler, passing `app.POAKeeper` (v33 pattern).
- `CHANGELOG.md`: entry under v34.

## §8. Release checklist (outside this change)

- Fill placeholders: confirmed consensus pubkeys (validator runs
  `strided tendermint show-validator` and echoes the key back in writing) + fresh payout
  addresses for both incoming validators; re-run the mainnet export suite.
- Both incoming nodes synced and running with the confirmed keys **before** the upgrade
  height — during the transition consensus needs 6-of-8 and the 6 continuing validators
  provide exactly that (zero margin until the new nodes sign at height +2).
- Notify outgoing validators: accrued POA fees remain withdrawable after removal;
  stToken payouts stop at the upgrade height.
- Relayers: update counterparty clients promptly around the upgrade; check no client is
  near expiry.
- Expect proposer-order churn in dashboards for a rotation or two; the new validators
  propose immediately (join penalty erased by priority rescale — expected).
