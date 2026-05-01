# Stride v33 — POA Migration Rehearsal on k8s Integration-Tests

## §1. Overview and objective

Run a one-time rehearsal of the v32 → v33 (ICS → POA) migration on the k8s
integration-tests network. The rehearsal validates that the migration handler
runs cleanly end-to-end under multi-node consensus and produces the expected
post-upgrade runtime behavior (block production, fee routing, govenator
rewards, POA payouts).

**This is a throwaway rehearsal.** The branch (`poa-migration-rehearsal`) is
explicitly not for merge. We intentionally hardcode test-only constants into
the v33 source to make the binary accept our test validator set, and we strip
the integration-tests setup down to the minimum needed for the rehearsal.

**Companion spec:** `2026-04-27-ics-to-poa-migration-design.md` (the v33
migration design itself). This rehearsal exercises that migration; it does
not modify it. Refer to the migration spec for handler internals and §6
(testing strategy) which deferred multi-node testing to "a separate, much
larger workstream" — this rehearsal is that workstream.

**Pre-conditions:**

- v33 implementation is complete on this branch (the migration handler,
  POA wiring, distrwrapper, ante handler change, store upgrades).
- v32.0.0 is tagged on the upstream Stride repo (it is).
- Dev k8s cluster has capacity for ~5 cores and ~6GB allocated to a single
  helm release.

**Success criteria:**

1. The v32 network spins up with 8 Stride validators producing blocks.
2. The v33 upgrade proposal passes and cosmovisor swaps binaries cleanly.
3. Block N+1 (post-upgrade) signs with the same 8 validators as block N.
   No interruption to block production.
4. Tx fees submitted post-upgrade land in the POA module account, not in
   `fee_collector`.
5. Govenator delegators continue to accrue rewards and can withdraw them.
6. POA validators can pull their share of accumulated fees via
   `MsgWithdrawFees` and the resulting balance landings match the
   expected pro-rata split.

**Non-goals:**

- Testing the LS revenue auction → strdburner pipeline (covered by unit
  tests; would require adding a host zone — not worth the setup cost).
- Testing gov proposals spanning the upgrade height (low marginal value
  over the design doc's Test 2 / Test 3).
- Testing slashing (POA doesn't slash).
- Testing the POA admin path (`MsgUpdateValidators`) — upstream POA code,
  not a Stride concern; admin presence is verified by querying params.
- Production-grade hygiene on test fixtures (the cons keys we generate are
  test-only by construction).

## §2. Network shape

```
                 ┌──────────────────────────┐
                 │  Stride (8 validators)   │
                 │  - chain-id: STRIDE      │
                 │  - bootstrap: ICS via    │
                 │    add-consumer-section  │
                 │  - 5 govenators (vals    │
                 │    0–4); vals 5–7 are    │
                 │    PSS-only              │
                 │  - v32 → v33 upgrade     │
                 └──────────────────────────┘
```

No Hub, no relayers, no host zones. Per the ICS provider investigation
(see brainstorm history): a fixed 8-validator opt-in PSS allowlist with no
rotations / slashes / opt-in events generates **zero** provider→consumer
IBC packets. The `add-consumer-section` genesis-injection shortcut produces
observably identical consumer keeper state to a real Hub provider for our
scenario. The migration handler reads consumer keeper state, not IBC
channel state, so the shortcut is sufficient.

The 5/3 govenator split mirrors the mainnet shape where ~3 of 8 PSS
validators have ICS-key-assigned consensus keys not in `x/staking`. This
exercises the `app/distrwrapper` stake-iteration path under multi-node
consensus — the exact case that exists to prevent a chain halt at every
BeginBlock post-migration.

## §3. Cons-key pinning and validator setup

The v33 handler enforces strict joins across two compile-time maps
(`app/upgrades/v33/validators.json` and `utils/poa.go::PoaValidatorSet`).
Any consensus address in `ccvconsumer` that lacks a moniker entry, or any
moniker that lacks an operator entry, halts the upgrade. Test validators
generated fresh by `binaryd init` would not match the mainnet hex addresses
baked into `validators.json`, so we pin the cons keys.

**Approach:**

1. Pre-generate 8 cons keys offline. Easiest mechanism: run `strided init
   tmp-N --chain-id rehearsal --home /tmp/cons-N` 8 times (N=0..7) and copy
   the resulting `/tmp/cons-N/config/priv_validator_key.json` for each.
   (CometBFT `gen-validator` would also work but the `init` path is
   guaranteed to produce the exact JSON shape strided expects.)
2. Stash the 8 blobs in `network/configs/keys.json` under a new `cons_keys`
   array (indexed 0–7, matching pod indices).
3. Modify `init-chain.sh` and `init-node.sh` to overwrite the
   auto-generated `priv_validator_key.json` with the pre-generated blob
   for the corresponding pod index, *before* any chain initialization that
   binds to the cons key.
4. Compute the hex address for each pinned cons key. These are the values
   we bake into the rehearsal-specific `validators.json`.

**Validator role split:**

- Pods 0–4 (val1–val5): both PSS validators (cons key in `ccvconsumer`)
  and govenators (registered via `tx staking create-validator`).
- Pods 5–7 (val6–val8): PSS validators only. `create-validator.sh` skips
  these via a `[[ "$POD_INDEX" -ge 5 ]] && exit 0` guard at the top.

## §4. Source-tree edits on this branch

All three edits live on `poa-migration-rehearsal` and are explicitly never
merged. Each modified source file gets a top-of-file comment:

```
// REHEARSAL ONLY — DO NOT MERGE
```

**`app/upgrades/v33/validators.json`** — replace the 8 mainnet entries
with the 8 hex addresses derived from the pinned cons keys. Monikers can
be `Validator1`–`Validator8`:

```json
{
  "<hex_addr_pod0>": "Validator1",
  "<hex_addr_pod1>": "Validator2",
  ...
  "<hex_addr_pod7>": "Validator8"
}
```

**`utils/poa.go::PoaValidatorSet`** — replace the 8 mainnet entries with
8 monikers → operator bech32s. Operator bech32s are derived from each
validator's mnemonic in `keys.json` (the `stride1...` account — *not* a
`stridevaloper1...` address; per the existing `PoaValidatorSet` shape,
`Operator` is an `sdk.AccAddress` bech32). `HubAddress` is unused in the
rehearsal path; leave it empty:

```go
var PoaValidatorSet = []PoaValidator{
    {Moniker: "Validator1", Operator: "stride1...", HubAddress: ""},
    ...
    {Moniker: "Validator8", Operator: "stride1...", HubAddress: ""},
}
```

**`app/upgrades/v33/constants.go::AdminMultisigAddress`** — replace
the mainnet multisig address with the existing `admin` account from
`keys.json`:

```go
const AdminMultisigAddress = "stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8"
```

These three edits make the v33 binary accept the test validator set as
its migration input. They are simultaneously the reason this branch must
never be merged: shipping any of them to mainnet would point the upgrade
handler at fictional validators with a test admin.

## §5. Integration-tests edits

**`integration-tests/network/values.yaml`:**

- Drop `cosmoshub` and `osmosis` from `activeChains`.
- Drop both relayer entries (`stride-cosmoshub` and `stride-osmosis`,
  both relayer and hermes types).
- Bump `chainConfigs.stride.numValidators: 3 → 8`.
- Optionally trim `chainConfigs` to remove unused `cosmoshub` and
  `osmosis` blocks.

**`integration-tests/network/configs/keys.json`:**

- Add 3 more validator entries (`val6`, `val7`, `val8`) with fresh
  mnemonics generated via `strided keys add --recover` flow or any
  bip39 generator. Currently the file has val1–val5.
- Add a top-level `cons_keys` array of 8 entries, each containing the
  full `priv_validator_key.json` blob produced by
  `strided tendermint gen-validator`.

**`integration-tests/network/scripts/init-chain.sh`:**

- After `binaryd init` runs (which generates `priv_validator_key.json`
  in `${validator_home}/config/`), overwrite that file with the cons
  key from `keys.json` for the matching pod index. The function loops
  `i = 1..NUM_VALIDATORS`, so the cons-key index is `i-1`.
- The validator pubkeys collected for `add-consumer-section` should
  now reflect the pinned cons keys (since they're read from the
  overwritten `priv_validator_key.json`).

**`integration-tests/network/scripts/init-node.sh`:**

- After `download_shared_file ... priv_validator_key.json`, additionally
  overwrite with the pinned cons key for `VALIDATOR_INDEX`. (The shared
  file already comes from the main node which had its key pinned; this
  is belt-and-suspenders to make the script self-contained if the
  pinning logic ever runs only on the main node.)

**`integration-tests/network/scripts/create-validator.sh`:**

- At the top of `main()` (after `wait_for_startup`), guard:
  ```bash
  POD_INDEX=${HOSTNAME##*-}
  if [[ "$POD_INDEX" -ge 5 ]]; then
      echo "Skipping govenator registration for PSS-only validator (pod $POD_INDEX)"
      exit 0
  fi
  ```
- Pods 5–7 thus boot with cons keys in `ccvconsumer` but no x/staking
  record.

**`integration-tests/network/scripts/upgrade.sh`:**

- Currently votes from val1–val3 hardcoded. Extend to vote from
  val1–val5. Vals 6–8 are not govenators and have no STRD to vote
  with.

## §6. Build and execution flow

**Build the upgrade image:**

```bash
cd integration-tests
UPGRADE_OLD_VERSION=v32.0.0 make build-stride-upgrade
```

This builds a single image containing both binaries (v32 from upstream
tag, v33 from this branch) and configures cosmovisor to swap at the
upgrade height. Per `Dockerfile.stride-upgrade`, the v32 binary lives at
`cosmovisor/genesis/bin/strided` and the v33 binary at
`cosmovisor/upgrades/v33/bin/strided`.

**Spin up the network:**

```bash
make start
```

`helm install` deploys 8 stride validator pods. Each runs `init-chain.sh`
(pod 0 only) and `init-node.sh` in `initContainer`, then starts cosmovisor
with the v32 binary. Pods 0–4 run `create-validator.sh` as `postStart`;
pods 5–7 skip it.

**Verify pre-upgrade state** (before triggering the upgrade):

- `kubectl exec stride-validator-0 -- strided q ccvconsumer all-cc-validators`
  returns 8 entries.
- `kubectl exec stride-validator-0 -- strided q staking validators`
  returns 5 entries.
- All 8 pods are signing blocks (check `strided status` from each).
- Snapshot: validator hash, govenator delegation totals, govenator
  operator account balances.

**Trigger the upgrade:**

```bash
make upgrade-stride
```

Submits the v33 software-upgrade proposal at `latest_height + 45`,
votes from val1–val5, polls until passed. Cosmovisor swaps binaries at
the upgrade height; pods restart with v33.

## §7. Test sequence

All four tests run after the upgrade has executed (cosmovisor swap
complete, network is back to producing blocks on v33).

### Test 1 — block production continues

The load-bearing assertion. The validator hash post-upgrade must equal
the validator hash pre-upgrade.

- Read the validator hash at the upgrade height from any validator's
  block header (`strided q block --type=height <upgrade_height>` or via
  CometBFT RPC).
- Read the validator hash at upgrade_height + 5.
- Assert equality.
- Verify all 8 pods have advanced past upgrade_height + 5 in
  `strided status`.

If this fails, the chain has halted or the validator set has shifted.
Recovery is manual; rehearsal halts.

### Test 2 — tx fees route to POA

- Pre-tx: snapshot bank balance on `fee_collector` and on the POA
  module account address (`strided q auth module-account poa`).
- Submit a tx (any tx — e.g., a bank send between two of the funded
  test accounts) with non-zero fees.
- Wait one block.
- Post-tx: re-query both balances. The POA account balance should have
  increased by the tx fee amount; `fee_collector` should be unchanged
  (or net of any inflation that landed and was then drained by
  `distrwrapper` in the same block — verify the delta is consistent).

This confirms the ante handler change took effect.

### Test 3 — govenator rewards accrue and are withdrawable

- Pick one of the val1–val5 operator accounts.
- Pre-baseline: `strided q distribution rewards <delegator>` to capture
  the current reward total.
- Wait several blocks (10–20).
- Post-baseline: re-query. Total should have strictly increased.
- Submit `tx distribution withdraw-rewards <validator> --from <delegator>`.
- Verify the delegator's bank balance increased by the withdrawn amount
  (minus tx fees).

This confirms the `distrwrapper` 85% slice is being allocated correctly
through the standard `x/distribution` machinery.

### Test 4 — POA fee distribution via MsgWithdrawFees

- Wait for fees to accumulate in the POA module account (test 2's tx +
  the 15% from `distrwrapper` over several blocks).
- Pick one validator (e.g., val1) and snapshot their bank balance.
- Submit `tx poa withdraw-fees --from val1` (exact CLI subcommand TBD
  during implementation; check the POA module's CLI).
- Verify val1's bank balance increased by approximately 1/8 of the POA
  module account balance pre-withdraw.
- Optionally repeat for val6 (a non-govenator) to confirm POA pays out
  to validators that have no x/staking record — the load-bearing case
  for the 5/3 split.

This confirms POA's lazy checkpoint accounting and pays out the actual
distributable amount.

## §8. Risks specific to the rehearsal

### R1: 8 pods exceed dev cluster capacity

At current pod limits (600m CPU / 700M memory each), 8 stride pods need
4.8 cores / 5.6GB. Should fit; if not, drop per-pod resources to ~400m /
500M, or scale to 6 validators (still maintaining a 4/2 split mirror).

### R2: Pinned cons keys leak

The 8 generated cons keys exist only in `network/configs/keys.json` on
this branch and produce no real-network identity. Hygiene: never copy
this file into another branch or deployment context.

### R3: Source-tree edits get merged

The three edits in §4 are catastrophic if shipped to mainnet (admin =
test account, validators = test cons addrs). Mitigations, in order of
defensive value:

- Each modified source file carries a `// REHEARSAL ONLY — DO NOT MERGE`
  header comment (or `// REHEARSAL ONLY — DO NOT MERGE` JSON sidecar
  comment for `validators.json` since JSON has no comments — use a
  separate `validators.json.REHEARSAL_ONLY` marker file alongside).
- The branch name (`poa-migration-rehearsal`) signals intent.
- Reviewer of any v33-related PR should diff against this branch's
  modified files as a final sanity check before approving:
  `git diff main..poa-migration-rehearsal -- app/upgrades/v33/validators.json
  utils/poa.go app/upgrades/v33/constants.go`.

### R4: Pre-upgrade `add-consumer-section` may need adjustment for 8 vals

The current loop in `init-chain.sh::add_validators` iterates
`NUM_VALIDATORS` and concatenates pubkeys. With `NUM_VALIDATORS=8` this
should work as-is (the binary accepts a comma-separated list of any
length). Verify during implementation; trivial fix if not.

### R5: Vals 6–8 have no genesis account funding

`init-chain.sh::add_genesis_account` is called for every validator
based on `keys.json::validators[*]`, regardless of whether they later
become a govenator. Pods 5–7 do receive a genesis balance (so their
operator addresses can later pay tx fees if needed for test 4) — this
is correct, the guard in `create-validator.sh` only skips the staking
registration, not the account funding.

## §9. Out of scope

- LS revenue flow testing (no host zone in the rehearsal network).
- Public testnet rehearsal (no public testnet).
- Multisig signing rehearsal (operational concern, not v33 correctness).
- POA admin path (`MsgUpdateValidators`) — upstream POA code, admin
  presence verified by query.
- Slashing (POA doesn't slash).
- Performance benchmarking (small dev cluster, not representative).
- Recovery rehearsal (if upgrade fails at height N, recovery is manual
  on the test network — same as mainnet would be, but not what we're
  testing here).

## §10. Summary

8-validator Stride network on k8s, no Hub, 5/3 govenator/PSS-only split
to mirror mainnet shape. Pin cons keys offline, bake matching test
addresses into a rehearsal-only fork of `validators.json`,
`PoaValidatorSet`, and `AdminMultisigAddress`. Build v32+v33 image,
spin up, propose-upgrade-vote, run 4 post-upgrade checks: block
continuity, tx fee routing, govenator reward accrual + withdrawal,
POA `MsgWithdrawFees` payout. Branch is throwaway.
