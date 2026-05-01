# POA Migration on Standard Integration-Tests Network — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Run the v32 → v33 (ICS → POA) upgrade against the standard k8s integration-tests network (3 stride validators + cosmoshub + osmosis) and run `make test-core` post-upgrade.

**Architecture:** Pin 3 stride consensus keys offline, bake the matching hex addresses + operator bech32s + admin into a throwaway fork of `validators.json`, `PoaValidatorSet`, and `AdminMultisigAddress`. Change `ExpectedValidatorCount` from 8 to 3. Existing hub/osmosis/relayer setup is untouched. Branch is throwaway (not for merge), same posture as `poa-migration-rehearsal`.

**Tech Stack:** Bash scripts, Helm, kubectl, Cosmos SDK CLI (`strided`), Go (test-only edits to v33 source).

**Companion spec:** `docs/superpowers/specs/2026-05-01-poa-migration-ig-tests-design.md`

**Branch:** `poa-migration-ig-tests` — this work is not for merge.

**Expected test breakage on this branch:** `app/upgrades/v33/mainnet_export_test.go` will fail because mainnet hex cons addresses are no longer in our test `validators.json`. This is acceptable — branch is throwaway. The handler's own unit tests (`app/upgrades/v33/helpers_test.go`, `app/upgrades/v33/upgrades_test.go`) install their own monikers/operators and should keep passing.

---

## File map

**Created:**
- `integration-tests/scripts/generate_ig_test_state.sh` — one-time helper that emits cons keys, hex addresses, and operator bech32s as a single JSON document.
- `app/upgrades/v33/validators.json.REHEARSAL_ONLY` — sidecar marker file (JSON has no comments).

**Modified (test-only, throwaway):**
- `integration-tests/network/configs/keys.json` — add a top-level `cons_keys` array of 3 `priv_validator_key.json` blobs.
- `integration-tests/network/scripts/init-chain.sh` — overwrite the auto-generated cons key with the pinned blob, gated on `CHAIN_NAME == "stride"`.
- `integration-tests/network/scripts/init-node.sh` — same overwrite for non-main stride pods.
- `app/upgrades/v33/validators.json` — replace 8 mainnet entries with 3 test entries.
- `app/upgrades/v33/constants.go` — `ExpectedValidatorCount: 8 → 3`, swap admin to keys.json admin, add `// REHEARSAL ONLY — DO NOT MERGE` header.
- `utils/poa.go` — replace `PoaValidatorSet` with 3 test entries, add header marker.

---

## Phase 1: Cons-key generator and state

### Task 1: Write the generator script

**Files:**
- Create: `integration-tests/scripts/generate_ig_test_state.sh`

This script produces all the deterministic inputs we need to bake into both the integration-tests config AND the v33 source: 3 `priv_validator_key.json` blobs, their lowercase-hex consensus addresses, and 3 operator bech32 addresses derived from val1/val2/val3 mnemonics.

Mirrors `integration-tests/scripts/generate_rehearsal_state.sh` from the `poa-migration-rehearsal` branch, scaled down to 3 vals and reusing existing mnemonics (no need to generate new ones).

- [ ] **Step 1: Create the script**

```bash
#!/bin/bash
# generate_ig_test_state.sh
# One-time helper for the POA migration integration-tests run.
# Produces: 3 priv_validator_key.json blobs (cons keys), their lowercase
# hex consensus addresses, and operator bech32 addresses derived from
# val1/val2/val3 mnemonics. Outputs a single JSON document on stdout
# for the implementer to inject into keys.json, validators.json, and
# PoaValidatorSet.
#
# Requires: strided binary on PATH (any recent version that knows
# `strided init` and `strided keys`).

set -eu

WORKDIR=$(mktemp -d)
trap "rm -rf $WORKDIR" EXIT

KEYS_FILE="${KEYS_FILE:-integration-tests/network/configs/keys.json}"
NUM_VALIDATORS=3

# Sanity-check that keys.json has at least 3 validator entries.
existing_count=$(jq '.validators | length' "$KEYS_FILE")
if [[ "$existing_count" -lt "$NUM_VALIDATORS" ]]; then
    echo "ERROR: keys.json has only $existing_count validators; need $NUM_VALIDATORS." >&2
    exit 1
fi

# Generate 3 cons keys using `strided init`. Write priv_validator_key.json
# verbatim — same shape strided expects on disk.
declare -a CONS_KEYS_JSON
declare -a HEX_ADDRS

for ((i=0; i<NUM_VALIDATORS; i++)); do
    home="${WORKDIR}/cons-${i}"
    strided init "tmp-${i}" --chain-id ig-tests --home "$home" >/dev/null 2>&1
    cons_key=$(cat "${home}/config/priv_validator_key.json")
    hex_addr=$(jq -r '.address' <<<"$cons_key" | tr 'A-Z' 'a-z')

    CONS_KEYS_JSON[$i]="$cons_key"
    HEX_ADDRS[$i]="$hex_addr"
done

# Derive operator bech32 from each validator mnemonic (val1, val2, val3).
declare -a OPERATORS
for ((i=0; i<NUM_VALIDATORS; i++)); do
    mnemonic=$(jq -r ".validators[$i].mnemonic" "$KEYS_FILE")
    name="ig-tests-val-$i"
    keyring="${WORKDIR}/keyring-${i}"
    echo "$mnemonic" | strided keys add "$name" \
        --recover --keyring-backend test --home "$keyring" >/dev/null 2>&1
    operator=$(strided keys show "$name" -a \
        --keyring-backend test --home "$keyring")
    OPERATORS[$i]="$operator"
done

# Emit a single JSON document with the cons keys, hex addrs, and operators.
jq -n \
  --argjson cons_keys "$(printf '%s\n' "${CONS_KEYS_JSON[@]}" | jq -s .)" \
  --argjson hex_addrs "$(printf '%s\n' "${HEX_ADDRS[@]}" | jq -R . | jq -s .)" \
  --argjson operators "$(printf '%s\n' "${OPERATORS[@]}" | jq -R . | jq -s .)" \
  '{
     cons_keys: $cons_keys,
     hex_addrs: $hex_addrs,
     operators: $operators
   }'
```

- [ ] **Step 2: Make executable**

```bash
chmod +x integration-tests/scripts/generate_ig_test_state.sh
```

- [ ] **Step 3: Verify shell syntax**

```bash
bash -n integration-tests/scripts/generate_ig_test_state.sh
```

Expected: no output.

- [ ] **Step 4: Commit**

```bash
git add integration-tests/scripts/generate_ig_test_state.sh
git commit -m "feat(ig-tests): add cons-key/operator generator script for 3-val POA test"
```

---

### Task 2: Run generator and inject `cons_keys` into keys.json

**Files:**
- Modify: `integration-tests/network/configs/keys.json` (add `cons_keys` array)

- [ ] **Step 1: Run the generator and save output**

```bash
cd /Users/sampocs/Documents/Projects/stride
bash integration-tests/scripts/generate_ig_test_state.sh > /tmp/ig-test-state.json
```

Expected: exits 0; `/tmp/ig-test-state.json` is valid JSON with three top-level keys (`cons_keys`, `hex_addrs`, `operators`), each an array of length 3.

- [ ] **Step 2: Verify output shape**

```bash
jq '.cons_keys | length, .hex_addrs | length, .operators | length' /tmp/ig-test-state.json
```

Expected: three lines, each `3`.

- [ ] **Step 3: Inject `cons_keys` into keys.json**

```bash
jq --slurpfile s /tmp/ig-test-state.json \
   '. + {cons_keys: $s[0].cons_keys}' \
   integration-tests/network/configs/keys.json \
   > /tmp/keys.json && \
mv /tmp/keys.json integration-tests/network/configs/keys.json
```

- [ ] **Step 4: Verify keys.json has cons_keys array**

```bash
jq '.cons_keys | length' integration-tests/network/configs/keys.json
```

Expected: `3`.

- [ ] **Step 5: Keep `/tmp/ig-test-state.json` around**

Tasks 5 and 7 will read `hex_addrs` and `operators` from this file. Do not delete until the v33 source edits are in.

- [ ] **Step 6: Commit**

```bash
git add integration-tests/network/configs/keys.json
git commit -m "feat(ig-tests): pin 3 cons keys in keys.json"
```

---

## Phase 2: Integration-tests script edits

### Task 3: Pin cons keys in init-chain.sh (stride only)

**Files:**
- Modify: `integration-tests/network/scripts/init-chain.sh`

The `add_validators` function loops over `NUM_VALIDATORS` for whichever chain is running. We need to overwrite the auto-generated `priv_validator_key.json` with the pinned blob from `keys.json::cons_keys[i-1]`, but ONLY for the stride chain — hub and osmosis must keep their auto-generated keys.

- [ ] **Step 1: Edit `add_validators()` to overwrite cons key for stride only**

Open `integration-tests/network/scripts/init-chain.sh`. Find the `for (( i=1; i <= $NUM_VALIDATORS; i++ ))` loop in `add_validators()` (around line 66). After the `if [[ "$i" == "1" ]]; then validator_home=${CHAIN_HOME} else ... fi` block (around line 79) and BEFORE the `# Save the node IDs and keys to the API` comment (around line 81), insert:

```bash
        # ig-tests POA upgrade: overwrite the auto-generated stride cons
        # key with the pinned blob from keys.json::cons_keys[i-1]. This
        # makes the consensus hex addresses match the values baked into
        # app/upgrades/v33/validators.json. Hub/osmosis keep their
        # auto-generated keys.
        if [[ "$CHAIN_NAME" == "stride" ]]; then
            cons_key_index=$((i - 1))
            jq -r ".cons_keys[$cons_key_index]" ${KEYS_FILE} \
                > ${validator_home}/config/priv_validator_key.json
        fi
```

The result should look like:

```bash
        # Use a separate directory for the non-main nodes so we can generate unique validator keys
        if [[ "$i" == "1" ]]; then 
            validator_home=${CHAIN_HOME}
        else 
            validator_home=/tmp/${CHAIN_NAME}-${name} && rm -rf $validator_home
            $BINARY init $name --chain-id $CHAIN_ID --overwrite --home ${validator_home} &> /dev/null
        fi

        # ig-tests POA upgrade: overwrite the auto-generated stride cons
        # key with the pinned blob from keys.json::cons_keys[i-1]. ...
        if [[ "$CHAIN_NAME" == "stride" ]]; then
            cons_key_index=$((i - 1))
            jq -r ".cons_keys[$cons_key_index]" ${KEYS_FILE} \
                > ${validator_home}/config/priv_validator_key.json
        fi

        # Save the node IDs and keys to the API
        $BINARY tendermint show-node-id --home ${validator_home} > node_id.txt
```

The position is critical — the overwrite must happen *before* `$BINARY tendermint show-node-id` and `validator_public_keys+=...` because both read from `priv_validator_key.json`.

- [ ] **Step 2: Verify shell syntax**

```bash
bash -n integration-tests/network/scripts/init-chain.sh
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add integration-tests/network/scripts/init-chain.sh
git commit -m "feat(ig-tests): pin stride cons keys in init-chain.sh"
```

---

### Task 4: Pin cons keys in init-node.sh (stride only)

**Files:**
- Modify: `integration-tests/network/scripts/init-node.sh`

Non-main pods download `priv_validator_key.json` from the API. The shared file already came from a main-node init-chain that was pinned, so it's already correct — but we re-pin here for self-consistency, gated on stride.

- [ ] **Step 1: Edit `update_config()` to overwrite cons key for stride only**

Open `integration-tests/network/scripts/init-node.sh`. Find the existing line (around line 60):

```bash
    download_shared_file ${VALIDATOR_KEYS_DIR}/${CHAIN_NAME}/val${VALIDATOR_INDEX}.json ${CHAIN_HOME}/config/priv_validator_key.json 
```

Right after that line, insert:

```bash

    # ig-tests POA upgrade: re-pin the stride cons key from
    # keys.json::cons_keys[VALIDATOR_INDEX-1] for self-consistency.
    # The shared file already came from the pinned main node, but
    # re-pinning here makes init-node.sh independent of init-chain.sh
    # ordering.
    if [[ "$CHAIN_NAME" == "stride" ]]; then
        cons_key_index=$((VALIDATOR_INDEX - 1))
        jq -r ".cons_keys[$cons_key_index]" ${KEYS_FILE} \
            > ${CHAIN_HOME}/config/priv_validator_key.json
    fi
```

- [ ] **Step 2: Verify shell syntax**

```bash
bash -n integration-tests/network/scripts/init-node.sh
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add integration-tests/network/scripts/init-node.sh
git commit -m "feat(ig-tests): pin stride cons keys in init-node.sh"
```

---

## Phase 3: v33 source edits (REHEARSAL ONLY — DO NOT MERGE)

### Task 5: Replace v33/validators.json and add sidecar marker

**Files:**
- Modify: `app/upgrades/v33/validators.json`
- Create: `app/upgrades/v33/validators.json.REHEARSAL_ONLY`

- [ ] **Step 1: Capture the 3 hex cons addresses from the generator output**

```bash
jq -r '.hex_addrs[]' /tmp/ig-test-state.json
```

Expected: 3 lines of 40-char lowercase hex, one per line. Capture these as `<hex_addr_pod0>`, `<hex_addr_pod1>`, `<hex_addr_pod2>` for use in step 2.

- [ ] **Step 2: Replace `app/upgrades/v33/validators.json`**

Use the Edit tool (or your editor of choice) to replace the file's contents with exactly:

```json
{
  "<hex_addr_pod0>": "Validator1",
  "<hex_addr_pod1>": "Validator2",
  "<hex_addr_pod2>": "Validator3"
}
```

Substitute each `<hex_addr_pod*>` with the actual 40-char hex from step 1.

- [ ] **Step 3: Create the sidecar marker file**

Create `app/upgrades/v33/validators.json.REHEARSAL_ONLY` with this content:

```
REHEARSAL ONLY — DO NOT MERGE

This branch (poa-migration-ig-tests) replaces the 8 mainnet validators in
validators.json with 3 test validators that match cons keys pinned in
integration-tests/network/configs/keys.json::cons_keys. Shipping this
file to mainnet would point the v33 upgrade handler at fictional
validators and halt the upgrade.

Companion files modified on this branch:
  - app/upgrades/v33/constants.go (ExpectedValidatorCount, AdminMultisigAddress)
  - utils/poa.go (PoaValidatorSet)

Pre-merge sanity check on any v33-related PR:
  git diff main..poa-migration-ig-tests -- \
    app/upgrades/v33/validators.json \
    app/upgrades/v33/constants.go \
    utils/poa.go
```

- [ ] **Step 4: Verify validators.json parses and has 3 entries**

```bash
jq 'length' app/upgrades/v33/validators.json
```

Expected: `3`.

- [ ] **Step 5: Commit**

```bash
git add app/upgrades/v33/validators.json app/upgrades/v33/validators.json.REHEARSAL_ONLY
git commit -m "test(ig-tests): swap validators.json to 3 test cons addresses"
```

---

### Task 6: Update v33 constants.go (count + admin + marker)

**Files:**
- Modify: `app/upgrades/v33/constants.go`

- [ ] **Step 1: Replace the file contents**

Replace `app/upgrades/v33/constants.go` with:

```go
// REHEARSAL ONLY — DO NOT MERGE
// This branch (poa-migration-ig-tests) replaces ExpectedValidatorCount and
// AdminMultisigAddress with test-only values. Shipping these to mainnet
// would point the v33 upgrade handler at the wrong admin and the wrong
// validator count. See app/upgrades/v33/validators.json.REHEARSAL_ONLY.

package v33

import (
	_ "embed"
	"encoding/json"
)

// UpgradeName is the SDK upgrade plan name. Match the binary release tag.
const UpgradeName = "v33"

// AdminMultisigAddress is the bech32 address that POA recognizes as its admin
// post-upgrade
const AdminMultisigAddress = "stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8"

// ExpectedValidatorCount is enforced by the upgrade handler. Panics if
// consumerKeeper.GetAllCCValidator returns a different count.
const ExpectedValidatorCount = 3

//go:embed validators.json
var validatorsJSON []byte

// ValidatorMonikers maps lowercase hex of the consensus address (the raw 20-byte
// address returned by ccv consumer GetAllCCValidator) to the validator's moniker.
// ICS does not store monikers on the consumer chain — they live on the Hub.
// We pre-bake them so they appear correctly in POA's validator records.
//
// Populated from validators.json at init. Regenerate with scripts/fetch_validator_monikers.sh.
var ValidatorMonikers = func() map[string]string {
	m := make(map[string]string)
	if err := json.Unmarshal(validatorsJSON, &m); err != nil {
		panic("v33: failed to parse embedded validators.json: " + err.Error())
	}
	if len(m) != ExpectedValidatorCount {
		panic("v33: validators.json has wrong validator count")
	}
	return m
}()
```

The two value changes vs. main:
- `AdminMultisigAddress`: `"stride1fduug6m38gyuqt3wcgc2kcgr9nnte0n7ssn27e"` → `"stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8"` (the keys.json admin).
- `ExpectedValidatorCount`: `8` → `3`.

- [ ] **Step 2: Verify the file compiles**

```bash
go build ./app/upgrades/v33/...
```

Expected: no output (clean build). The `init`-time check on `len(m) != ExpectedValidatorCount` reads `validators.json` at compile time via `//go:embed`, so this will fail if Task 5's `validators.json` doesn't have exactly 3 entries.

- [ ] **Step 3: Commit**

```bash
git add app/upgrades/v33/constants.go
git commit -m "test(ig-tests): point v33 admin and count at the 3-val test setup"
```

---

### Task 7: Update PoaValidatorSet in utils/poa.go

**Files:**
- Modify: `utils/poa.go`

- [ ] **Step 1: Capture the 3 operator bech32s from the generator output**

```bash
jq -r '.operators[]' /tmp/ig-test-state.json
```

Expected: 3 lines of `stride1...` bech32. Capture as `<operator_val1>`, `<operator_val2>`, `<operator_val3>`.

- [ ] **Step 2: Replace the `PoaValidatorSet` block**

Open `utils/poa.go`. The current contents have an 8-entry `PoaValidatorSet` declared with `var PoaValidatorSet = []PoaValidator{...}`. Replace just that variable declaration block with:

```go
// REHEARSAL ONLY — DO NOT MERGE
// This branch (poa-migration-ig-tests) replaces the 8 mainnet entries with
// 3 test validators matching app/upgrades/v33/validators.json. See that
// file's sidecar REHEARSAL_ONLY marker for context.
var PoaValidatorSet = []PoaValidator{
	{Moniker: "Validator1", Operator: "<operator_val1>", HubAddress: ""},
	{Moniker: "Validator2", Operator: "<operator_val2>", HubAddress: ""},
	{Moniker: "Validator3", Operator: "<operator_val3>", HubAddress: ""},
}
```

Substitute each `<operator_val*>` with the actual `stride1...` bech32 from step 1.

Leave the `PoaValidator` struct definition untouched — only the `PoaValidatorSet` slice changes.

- [ ] **Step 3: Verify the file compiles**

```bash
go build ./...
```

Expected: clean build.

- [ ] **Step 4: Run the v33 handler unit tests**

```bash
go test ./app/upgrades/v33/...
```

Expected: `helpers_test.go` and `upgrades_test.go` pass; `mainnet_export_test.go` FAILS with errors about hex addresses not present in `ValidatorMonikers`. The mainnet-export failure is expected and acceptable on this throwaway branch.

If `helpers_test.go` or `upgrades_test.go` fails, something is wrong with the edits — those tests install their own monikers and operators and shouldn't be affected by the count/value swaps. Stop and investigate before committing.

- [ ] **Step 5: Commit**

```bash
git add utils/poa.go
git commit -m "test(ig-tests): swap PoaValidatorSet to 3 test entries"
```

---

## Phase 4: Build, run, test

### Task 8: Build the upgrade image

**Files:** none modified — this is an image build.

- [ ] **Step 1: Verify v32.0.0 tag exists locally**

```bash
git tag -l 'v32.0.0'
```

Expected: `v32.0.0`. If empty, fetch tags: `git fetch --tags`.

- [ ] **Step 2: Verify working tree is clean**

```bash
git status
```

Expected: `nothing to commit, working tree clean`. The `build.sh` script aborts if there are uncommitted changes (it checks out `v32.0.0` mid-build).

- [ ] **Step 3: Build the upgrade image**

```bash
cd integration-tests
UPGRADE_OLD_VERSION=v32.0.0 make build-stride-upgrade
```

Expected: builds two images (old `core:stride-v32.0.0` and new `core:stride-upgrade-old` + the latest stride image), pushes to GCR. End-to-end takes ~5-10 minutes.

If this fails because the v32.0.0 build can't compile against the current local SDK, that's an upstream issue — re-tag from a known-good commit, or check the rehearsal branch for a working approach.

- [ ] **Step 4: Confirm image is ready**

```bash
docker images | grep stride
```

Expected: at minimum `core:stride-upgrade-old` and a recent stride-tests image present.

- [ ] **Step 5: No commit required**

This step produces no source-tree changes.

---

### Task 9: Spin up network and verify pre-upgrade state

- [ ] **Step 1: Verify k8s namespace is empty**

```bash
cd integration-tests
make check-empty-namespace
```

Expected: exits 0. If pods exist, run `make stop` first.

- [ ] **Step 2: Start the network**

```bash
make start
```

Expected: helm install completes, then `wait-for-startup` polls until all pods are ready (~3-5 minutes). Do NOT proceed until this returns.

- [ ] **Step 3: Verify all 3 stride pods are producing blocks**

```bash
for i in 0 1 2; do
    kubectl exec stride-validator-$i -c validator -n integration -- \
        strided status | jq -r '.SyncInfo.latest_block_height // .sync_info.latest_block_height'
done
```

Expected: 3 lines, each a positive integer. Re-run a few seconds later — heights should increase.

- [ ] **Step 4: Verify ccvconsumer hex addresses match validators.json**

```bash
kubectl exec stride-validator-0 -c validator -n integration -- \
    strided q ccvconsumer all-cc-validators --output json | jq -r '.validators[].address'
```

Expected: 3 base64-encoded addresses. To verify they match `validators.json`:

```bash
kubectl exec stride-validator-0 -c validator -n integration -- \
    strided q ccvconsumer all-cc-validators --output json | \
    jq -r '.validators[].address' | \
    while read b64; do
        echo "$b64" | base64 -d | xxd -p -c 40
    done | sort
```

Expected: 3 lines matching exactly the keys in `app/upgrades/v33/validators.json` (sort the keys with `jq -r 'keys[]' app/upgrades/v33/validators.json | sort` to compare).

If the addresses don't match, the cons-key pinning failed — check Task 3's edit and verify `keys.json::cons_keys` is non-empty.

- [ ] **Step 5: Verify host zones registered**

```bash
kubectl exec stride-validator-0 -c validator -n integration -- \
    strided q stakeibc list-host-zone --output json | jq '.host_zone | length'
```

Expected: `2` (cosmoshub + osmosis). May take a minute or two after `make start` completes.

- [ ] **Step 6: No commit required**

This is a runtime verification step.

---

### Task 10: Trigger the upgrade and verify post-upgrade state

- [ ] **Step 1: Capture pre-upgrade validator hash**

```bash
PRE_HASH=$(kubectl exec stride-validator-0 -c validator -n integration -- \
    strided q block --type=height $(kubectl exec stride-validator-0 -c validator -n integration -- \
        strided status | jq -r '.SyncInfo.latest_block_height // .sync_info.latest_block_height') \
    --output json | jq -r '.block.header.validators_hash')
echo "Pre-upgrade validator hash: $PRE_HASH"
```

Save this value for step 4.

- [ ] **Step 2: Submit and pass the upgrade proposal**

```bash
make upgrade-stride
```

Expected: proposal is submitted at `latest_height + 45`, val1/val2/val3 vote yes, polls until `PROPOSAL_STATUS_PASSED`. Exits 0.

- [ ] **Step 3: Wait for the upgrade height to be reached and pods to restart**

```bash
# Check that all 3 stride pods are still up and producing blocks past the upgrade height
for i in 0 1 2; do
    height=$(kubectl exec stride-validator-$i -c validator -n integration -- \
        strided status | jq -r '.SyncInfo.latest_block_height // .sync_info.latest_block_height')
    echo "stride-validator-$i height: $height"
done
```

Wait until all three heights are at least `upgrade_height + 5`. Cosmovisor will restart each pod automatically; expect a 30-60s gap as the pod restarts on the v33 binary.

- [ ] **Step 4: Verify validator hash unchanged across the upgrade**

```bash
POST_HASH=$(kubectl exec stride-validator-0 -c validator -n integration -- \
    strided q block --type=height <upgrade_height> --output json | \
    jq -r '.block.header.validators_hash')
echo "Post-upgrade validator hash: $POST_HASH"
```

Substitute `<upgrade_height>` with the value printed in `make upgrade-stride` output (search for "Submitting proposal for v33 at height").

Expected: `POST_HASH` matches `PRE_HASH` from step 1. If different, the validator set shifted across the upgrade — investigate before proceeding.

- [ ] **Step 5: Verify POA module account exists post-upgrade**

```bash
kubectl exec stride-validator-0 -c validator -n integration -- \
    strided q auth module-account poa --output json | jq -r '.account.value.address'
```

Expected: a `stride1...` address (the POA module account). Confirms the POA module was wired in.

- [ ] **Step 6: Verify x/poa params show correct admin**

```bash
kubectl exec stride-validator-0 -c validator -n integration -- \
    strided q poa params --output json | jq -r '.params.admin'
```

Expected: `stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8` (the keys.json admin).

- [ ] **Step 7: No commit required**

Runtime verification.

---

### Task 11: Run make test-core

- [ ] **Step 1: Run the test suite**

```bash
cd integration-tests
make test-core
```

Expected: vitest runs `client/test/core.test.ts` against both `cosmoshub` and `osmosis` host zones; all tests pass.

The suite covers:
- IBC transfers (stride ↔ host)
- Liquid staking (deposit, redemption rate, unbonding, redemption record removal)
- Host zone state queries

If any test fails, capture the failure first — most likely cause is a one-time relayer flake at the upgrade height (per spec §7 R3). Retry once: `make test-core` again. If failures persist, the failure is real — investigate before declaring success.

- [ ] **Step 2: No commit required**

Pure verification.

---

## Done

Plan execution is complete when all 11 tasks above are checked off and `make test-core` has passed. The branch should NOT be merged — it has hard-coded test values in v33 source files. Pre-merge defense:

```bash
# Diff the throwaway-marked files vs. main; non-empty means the branch
# still carries throwaway state and should not be merged.
git diff main..poa-migration-ig-tests -- \
    app/upgrades/v33/validators.json \
    app/upgrades/v33/validators.json.REHEARSAL_ONLY \
    app/upgrades/v33/constants.go \
    utils/poa.go
```
