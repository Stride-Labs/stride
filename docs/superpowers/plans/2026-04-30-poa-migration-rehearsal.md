# POA Migration Rehearsal — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Run a one-time v32 → v33 (ICS → POA) migration rehearsal on the k8s integration-tests network with 8 Stride validators (5 govenators + 3 PSS-only) to validate the migration handler under multi-node consensus.

**Architecture:** Strip integration-tests network down to Stride-only with 8 validators. Pin cons keys offline so they match test-only constants baked into a rehearsal-only fork of `validators.json`, `PoaValidatorSet`, and `AdminMultisigAddress`. Build v32+v33 upgrade image, run, propose upgrade, run 4 post-upgrade verification checks.

**Tech Stack:** Bash scripts, Helm, kubectl, Cosmos SDK CLI (`strided`), Go (test-only edits to v33 source).

**Companion spec:** `docs/superpowers/specs/2026-04-30-poa-migration-rehearsal-design.md`

**Branch:** `poa-migration-rehearsal` — this work is not for merge.

**Expected test breakage on this branch:** `app/upgrades/v33/mainnet_export_test.go` will fail because mainnet state's real cons addresses won't be in our test `validators.json`. This is acceptable — branch is throwaway.

---

## Phase 1: Pre-rehearsal one-time state generation

### Task 1: Write the rehearsal-state generator script

**Files:**
- Create: `integration-tests/scripts/generate_rehearsal_state.sh`

This script produces all the deterministic inputs we need to bake into both the integration-tests config AND the v33 source. Run once locally; capture its output for use in later tasks.

- [ ] **Step 1: Create the script**

```bash
#!/bin/bash
# generate_rehearsal_state.sh
# One-time helper for the POA migration rehearsal.
# Produces: 8 priv_validator_key.json blobs (cons keys), their hex consensus
# addresses, and operator bech32 addresses derived from each validator
# mnemonic. Outputs everything to a single JSON document on stdout for the
# implementer to copy into keys.json, validators.json, and PoaValidatorSet.
#
# Requires: strided binary on PATH (any recent version that knows
# `strided init` and `strided keys`).

set -eu

WORKDIR=$(mktemp -d)
trap "rm -rf $WORKDIR" EXIT

KEYS_FILE="${KEYS_FILE:-integration-tests/network/configs/keys.json}"
NUM_VALIDATORS=8

# Ensure 8 validator entries exist in keys.json. If only 5 are present today,
# the script generates 3 fresh mnemonics and prints them so the implementer
# can append them to keys.json before re-running.
existing_count=$(jq '.validators | length' "$KEYS_FILE")
if [[ "$existing_count" -lt "$NUM_VALIDATORS" ]]; then
    echo "ERROR: keys.json has only $existing_count validators; need $NUM_VALIDATORS." >&2
    echo "Generating $((NUM_VALIDATORS - existing_count)) fresh mnemonic(s):" >&2
    for ((i=existing_count+1; i<=NUM_VALIDATORS; i++)); do
        mnemonic=$(strided keys mnemonic 2>/dev/null)
        echo "  val${i}: ${mnemonic}" >&2
    done
    echo "Append the above to keys.json::validators[] then re-run." >&2
    exit 1
fi

# Generate 8 cons keys using `strided init`. We write the priv_validator_key.json
# verbatim — same shape strided expects on disk.
declare -a CONS_KEYS_JSON
declare -a HEX_ADDRS

for ((i=0; i<NUM_VALIDATORS; i++)); do
    home="${WORKDIR}/cons-${i}"
    strided init "tmp-${i}" --chain-id rehearsal --home "$home" >/dev/null 2>&1
    cons_key=$(cat "${home}/config/priv_validator_key.json")
    hex_addr=$(jq -r '.address' <<<"$cons_key" | tr 'A-Z' 'a-z')

    CONS_KEYS_JSON[$i]="$cons_key"
    HEX_ADDRS[$i]="$hex_addr"
done

# Derive operator bech32 from each validator mnemonic.
declare -a OPERATORS
for ((i=0; i<NUM_VALIDATORS; i++)); do
    mnemonic=$(jq -r ".validators[$i].mnemonic" "$KEYS_FILE")
    name="rehearsal-val-$i"
    # Use a fresh keyring per loop to avoid name collisions across runs.
    keyring="${WORKDIR}/keyring-${i}"
    echo "$mnemonic" | strided keys add "$name" \
        --recover --keyring-backend test --home "$keyring" >/dev/null 2>&1
    operator=$(strided keys show "$name" -a \
        --keyring-backend test --home "$keyring")
    OPERATORS[$i]="$operator"
done

# Emit a single JSON document the implementer can paste into the relevant files.
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
chmod +x integration-tests/scripts/generate_rehearsal_state.sh
```

- [ ] **Step 3: Commit**

```bash
git add integration-tests/scripts/generate_rehearsal_state.sh
git commit -m "feat(rehearsal): add cons-key/operator generator script"
```

---

### Task 2: Add 3 new validator mnemonics to keys.json

**Files:**
- Modify: `integration-tests/network/configs/keys.json` (append val6, val7, val8 entries)

The generator script above will fail with a clear message and print 3 fresh mnemonics to copy in. This task captures the result.

- [ ] **Step 1: Run the generator (it will fail)**

```bash
cd /Users/sampocs/Documents/Projects/stride
bash integration-tests/scripts/generate_rehearsal_state.sh
```

Expected: exits with `ERROR: keys.json has only 5 validators` plus 3 mnemonics on stderr.

- [ ] **Step 2: Append val6, val7, val8 entries to keys.json**

The existing `validators` array in `integration-tests/network/configs/keys.json` has 5 entries. Add 3 more, using the mnemonics from Step 1's stderr output:

```json
{
  "name": "val6",
  "mnemonic": "<mnemonic 1 from generator output>"
},
{
  "name": "val7",
  "mnemonic": "<mnemonic 2 from generator output>"
},
{
  "name": "val8",
  "mnemonic": "<mnemonic 3 from generator output>"
}
```

- [ ] **Step 3: Verify keys.json parses and has 8 entries**

```bash
jq '.validators | length' integration-tests/network/configs/keys.json
```

Expected: `8`

- [ ] **Step 4: Commit**

```bash
git add integration-tests/network/configs/keys.json
git commit -m "feat(rehearsal): add val6-val8 mnemonics for 8-validator rehearsal"
```

---

### Task 3: Run generator and capture cons keys + addresses

**Files:**
- Modify: `integration-tests/network/configs/keys.json` (add `cons_keys` array)

- [ ] **Step 1: Run the generator and save output**

```bash
cd /Users/sampocs/Documents/Projects/stride
bash integration-tests/scripts/generate_rehearsal_state.sh \
    > /tmp/rehearsal-state.json
```

Expected: exits 0; file is valid JSON with `cons_keys` (array of 8 objects), `hex_addrs` (array of 8 hex strings), `operators` (array of 8 stride1... bech32s).

- [ ] **Step 2: Verify output shape**

```bash
jq '.cons_keys | length, .hex_addrs | length, .operators | length' /tmp/rehearsal-state.json
```

Expected: three lines, each `8`.

- [ ] **Step 3: Inject `cons_keys` into keys.json**

```bash
jq --slurpfile s /tmp/rehearsal-state.json \
   '. + {cons_keys: $s[0].cons_keys}' \
   integration-tests/network/configs/keys.json \
   > /tmp/keys.json && \
mv /tmp/keys.json integration-tests/network/configs/keys.json
```

- [ ] **Step 4: Verify keys.json has cons_keys array**

```bash
jq '.cons_keys | length' integration-tests/network/configs/keys.json
```

Expected: `8`.

- [ ] **Step 5: Save the hex_addrs and operators arrays for later tasks**

Keep `/tmp/rehearsal-state.json` around — Tasks 10 and 11 will read `hex_addrs` and `operators` from it.

- [ ] **Step 6: Commit**

```bash
git add integration-tests/network/configs/keys.json
git commit -m "feat(rehearsal): pin 8 cons keys in keys.json"
```

---

## Phase 2: Integration-tests config updates

### Task 4: Strip cosmoshub and osmosis from values.yaml

**Files:**
- Modify: `integration-tests/network/values.yaml`

- [ ] **Step 1: Replace values.yaml contents**

```yaml
namespace: integration

images:
  chains: gcr.io/stride-nodes/integration-tests/chains
  relayer: gcr.io/stride-nodes/integration-tests/relayer:v2.5.2
  hermes: gcr.io/stride-nodes/integration-tests/hermes:v1.9.0

activeChains:
  - stride

# No relayers — Stride is the only chain in this rehearsal network.
relayers: []

chainConfigs:
  stride:
    binary: strided
    version: latest
    numValidators: 8
    home: .stride
    denom: ustrd
    decimals: 6
```

- [ ] **Step 2: Lint helm chart**

```bash
cd integration-tests && make lint
```

Expected: `1 chart(s) linted, 0 chart(s) failed`.

- [ ] **Step 3: Commit**

```bash
git add integration-tests/network/values.yaml
git commit -m "feat(rehearsal): trim values.yaml to stride-only with 8 validators"
```

---

### Task 5: Pin cons keys in init-chain.sh

**Files:**
- Modify: `integration-tests/network/scripts/init-chain.sh`

The `add_validators` function (lines ~62-96 of init-chain.sh) currently calls `binaryd init` for each validator (which generates a fresh `priv_validator_key.json`). We want to overwrite that file with the pinned cons key from `keys.json::cons_keys[i-1]` BEFORE the pubkey is read for `add-consumer-section`.

- [ ] **Step 1: Modify the validator loop in `add_validators()`**

In `init-chain.sh`, find the `add_validators()` function. Inside the `for (( i=1; i <= $NUM_VALIDATORS; i++ ))` loop, after the `if [[ "$i" == "1" ]]; then validator_home=${CHAIN_HOME} else ... $BINARY init ... fi` block, add this just before the `# Save the node IDs and keys to the API` comment:

```bash
# Overwrite the auto-generated cons key with the pinned rehearsal cons key.
# The index in cons_keys is i-1 (jq is 0-indexed; the loop is 1-indexed).
cons_key_index=$((i - 1))
jq -r ".cons_keys[$cons_key_index]" ${KEYS_FILE} \
    > ${validator_home}/config/priv_validator_key.json
```

- [ ] **Step 2: Verify shell syntax**

```bash
bash -n integration-tests/network/scripts/init-chain.sh
```

Expected: no output (exit 0).

- [ ] **Step 3: Commit**

```bash
git add integration-tests/network/scripts/init-chain.sh
git commit -m "feat(rehearsal): pin cons keys in init-chain.sh"
```

---

### Task 6: Pin cons keys in init-node.sh

**Files:**
- Modify: `integration-tests/network/scripts/init-node.sh`

Each non-main validator pod runs `init-node.sh` which downloads its `priv_validator_key.json` from the API (originally uploaded by pod 0). With Task 5's change, that uploaded file is already the pinned cons key — but we add a belt-and-suspenders overwrite for self-containment.

- [ ] **Step 1: Add overwrite to `update_config()`**

In `init-node.sh::update_config()`, after the line:

```bash
download_shared_file ${VALIDATOR_KEYS_DIR}/${CHAIN_NAME}/val${VALIDATOR_INDEX}.json ${CHAIN_HOME}/config/priv_validator_key.json 
```

Add:

```bash
# Belt-and-suspenders: ensure the downloaded cons key matches the pinned one
# from keys.json. This makes the script self-contained even if pod 0's
# pinning step ever regresses.
cons_key_index=$((VALIDATOR_INDEX - 1))
jq -r ".cons_keys[$cons_key_index]" ${KEYS_FILE} \
    > ${CHAIN_HOME}/config/priv_validator_key.json
```

- [ ] **Step 2: Verify shell syntax**

```bash
bash -n integration-tests/network/scripts/init-node.sh
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add integration-tests/network/scripts/init-node.sh
git commit -m "feat(rehearsal): pin cons keys in init-node.sh"
```

---

### Task 7: Skip govenator registration on pods 5-7

**Files:**
- Modify: `integration-tests/network/scripts/create-validator.sh`

- [ ] **Step 1: Add the skip guard**

`create-validator.sh` already sources `scripts/config.sh` which sets
`POD_INDEX=${HOSTNAME##*-}`. Replace the existing `main()` function with:

```bash
main() {
    # PSS-only validators: pods with index >= 5 do NOT register as govenators.
    # This mirrors the mainnet shape where ~3 of 8 PSS validators have
    # ICS-key-assigned consensus keys not in x/staking, exercising the
    # distrwrapper stake-iteration path post-upgrade.
    if [[ "$POD_INDEX" -ge 5 ]]; then
        echo "Skipping govenator registration for PSS-only validator (pod $POD_INDEX)"
        exit 0
    fi

    echo "Adding validator..."
    wait_for_startup
    add_keys
    create_validator
    echo "Done"
}
```

- [ ] **Step 2: Verify shell syntax**

```bash
bash -n integration-tests/network/scripts/create-validator.sh
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add integration-tests/network/scripts/create-validator.sh
git commit -m "feat(rehearsal): skip govenator registration on pods 5-7"
```

---

### Task 8: Vote from val1-val5 in upgrade.sh

**Files:**
- Modify: `integration-tests/network/scripts/upgrade.sh`

Currently `upgrade.sh` hardcodes `STRIDED0`, `STRIDED1`, `STRIDED2` and votes from val1, val2, val3. With 5 govenators we want votes from val1–val5.

- [ ] **Step 1: Replace upgrade.sh in full**

Replace the entire contents of `integration-tests/network/scripts/upgrade.sh` with:

```bash
#!/bin/bash

# NOTE: This script should be run locally (outside of the k8s pod)

set -eu

NAMESPACE=integration
UPGRADE_BUFFER=45 # blocks

# Helper to invoke strided in a specific validator pod.
strided_in() {
    pod_index="$1"; shift
    kubectl exec -it stride-validator-$pod_index -c validator -n $NAMESPACE -- strided "$@"
}

trim_tx() {
    grep -E "code:|txhash:" | sed 's/^[[:space:]]*//'
}

upgrade_name=$(kubectl exec stride-validator-0 -c validator -n $NAMESPACE -- printenv UPGRADE_NAME)
latest_height=$(strided_in 0 status | jq -r 'if .SyncInfo then .SyncInfo.latest_block_height else .sync_info.latest_block_height end')
upgrade_height=$((latest_height+UPGRADE_BUFFER))

echo -e "\nSubmitting proposal for $upgrade_name at height $upgrade_height...\n"
kubectl exec -it stride-validator-0 -c validator -n $NAMESPACE -- \
    bash scripts/propose_upgrade.sh $upgrade_name $upgrade_height | trim_tx

sleep 5
echo -e "\nProposal:\n"
proposal_id=$(strided_in 0 q gov proposals --output json | jq -r '.proposals | max_by(.id | tonumber).id')
strided_in 0 query gov proposal $proposal_id

sleep 1
echo -e "\nVoting on proposal #$proposal_id...\n"
for i in 0 1 2 3 4; do
    val_name="val$((i + 1))"
    echo "${val_name}:"
    strided_in "$i" tx gov vote $proposal_id yes --from "$val_name" -y | trim_tx
done

sleep 5
echo -e "\nVote confirmation:\n"
strided_in 0 query gov tally $proposal_id

echo -e "\nProposal Status:\n"
while true; do
    status=$(strided_in 0 query gov proposal $proposal_id --output json | jq -r '.proposal.status')
    if [[ "$status" == "PROPOSAL_STATUS_VOTING_PERIOD" ]]; then
        echo "Proposal still in progress..."
        sleep 5
    elif [[ "$status" == "PROPOSAL_STATUS_PASSED" ]]; then
        echo "Proposal passed!"
        exit 0
    elif [[ "$status" == "PROPOSAL_STATUS_REJECTED" ]]; then
        echo "Proposal Failed!"
        exit 1
    else
        echo "Unknown proposal status: $status"
        exit 1
    fi
done
```

- [ ] **Step 2: Verify shell syntax**

```bash
bash -n integration-tests/network/scripts/upgrade.sh
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add integration-tests/network/scripts/upgrade.sh
git commit -m "feat(rehearsal): vote from val1-val5 in upgrade.sh"
```

---

## Phase 3: Smoke test v32 spinup

### Task 9: Build the upgrade image and verify v32 startup

This task validates Phase 2's config edits before we touch v33 source. Build the upgrade image (which contains both v32 and v33 binaries — the v33 here is whatever the branch currently has, with mainnet `validators.json`/`PoaValidatorSet`/`AdminMultisigAddress` still intact). Cosmovisor will boot from v32 and never trigger the upgrade since we won't propose it. We just verify the v32 startup phase works against our edited config.

**Files:** none modified.

- [ ] **Step 1: Verify clean working tree**

```bash
cd /Users/sampocs/Documents/Projects/stride
git status --short
```

Expected: empty output. The upgrade build script does `git checkout v32.0.0` internally and refuses to run if there are uncommitted changes.

- [ ] **Step 2: Build the upgrade image**

```bash
cd /Users/sampocs/Documents/Projects/stride/integration-tests
UPGRADE_OLD_VERSION=v32.0.0 make build-stride-upgrade
```

Expected: build succeeds, image pushed to `gcr.io/stride-nodes/integration-tests/chains/stride:latest`.

- [ ] **Step 2: Spin up the network**

```bash
cd /Users/sampocs/Documents/Projects/stride/integration-tests
make start
```

Expected: helm install completes, `make wait-for-startup` exits 0 once all 8 stride pods are ready.

- [ ] **Step 3: Verify 8 cons validators present**

```bash
kubectl exec -n integration stride-validator-0 -c validator -- \
    strided q ccvconsumer all-cc-validators --output json \
    | jq '. | length'
```

Expected: `8`.

- [ ] **Step 4: Verify cons hex addresses match the pinned values**

```bash
kubectl exec -n integration stride-validator-0 -c validator -- \
    strided q ccvconsumer all-cc-validators --output json \
    | jq -r '.[].address' | sort \
> /tmp/onchain-hex-addrs.txt

jq -r '.hex_addrs[]' /tmp/rehearsal-state.json | sort \
> /tmp/expected-hex-addrs.txt

diff /tmp/onchain-hex-addrs.txt /tmp/expected-hex-addrs.txt
```

Expected: no diff (both lists identical, sorted).

- [ ] **Step 5: Verify exactly 5 govenators**

```bash
kubectl exec -n integration stride-validator-0 -c validator -- \
    strided q staking validators --output json \
    | jq '.validators | length'
```

Expected: `5`.

- [ ] **Step 6: Verify all 8 pods are signing blocks**

```bash
for i in $(seq 0 7); do
  kubectl exec -n integration stride-validator-$i -c validator -- \
    strided status | jq -r '.sync_info.latest_block_height // .SyncInfo.latest_block_height'
done
```

Expected: all 8 print non-zero heights, growing.

- [ ] **Step 7: Tear down**

```bash
make stop
```

If any step above fails, **stop and debug** before proceeding to Phase 4. Fixing config issues with v33 source already edited is much harder than fixing them now.

---

## Phase 4: v33 source-tree rehearsal-only edits

### Task 10: Replace validators.json with test cons addresses

**Files:**
- Modify: `app/upgrades/v33/validators.json` (full replacement)
- Create: `app/upgrades/v33/validators.json.REHEARSAL_ONLY` (sidecar marker)

- [ ] **Step 1: Generate the new validators.json**

```bash
cd /Users/sampocs/Documents/Projects/stride
jq -r '
  .hex_addrs as $h
  | reduce range(0; ($h | length)) as $i ({};
      . + { ($h[$i]): "Validator\($i + 1)" })
' /tmp/rehearsal-state.json \
> app/upgrades/v33/validators.json
```

- [ ] **Step 2: Verify**

```bash
jq '. | length' app/upgrades/v33/validators.json
```

Expected: `8`.

```bash
jq -r 'to_entries | .[].value' app/upgrades/v33/validators.json | sort
```

Expected: `Validator1` through `Validator8`, one per line.

- [ ] **Step 3: Create the sidecar marker file**

`validators.json` cannot carry a comment. The marker file is a tripwire for anyone reading the directory.

```bash
cat > app/upgrades/v33/validators.json.REHEARSAL_ONLY <<'EOF'
This directory's `validators.json` has been replaced with TEST consensus
addresses on the `poa-migration-rehearsal` branch. DO NOT MERGE this branch.

If you are reviewing a v33 PR and see this file, the PR is including
rehearsal artifacts. The mainnet validators.json should map 8 real
mainnet hex consensus addresses to monikers like "Polkachu", "L5", etc.
EOF
```

- [ ] **Step 4: Commit**

```bash
git add app/upgrades/v33/validators.json app/upgrades/v33/validators.json.REHEARSAL_ONLY
git commit -m "rehearsal: replace v33/validators.json with test cons addresses

REHEARSAL ONLY -- DO NOT MERGE. Required for the test validator set to
pass the v33 handler's moniker join."
```

---

### Task 11: Replace utils/poa.go PoaValidatorSet

**Files:**
- Modify: `utils/poa.go`

- [ ] **Step 1: Read the current operators array from /tmp**

```bash
jq -r '.operators[]' /tmp/rehearsal-state.json
```

Capture the 8 stride1... addresses (one per line).

- [ ] **Step 2: Replace utils/poa.go**

Replace the entire contents of `utils/poa.go` with:

```go
// REHEARSAL ONLY — DO NOT MERGE
// PoaValidatorSet has been replaced with test operator addresses generated
// from the rehearsal validator mnemonics in
// integration-tests/network/configs/keys.json. The mainnet set lives on the
// `main` branch.

package utils

import sdkmath "cosmossdk.io/math"

// WARNING: DO NOT MODIFY

// Validators are paid 15% of revenue
var PoaValPaymentRate = sdkmath.LegacyMustNewDecFromStr("0.15")

type PoaValidator struct {
	Moniker    string
	Operator   string // sdk.AccAddress bech32 — the payout + POA OperatorAddress
	HubAddress string
}

var PoaValidatorSet = []PoaValidator{
	{Moniker: "Validator1", Operator: "<operator0 from rehearsal-state.json>", HubAddress: ""},
	{Moniker: "Validator2", Operator: "<operator1 from rehearsal-state.json>", HubAddress: ""},
	{Moniker: "Validator3", Operator: "<operator2 from rehearsal-state.json>", HubAddress: ""},
	{Moniker: "Validator4", Operator: "<operator3 from rehearsal-state.json>", HubAddress: ""},
	{Moniker: "Validator5", Operator: "<operator4 from rehearsal-state.json>", HubAddress: ""},
	{Moniker: "Validator6", Operator: "<operator5 from rehearsal-state.json>", HubAddress: ""},
	{Moniker: "Validator7", Operator: "<operator6 from rehearsal-state.json>", HubAddress: ""},
	{Moniker: "Validator8", Operator: "<operator7 from rehearsal-state.json>", HubAddress: ""},
}
```

Substitute the placeholder `<operator0>` ... `<operator7>` with the actual bech32 strings from Step 1.

- [ ] **Step 3: Verify the file builds**

```bash
go build ./utils/...
```

Expected: no output (exit 0).

- [ ] **Step 4: Commit**

```bash
git add utils/poa.go
git commit -m "rehearsal: replace PoaValidatorSet with test operator addresses

REHEARSAL ONLY -- DO NOT MERGE."
```

---

### Task 12: Replace AdminMultisigAddress in constants.go

**Files:**
- Modify: `app/upgrades/v33/constants.go`

- [ ] **Step 1: Edit constants.go**

Find:

```go
const AdminMultisigAddress = "stride1fduug6m38gyuqt3wcgc2kcgr9nnte0n7ssn27e"
```

Replace with:

```go
// REHEARSAL ONLY — DO NOT MERGE
// AdminMultisigAddress points at the integration-tests `admin` account so the
// rehearsal can drive POA admin behavior with a known signer. The mainnet
// multisig address lives on the `main` branch.
const AdminMultisigAddress = "stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8"
```

- [ ] **Step 2: Verify the package builds**

```bash
go build ./app/upgrades/v33/...
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add app/upgrades/v33/constants.go
git commit -m "rehearsal: point AdminMultisigAddress at integration-tests admin

REHEARSAL ONLY -- DO NOT MERGE."
```

---

## Phase 5: Build and run upgrade

### Task 13: Build the v32+v33 upgrade image

**Files:** none modified.

`network/scripts/build.sh` requires a clean working tree (it does `git checkout v32.0.0` to build the old binary). All rehearsal commits from Phases 1-4 must already be committed before running this.

The same script tags the upgrade image as `chains/stride:latest` — same tag the non-upgrade build uses — so no `values.yaml` adjustment is needed.

- [ ] **Step 1: Verify clean working tree**

```bash
cd /Users/sampocs/Documents/Projects/stride
git status --short
```

Expected: empty output (no uncommitted changes).

- [ ] **Step 2: Build the upgrade image**

```bash
cd /Users/sampocs/Documents/Projects/stride/integration-tests
UPGRADE_OLD_VERSION=v32.0.0 make build-stride-upgrade
```

Expected: build completes, image pushed to GCR as `gcr.io/stride-nodes/integration-tests/chains/stride:latest`. The image contains v32 at `cosmovisor/genesis/bin/strided` and v33 (built from this branch with all rehearsal edits) at `cosmovisor/upgrades/v33/bin/strided`. The build script auto-detects `v33` as the upgrade name from `ls app/upgrades | sort -V | tail -1`.

- [ ] **Step 3: Verify both binaries are in the local image**

```bash
docker run --rm core:stride-upgrade-old strided version
docker run --rm core:stride strided version
```

Expected: first prints `v32.0.0` (or close); second prints a version string from the current branch.

---

### Task 14: Spin up network on v32 and trigger upgrade

**Files:** none modified.

- [ ] **Step 1: Spin up the network**

```bash
cd /Users/sampocs/Documents/Projects/stride/integration-tests
make start
```

Expected: helm install completes; `make wait-for-startup` exits 0.

- [ ] **Step 2: Re-run the Task 9 sanity checks**

```bash
# 8 cons validators
kubectl exec -n integration stride-validator-0 -c validator -- \
    strided q ccvconsumer all-cc-validators --output json | jq '. | length'
# Expected: 8

# 5 govenators
kubectl exec -n integration stride-validator-0 -c validator -- \
    strided q staking validators --output json | jq '.validators | length'
# Expected: 5
```

- [ ] **Step 3: Snapshot pre-upgrade state**

```bash
# Pre-upgrade validator hash from a recent block
PRE_HEIGHT=$(kubectl exec -n integration stride-validator-0 -c validator -- \
    strided status | jq -r '.sync_info.latest_block_height // .SyncInfo.latest_block_height')
PRE_HASH=$(kubectl exec -n integration stride-validator-0 -c validator -- \
    strided q block --type=height $PRE_HEIGHT --output json \
    | jq -r '.block.header.validators_hash')
echo "Pre-upgrade height: $PRE_HEIGHT, validators_hash: $PRE_HASH"
```

Save these values — they're used in Task 15.

- [ ] **Step 4: Submit and pass the upgrade**

```bash
make upgrade-stride
```

Expected: prints proposal submission, 5 votes (one per val1-val5), tally output, and `Proposal passed!`.

- [ ] **Step 5: Wait for the upgrade to execute**

The upgrade height is `latest_height + 45` from when the proposal was submitted. Wait for cosmovisor to swap binaries:

```bash
kubectl logs -n integration stride-validator-0 -c validator -f
```

Watch for `UPGRADE NEEDED` panic, binary swap, and the chain resuming on v33. Ctrl-C once you see new blocks being signed post-restart.

---

## Phase 6: Post-upgrade tests

### Task 15: Test 1 — block production continues

- [ ] **Step 1: Verify all 8 pods recovered post-upgrade**

```bash
for i in $(seq 0 7); do
  height=$(kubectl exec -n integration stride-validator-$i -c validator -- \
    strided status | jq -r '.sync_info.latest_block_height // .SyncInfo.latest_block_height')
  echo "validator-$i: $height"
done
```

Expected: all 8 print heights well past the upgrade height (and increasing if you re-run).

- [ ] **Step 2: Read the post-upgrade validators_hash**

Pick a height a few blocks past the upgrade height:

```bash
POST_HEIGHT=<value from Step 1, e.g., upgrade_height + 5>
POST_HASH=$(kubectl exec -n integration stride-validator-0 -c validator -- \
    strided q block --type=height $POST_HEIGHT --output json \
    | jq -r '.block.header.validators_hash')
echo "Post-upgrade height: $POST_HEIGHT, validators_hash: $POST_HASH"
```

- [ ] **Step 3: Compare to pre-upgrade**

```bash
echo "Pre:  $PRE_HASH"
echo "Post: $POST_HASH"
[[ "$PRE_HASH" == "$POST_HASH" ]] && echo "PASS" || echo "FAIL"
```

Expected: `PASS`.

If FAIL: chain has shifted validators across the upgrade — the migration is broken. Stop, capture logs, do not proceed.

---

### Task 16: Test 2 — tx fees route to POA

- [ ] **Step 1: Find the POA module account address**

```bash
POA_ADDR=$(kubectl exec -n integration stride-validator-0 -c validator -- \
    strided q auth module-account poa --output json \
    | jq -r '.account.value.address // .account.address')
echo "POA module account: $POA_ADDR"
```

- [ ] **Step 2: Find the fee_collector module account address**

```bash
FEE_ADDR=$(kubectl exec -n integration stride-validator-0 -c validator -- \
    strided q auth module-account fee_collector --output json \
    | jq -r '.account.value.address // .account.address')
echo "fee_collector: $FEE_ADDR"
```

- [ ] **Step 3: Snapshot balances**

```bash
PRE_POA=$(kubectl exec -n integration stride-validator-0 -c validator -- \
    strided q bank balances $POA_ADDR --output json \
    | jq -r '.balances[] | select(.denom=="ustrd").amount // "0"')
PRE_FEE=$(kubectl exec -n integration stride-validator-0 -c validator -- \
    strided q bank balances $FEE_ADDR --output json \
    | jq -r '.balances[] | select(.denom=="ustrd").amount // "0"')
echo "Pre-tx — POA: $PRE_POA, fee_collector: $PRE_FEE"
```

- [ ] **Step 4: Submit a tx with non-zero fees**

```bash
ADMIN_ADDR=stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8
FAUCET_ADDR=$(kubectl exec -n integration stride-validator-0 -c validator -- \
    strided keys show faucet -a)
kubectl exec -n integration stride-validator-0 -c validator -- \
    strided tx bank send $FAUCET_ADDR $ADMIN_ADDR 100ustrd \
    --fees 5000ustrd --from faucet -y
```

- [ ] **Step 5: Wait one block, snapshot again**

```bash
sleep 6
POST_POA=$(kubectl exec -n integration stride-validator-0 -c validator -- \
    strided q bank balances $POA_ADDR --output json \
    | jq -r '.balances[] | select(.denom=="ustrd").amount // "0"')
echo "Post-tx — POA: $POST_POA (was $PRE_POA, +$((POST_POA - PRE_POA)))"
```

Expected: `POST_POA - PRE_POA == 5000` (or close to it, accounting for any inflation that landed in the same block — the *delta* should at minimum include the 5000 fee).

If POA went up by ≥5000: PASS — fees routed to POA. If POA is unchanged and `fee_collector` went up by 5000: FAIL — ante handler change didn't take effect.

---

### Task 17: Test 3 — govenator rewards accrue and withdraw

- [ ] **Step 1: Pick a delegator (val1's self-delegation)**

```bash
DELEGATOR=$(kubectl exec -n integration stride-validator-0 -c validator -- \
    strided keys show val1 -a)
VALIDATOR=$(kubectl exec -n integration stride-validator-0 -c validator -- \
    strided q staking validators --output json \
    | jq -r '.validators[0].operator_address')
echo "Delegator: $DELEGATOR, Validator: $VALIDATOR"
```

- [ ] **Step 2: Pre-baseline rewards**

```bash
PRE_REWARDS=$(kubectl exec -n integration stride-validator-0 -c validator -- \
    strided q distribution rewards $DELEGATOR --output json \
    | jq -r '.total[] | select(.denom=="ustrd").amount // "0"')
echo "Pre-baseline rewards: $PRE_REWARDS"
```

- [ ] **Step 3: Wait 10–20 blocks**

```bash
sleep 60
```

- [ ] **Step 4: Post-baseline rewards**

```bash
POST_REWARDS=$(kubectl exec -n integration stride-validator-0 -c validator -- \
    strided q distribution rewards $DELEGATOR --output json \
    | jq -r '.total[] | select(.denom=="ustrd").amount // "0"')
echo "Post-baseline rewards: $POST_REWARDS"
```

Expected: `POST_REWARDS > PRE_REWARDS` (strict increase). If equal: distribution isn't allocating — distrwrapper bug.

- [ ] **Step 5: Snapshot delegator bank balance**

```bash
PRE_BAL=$(kubectl exec -n integration stride-validator-0 -c validator -- \
    strided q bank balances $DELEGATOR --output json \
    | jq -r '.balances[] | select(.denom=="ustrd").amount // "0"')
echo "Pre-withdraw balance: $PRE_BAL"
```

- [ ] **Step 6: Withdraw rewards**

```bash
kubectl exec -n integration stride-validator-0 -c validator -- \
    strided tx distribution withdraw-rewards $VALIDATOR \
    --fees 5000ustrd --from val1 -y
sleep 6
```

- [ ] **Step 7: Verify balance increased**

```bash
POST_BAL=$(kubectl exec -n integration stride-validator-0 -c validator -- \
    strided q bank balances $DELEGATOR --output json \
    | jq -r '.balances[] | select(.denom=="ustrd").amount // "0"')
echo "Post-withdraw balance: $POST_BAL (delta: $((POST_BAL - PRE_BAL)))"
```

Expected: `POST_BAL > PRE_BAL - 5000` (i.e., rewards > tx fee). The exact delta should be roughly `POST_REWARDS - 5000`.

---

### Task 18: Test 4 — POA fee distribution via MsgWithdrawFees

- [ ] **Step 1: Confirm POA module account has accumulated balance**

```bash
POA_BAL=$(kubectl exec -n integration stride-validator-0 -c validator -- \
    strided q bank balances $POA_ADDR --output json \
    | jq -r '.balances[] | select(.denom=="ustrd").amount // "0"')
echo "POA balance: $POA_BAL"
```

If $POA_BAL is small, wait more blocks and re-query — the 15% slice from `distrwrapper` should accumulate steadily.

- [ ] **Step 2: Snapshot val1's operator account balance**

```bash
VAL1_OP=$(kubectl exec -n integration stride-validator-0 -c validator -- \
    strided keys show val1 -a)
PRE_VAL1=$(kubectl exec -n integration stride-validator-0 -c validator -- \
    strided q bank balances $VAL1_OP --output json \
    | jq -r '.balances[] | select(.denom=="ustrd").amount // "0"')
echo "val1 pre-withdraw: $PRE_VAL1"

# Cross-check: query the withdrawable amount POA thinks val1 is owed.
WITHDRAWABLE=$(kubectl exec -n integration stride-validator-0 -c validator -- \
    strided q poa withdrawable-fees $VAL1_OP --output json \
    | jq -r '.amount[] | select(.denom=="ustrd").amount // "0"')
echo "val1 withdrawable per POA: $WITHDRAWABLE"
```

- [ ] **Step 3: Submit MsgWithdrawFees**

```bash
kubectl exec -n integration stride-validator-0 -c validator -- \
    strided tx poa withdraw-fees \
    --fees 5000ustrd --from val1 -y
sleep 6
```

`withdraw-fees` takes no positional args — the operator is inferred from the `--from` key (POA CLI: `enterprise/poa@v1.0.0/x/poa/client/cli/tx.go:234-258`).

- [ ] **Step 4: Verify val1 received their 1/8 share**

```bash
POST_VAL1=$(kubectl exec -n integration stride-validator-0 -c validator -- \
    strided q bank balances $VAL1_OP --output json \
    | jq -r '.balances[] | select(.denom=="ustrd").amount // "0"')
echo "val1 post-withdraw: $POST_VAL1 (delta: $((POST_VAL1 - PRE_VAL1)))"
```

Expected: `delta == WITHDRAWABLE - 5000` (within rounding). The `WITHDRAWABLE` query from Step 2 is the source of truth — POA's lazy checkpoint accounts for fees per validator, so `WITHDRAWABLE` is what you'd literally receive minus the tx fee.

- [ ] **Step 5: Optional — repeat with val6 to test the PSS-only case**

```bash
VAL6_OP=$(kubectl exec -n integration stride-validator-0 -c validator -- \
    strided keys show val6 -a)
PRE_VAL6=$(kubectl exec -n integration stride-validator-0 -c validator -- \
    strided q bank balances $VAL6_OP --output json \
    | jq -r '.balances[] | select(.denom=="ustrd").amount // "0"')

kubectl exec -n integration stride-validator-5 -c validator -- \
    strided tx poa withdraw-fees \
    --fees 5000ustrd --from val6 -y
sleep 6

POST_VAL6=$(kubectl exec -n integration stride-validator-5 -c validator -- \
    strided q bank balances $VAL6_OP --output json \
    | jq -r '.balances[] | select(.denom=="ustrd").amount // "0"')
echo "val6 post-withdraw: $POST_VAL6 (delta: $((POST_VAL6 - PRE_VAL6)))"
```

Expected: delta > 0. This confirms POA pays out to validators that have no x/staking record — the load-bearing case for the 5/3 split.

Note: val6 needs ustrd in its account to pay the tx fee. If `PRE_VAL6` is 0, fund it first via `strided tx bank send` from the faucet.

---

### Task 19: Tear down and document

**Files:** none modified (or commit a brief rehearsal-results notes file if you like).

- [ ] **Step 1: Tear down the network**

```bash
cd /Users/sampocs/Documents/Projects/stride/integration-tests
make stop
```

- [ ] **Step 2: Capture results**

In whatever channel makes sense (Slack, internal doc, etc.), report:

- Did Test 1 PASS? (block production / validators_hash unchanged)
- Did Test 2 PASS? (tx fees → POA, not fee_collector)
- Did Test 3 PASS? (govenator rewards accrue + withdraw)
- Did Test 4 PASS? (POA `MsgWithdrawFees` paid val1 ~ 1/8 of POA balance; val6 also paid)
- Any unexpected log lines, panics, or warnings during the upgrade
- Wall-clock time of the upgrade (proposal pass → first post-upgrade block)

If all 4 pass, the v33 migration handler is validated under multi-node consensus. Mark the rehearsal complete.

---

## Notes for the implementer

**Branch hygiene:** Every commit on this branch should land on `poa-migration-rehearsal` and never on `main`. Before pushing, double-check `git branch --show-current`.

**Expected unit-test failure:** `app/upgrades/v33/mainnet_export_test.go` will fail when run against the rehearsal-edited `validators.json` (real mainnet cons addresses won't be in our test map). This is acceptable on this branch.

**If the helm install hangs:** check `kubectl get pods -n integration` and `kubectl describe pod stride-validator-0 -n integration`. Most common cause on first-time spinup with 8 vals: cluster resource quota exhausted. Mitigation: drop `cpu`/`memory` requests in `network/templates/validator.yaml` lines 107-113 from `600m`/`700M` to `400m`/`500M`.

**If the upgrade fails at the upgrade height:** capture full logs from all 8 pods (`for i in $(seq 0 7); do kubectl logs -n integration stride-validator-$i -c validator > /tmp/val-$i.log; done`). The most likely failure modes:

- "validator X has no moniker in v33 validators.json" — pinning didn't take effect; check Tasks 5/6 outcomes.
- "expected 8 validators in consumer keeper, got N" — `add-consumer-section` didn't get all 8 pubkeys; check init-chain.sh ran fully.
- POA InitGenesis panic — likely a pubkey decoding mismatch; verify the cons keys in `keys.json::cons_keys` are valid `priv_validator_key.json` blobs.

**Cleanup if the rehearsal is abandoned:** `git checkout main && git branch -D poa-migration-rehearsal` deletes the branch. Throwaway by design.
