# Stride v33 — POA Migration on the Standard Integration-Tests Network

## §1. Overview and objective

Run the v32 → v33 (ICS → POA) migration on the standard k8s
`integration-tests` network — the one with 3 Stride validators plus
`cosmoshub` and `osmosis` host zones — then run `make test-core`
against the post-upgrade chain.

This complements the `poa-migration-rehearsal` branch, which scaled the
network up to 8 stride-only validators to mirror mainnet shape. Here
the goal is the opposite: the smallest possible delta from the everyday
integration-tests network so that the existing `core.test.ts` suite
keeps working post-upgrade.

**This is a throwaway branch** (`poa-migration-ig-tests`), not for
merge — same posture as the rehearsal. Test-only constants are
hard-coded into the v33 source. Each modified file carries a top-of-file
`// REHEARSAL ONLY — DO NOT MERGE` marker.

**Companion specs:**
- `2026-04-27-ics-to-poa-migration-design.md` — the v33 migration itself.
- `2026-04-30-poa-migration-rehearsal-design.md` — the 8-validator rehearsal
  this design is patterned after.

**Pre-conditions:**
- v33 implementation is complete on this branch (handler, POA wiring,
  distrwrapper, ante handler change, store upgrades).
- `v32.0.0` is tagged on the upstream Stride repo.
- Existing integration-tests network can spin up `make start` cleanly.

**Success criteria:**
1. v32 network spins up with 3 stride validators, IBC channels open to
   `cosmoshub` and `osmosis`, host zones registered.
2. v33 upgrade proposal passes; cosmovisor swaps binaries cleanly; block
   production continues with the same validator hash across the upgrade
   height.
3. Tx fees submitted post-upgrade route to the POA module account, not
   `fee_collector`.
4. `make test-core` passes against both `cosmoshub` and `osmosis` host
   zones post-upgrade (IBC transfers, liquid staking, redemption,
   unbonding).

**Non-goals:**
- POA `MsgWithdrawFees` payout (covered by rehearsal).
- 5/3 govenator/PSS-only split (covered by rehearsal).
- LSM / auction / autopilot / burn test suites (run them later if useful;
  not in scope here).
- Production-grade hygiene on test fixtures.

## §2. Network shape

```
   ┌─────────────────────┐    ┌──────────────┐    ┌──────────────┐
   │ Stride (3 vals)     │◄──►│ Cosmos Hub   │    │ Osmosis      │
   │ - val1/val2/val3    │    │ (1 val)      │    │ (1 val)      │
   │ - cons keys pinned  │    │ unchanged    │    │ unchanged    │
   │ - all 3 govenators  │    └──────────────┘    └──────────────┘
   │ - v32 → v33 upgrade │           ▲                   ▲
   │                     │           │  relayers (rly + hermes)
   │                     │───────────┴───────────────────┘
   └─────────────────────┘
```

Same Helm chart, same `activeChains` (`stride`, `cosmoshub`, `osmosis`),
same relayer entries (rly + hermes for both stride↔hub and stride↔osmosis)
as the everyday integration-tests network. Nothing changes for hub or
osmosis; the migration handler only touches Stride state.

All 3 stride validators are also govenators (registered via
`tx staking create-validator` in `create-validator.sh`). No PSS-only
split — at scale 3 every validator participates, which is fine for
exercising the post-upgrade `distrwrapper` stake-iteration path on every
block.

## §3. Pinned cons-key generation

The v33 handler enforces strict joins from each consensus address (read
from `ccvconsumer`) to a moniker (in `app/upgrades/v33/validators.json`)
to a Stride operator bech32 (in `utils/poa.go::PoaValidatorSet`). Cons
keys generated fresh by `binaryd init` would not match the mainnet hex
addresses baked into `validators.json`, so we pin them.

**One-time, offline:**

1. Run `strided init tmp-N --chain-id ig-tests --home /tmp/cons-N` for
   N=0,1,2.
2. For each, capture `${home}/config/priv_validator_key.json` verbatim
   and the lowercase hex of `.address`.
3. Stash the 3 blobs in `integration-tests/network/configs/keys.json`
   under a new top-level `cons_keys: [...]` array, indexed 0–2 to match
   pod indices (`stride-validator-0` → `cons_keys[0]`, etc.).
4. Bake the 3 hex addresses into `app/upgrades/v33/validators.json`,
   keyed `Validator1`/`Validator2`/`Validator3`.
5. For each of val1/val2/val3, derive the operator bech32 by recovering
   their mnemonic from `keys.json` into a fresh keyring and running
   `strided keys show -a`. Bake the 3 results into
   `utils/poa.go::PoaValidatorSet`.

A small generator script (`integration-tests/scripts/generate_ig_test_state.sh`,
mirroring the rehearsal's `generate_rehearsal_state.sh` but for 3 pods)
emits all three artifacts as a single JSON document on stdout for the
implementer to copy into the right files.

## §4. Source-tree edits on this branch

All edits live on `poa-migration-ig-tests` and are explicitly never
merged. Each modified source file carries a top-of-file marker:

```
// REHEARSAL ONLY — DO NOT MERGE
```

(For `validators.json`: JSON has no comments — drop a sibling
`validators.json.REHEARSAL_ONLY` marker file instead.)

**`app/upgrades/v33/constants.go`:**

```go
const AdminMultisigAddress = "stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8"
const ExpectedValidatorCount = 3
```

The `AdminMultisigAddress` value is the existing `admin` account from
`integration-tests/network/configs/keys.json`. The `len(m) != ExpectedValidatorCount`
guard in the `validatorsJSON` init block keeps working — it now requires
exactly 3 entries in `validators.json`.

**`app/upgrades/v33/validators.json`:**

```json
{
  "<hex_addr_pod0>": "Validator1",
  "<hex_addr_pod1>": "Validator2",
  "<hex_addr_pod2>": "Validator3"
}
```

The three hex addresses come from §3 step 2.

**`utils/poa.go::PoaValidatorSet`:**

```go
var PoaValidatorSet = []PoaValidator{
    {Moniker: "Validator1", Operator: "stride1<val1-acct>", HubAddress: ""},
    {Moniker: "Validator2", Operator: "stride1<val2-acct>", HubAddress: ""},
    {Moniker: "Validator3", Operator: "stride1<val3-acct>", HubAddress: ""},
}
```

`Operator` is the `sdk.AccAddress` bech32 (`stride1...`) derived from
each validator's mnemonic in `keys.json` — *not* a `stridevaloper1...`
address. `HubAddress` is unused on this code path; leave it empty.

These three edits make the v33 binary accept the 3-validator test set
as its migration input. They are simultaneously the reason this branch
must never be merged: shipping any of them to mainnet would point the
upgrade handler at fictional validators with a test admin.

## §5. Integration-tests edits

**`integration-tests/network/configs/keys.json`:**
- Add a top-level `cons_keys: [...]` array of 3 `priv_validator_key.json`
  blobs (the verbatim JSON contents from §3 step 2).

**`integration-tests/network/scripts/init-chain.sh`:**
- In `add_validators` (loop `i=1..NUM_VALIDATORS`), after the per-pod
  `binaryd init` produces `${validator_home}/config/priv_validator_key.json`,
  overwrite that file with `cons_keys[i-1]` from `keys.json`, **before**
  the `tendermint show-node-id` / pubkey extraction calls. This ensures
  the comma-separated `validator_public_keys` passed to
  `add-consumer-section` reflects the pinned keys, and the `ccvconsumer`
  state at upgrade time matches `validators.json`.

**`integration-tests/network/scripts/init-node.sh`:**
- After the existing `download_shared_file ... priv_validator_key.json`,
  additionally overwrite the file with `cons_keys[VALIDATOR_INDEX-1]`
  from `keys.json`. The shared file already comes from a main node that
  was pinned in `init-chain.sh`; this is belt-and-suspenders so the
  node-init script is self-consistent.

**`integration-tests/network/scripts/upgrade.sh`:** no change. Already
votes from val1/val2/val3.

**`integration-tests/network/scripts/create-validator.sh`:** no change.
All 3 pods register as govenators.

**`integration-tests/network/values.yaml`:** no change. `numValidators: 3`
stays; cosmoshub, osmosis, and the 4 relayer entries stay.

## §6. Build and execution flow

**Build the upgrade image (once per code change to v32 or v33):**

```bash
cd integration-tests
UPGRADE_OLD_VERSION=v32.0.0 make build-stride-upgrade
```

This produces a single image with v32 (from upstream tag) at
`cosmovisor/genesis/bin/strided` and v33 (from this branch) at
`cosmovisor/upgrades/v33/bin/strided`. The existing `build.sh` already
seds the keys.json admin into `utils/admins.go` for the old binary; that
still works.

**Spin up the network:**

```bash
make start
```

`helm install` deploys 3 stride pods + 1 cosmoshub pod + 1 osmosis pod +
relayers. Stride pods run `init-chain.sh` (pod 0 only) / `init-node.sh`
(all pods), then start cosmovisor with the v32 binary. All 3 pods run
`create-validator.sh` post-startup.

**Verify pre-upgrade state:**
- All 3 stride pods producing blocks (`strided status`).
- ICA channels and host zones registered for both `cosmoshub` and
  `osmosis` (the existing `core.test.ts` setup helpers can do this — or
  smoke-test by hand: `strided q stakeibc list-host-zone` shows 2 entries).
- `strided q ccvconsumer all-cc-validators` returns 3 entries with hex
  addresses matching `validators.json`.

**Trigger the upgrade:**

```bash
make upgrade-stride
```

Submits the v33 software-upgrade proposal at `latest_height + 45`,
votes from val1/val2/val3, polls until passed. Cosmovisor swaps binaries
at the upgrade height; pods restart with v33.

**Run the tests:**

```bash
make test-core
```

This runs `client/test/core.test.ts` via vitest against both `cosmoshub`
and `osmosis` host zones. Test suite covers IBC transfers, liquid
staking, deposit records, redemption rate, unbonding, and redemption
record removal — all post-upgrade.

## §7. Risks

### R1: Cons-key drift between keys.json and validators.json

If the implementer regenerates cons keys without also regenerating
`validators.json`, `make start` will boot but the upgrade will halt at
`helpers.go:39` (no moniker found for the consensus hex). Mitigation:
the generator script emits both artifacts in one pass; commit them
together.

### R2: `mainnet_export_test.go` fails on this branch

Tests in `app/upgrades/v33/mainnet_export_test.go` are keyed off the
real mainnet hex addresses, which are no longer in our test
`validators.json`. Acceptable — same posture as the rehearsal branch.
Don't merge.

### R3: Relayer blip across the upgrade height

Cosmovisor restarts the stride pods at the upgrade height. The existing
relayer setup reconnects automatically, but if a packet is in-flight at
that exact moment, an IBC test in `core.test.ts` may see a one-time
flake. If this happens, retry `make test-core`; a permanent fix is out
of scope.

### R4: Source edits get merged

Catastrophic if shipped to mainnet (admin would point at the test
account, validators at fictional cons addrs). Mitigations, in order of
defensive value:

- Each modified source file carries a top-of-file `// REHEARSAL ONLY — DO NOT MERGE`
  marker. For `validators.json`, a sibling `validators.json.REHEARSAL_ONLY`
  marker file alongside.
- Branch name (`poa-migration-ig-tests`) signals intent.
- Final pre-merge sanity check on any v33-related PR:
  `git diff main..poa-migration-ig-tests -- app/upgrades/v33/validators.json
  utils/poa.go app/upgrades/v33/constants.go`.

### R5: ICA / host-zone state survives the upgrade

`core.test.ts` post-upgrade depends on host zones being registered and
ICA channels open before the test starts. The migration handler doesn't
touch `stakeibc` or IBC channel state, so this should hold; verify
during implementation by re-querying host zones after the upgrade
height. If anything regresses here, the migration handler is the place
to look.

## §8. Out of scope

- POA `MsgWithdrawFees` payout — covered by rehearsal.
- 5/3 govenator/PSS-only split — covered by rehearsal.
- LSM / auction / autopilot / burn test suites — easy to add later by
  invoking the corresponding `make test-*` targets; not in scope here.
- Public testnet rehearsal.
- Multisig signing rehearsal.
- Performance benchmarking.

## §9. Summary

3-validator Stride network plus cosmoshub and osmosis on the standard
k8s integration-tests network. Pin 3 cons keys offline, bake matching
test addresses into a throwaway fork of `validators.json`,
`PoaValidatorSet`, and `AdminMultisigAddress`, change
`ExpectedValidatorCount` from 8 to 3. Build v32+v33 image, spin up,
propose-upgrade-vote, run `make test-core` against the post-upgrade
chain. Branch is throwaway.
