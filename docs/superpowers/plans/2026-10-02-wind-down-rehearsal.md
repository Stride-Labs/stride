# Wind-Down Rehearsal Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers-fast:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the throwaway branch, network config and phase scripts that run the wind-down rehearsal on the k8s integration network, then run it and log every phase.

**Architecture:** Two branches: `wind-down-rehearsal` (this one; v35 with rehearsal constants, network config, scripts) and `wind-down-rehearsal-v34` (the `v34.1.0` tag with the staketia constants patched, used as the "old" binary). The network starts on v34, a seed script creates the pre-upgrade state, a gov proposal upgrades to v35, and one bash script per phase drives the ops window from the operator's machine through `kubectl exec`, appending every command and result to `docs/wind-down/rehearsal-log.md`.

**Tech Stack:** bash + kubectl + helm (existing `integration-tests/` harness), `strided`/`gaiad`/`osmosisd` CLIs inside the pods, python3 for the ops scripts, docker buildx for the images.

Spec: `docs/superpowers/specs/2026-10-02-wind-down-rehearsal-design.md`. Wind-down spec: `docs/superpowers/specs/2026-09-18-protocol-wind-down-design.md`.

## Global Constraints

- Everything is on the `wind-down-rehearsal` worktree at `/Users/sampocs/Documents/Projects/stride-worktrees/wind-down-rehearsal`. Neither branch is ever merged or pushed to main. The branch has no upstream; never `git push` it without `-u origin wind-down-rehearsal`.
- Kube context is `integration` (`kubectl config use-context integration`); namespace `integration`. Two clusters have an `integration` namespace; always confirm `kubectl config current-context` prints `integration` before any `kubectl` call.
- Old binary: tag `v34.1.0` (mainnet's version). Upgrade name: `v35`.
- Chain ids: `stride-test-1`, `cosmoshub-test-1`, `osmosis-test-1`. Validators: stride 4, cosmoshub 8, osmosis 3.
- Expected channels (relayers start in this order, verified in phase 0): stride `channel-0` ↔ cosmoshub `channel-0`; stride `channel-1` ↔ osmosis `channel-0`; cosmoshub `channel-1` ↔ osmosis `channel-1`. Stride connections: `connection-0` (hub), `connection-1` (osmosis).
- Denoms: ATOM on Stride `ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2` (`transfer/channel-0/uatom`); OSMO on Stride is `ibc/` + sha256 of `transfer/channel-1/uosmo` (computed in Task 1); stTokens `stuatom`, `stuosmo`.
- Epochs on Stride: day 180s, stride_epoch 45s, mint 5s, hour 30s. Unbonding 240s on every chain. Gov voting 30s.
- `WindDownTransferTimeout` is 15 minutes on the branch.
- Every phase script sources `integration-tests/rehearsal/lib.sh`, is idempotent where possible, and logs through `log_cmd` so the log file is the record.
- Inside the Stride pod the v34 CLI is `/usr/local/bin/strided`; after the upgrade the v35 CLI is `/home/validator/.stride/cosmovisor/upgrades/v35/bin/strided`. `lib.sh` exposes both as `strided_old` and `strided_new`; `strided` resolves to whichever matches the running binary.
- All multisig txs sign with `--sign-mode amino-json` (what a multisig does on mainnet).

---

## Foundation tasks (serial)

### Task 1: Rehearsal keys and addresses

**Files:**
- Create: `integration-tests/rehearsal/keys.json`
- Create: `integration-tests/rehearsal/addresses.env`
- Create: `integration-tests/rehearsal/gen_keys.sh`
- Modify: `integration-tests/network/configs/keys.json` (validators: 5 → 8 entries)

**Interfaces:**
- Produces: `addresses.env` with these variables, sourced by every later task:
  `ADMIN_MS_STRIDE`, `VAULT_MS_OSMO`, `ADMIN_MS_MEMBERS` (comma list of three key names `m1,m2,m3`), `SWEEP_OPERATOR`, `STAKETIA_DEPOSIT`, `STAKETIA_REDEMPTION`, `STAKETIA_CLAIM`, `STAKETIA_SAFE`, `STAKETIA_OPERATOR_STRIDE`, `STAKETIA_OPERATOR_HUB`, `HUB_MS_COSMOS` (the Hub delegation multisig), `HUB_MS_MEMBERS` (`d1,d2,d3`), `HUB_REWARD`, `ATOM_ON_STRIDE`, `OSMO_ON_STRIDE`, `HOLDER_BASE`, `HOLDER_VESTING`, and the osmo-prefixed forms `HOLDER_BASE_OSMO`, `HOLDER_VESTING_OSMO`, `USER1_STRIDE`, `USER1_OSMO`, `USER1_COSMOS`.
- Produces: `keys.json` with `{"multisig_members": [{name, mnemonic}×3], "hub_multisig_members": [×3], "sweep_operator": {}, "staketia": {"deposit", "redemption", "claim", "safe", "operator", "reward"}, "holders": {"base", "vesting"}}`.
- Review: no

- [ ] **Step 1: Write `gen_keys.sh`**

```bash
#!/bin/bash
# Generates every rehearsal key once and derives the addresses the branch hard-codes.
# Uses the locally installed strided only for key derivation (any version works).
set -euo pipefail
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KEYS=$DIR/keys.json
OUT=$DIR/addresses.env
HOME_TMP=$(mktemp -d)
K="strided keys --keyring-backend test --home $HOME_TMP"

if [[ -f $KEYS ]]; then echo "$KEYS exists; delete it to regenerate"; exit 1; fi

new_key() { # name -> json {name, mnemonic}
  local mnemonic; mnemonic=$(strided keys mnemonic 2>/dev/null)
  echo "$mnemonic" | $K add "$1" --recover >/dev/null 2>&1
  jq -n --arg n "$1" --arg m "$mnemonic" '{name:$n, mnemonic:$m}'
}
addr() { $K show "$1" -a; }
to_prefix() { # stride-addr prefix -> re-encoded
  python3 - "$1" "$2" <<'EOF'
import sys; sys.path.insert(0, "scripts/wind-down")
import bech32_ref
_, raw = bech32_ref.decode(sys.argv[1]); print(bech32_ref.encode(sys.argv[2], raw))
EOF
}
ibc_denom() { echo "ibc/$(printf '%s' "$1" | shasum -a 256 | cut -d' ' -f1 | tr a-f A-F)"; }

cd "$DIR/../.."   # repo root so scripts/wind-down is importable

m=$(jq -n "[$(new_key m1),$(new_key m2),$(new_key m3)]")
d=$(jq -n "[$(new_key d1),$(new_key d2),$(new_key d3)]")
sweep=$(new_key sweep-operator)
deposit=$(new_key st-deposit); redemption=$(new_key st-redemption); claim=$(new_key st-claim)
safe=$(new_key st-safe); operator=$(new_key st-operator); reward=$(new_key st-reward)
hbase=$(new_key holder-base); hvest=$(new_key holder-vesting)

$K add admin-ms --multisig m1,m2,m3 --multisig-threshold 2 >/dev/null
$K add hub-ms   --multisig d1,d2,d3 --multisig-threshold 2 >/dev/null

jq -n --argjson m "$m" --argjson d "$d" --argjson sweep "$sweep" \
  --argjson deposit "$deposit" --argjson redemption "$redemption" --argjson claim "$claim" \
  --argjson safe "$safe" --argjson operator "$operator" --argjson reward "$reward" \
  --argjson hbase "$hbase" --argjson hvest "$hvest" \
  '{multisig_members:$m, hub_multisig_members:$d, sweep_operator:$sweep,
    staketia:{deposit:$deposit, redemption:$redemption, claim:$claim, safe:$safe, operator:$operator, reward:$reward},
    holders:{base:$hbase, vesting:$hvest}}' > "$KEYS"

user1=$(jq -r '.users[0].mnemonic' integration-tests/network/configs/keys.json)
echo "$user1" | $K add user1 --recover >/dev/null 2>&1

admin_ms=$(addr admin-ms); hub_ms=$(addr hub-ms)
cat > "$OUT" <<EOF
ADMIN_MS_STRIDE=$admin_ms
VAULT_MS_OSMO=$(to_prefix $admin_ms osmo)
ADMIN_MS_MEMBERS=m1,m2,m3
HUB_MS_COSMOS=$(to_prefix $hub_ms cosmos)
HUB_MS_MEMBERS=d1,d2,d3
SWEEP_OPERATOR=$(addr sweep-operator)
STAKETIA_DEPOSIT=$(addr st-deposit)
STAKETIA_REDEMPTION=$(addr st-redemption)
STAKETIA_CLAIM=$(addr st-claim)
STAKETIA_SAFE=$(addr st-safe)
STAKETIA_OPERATOR_STRIDE=$(addr st-operator)
STAKETIA_OPERATOR_HUB=$(to_prefix $(addr st-operator) cosmos)
HUB_REWARD=$(to_prefix $(addr st-reward) cosmos)
HOLDER_BASE=$(addr holder-base)
HOLDER_BASE_OSMO=$(to_prefix $(addr holder-base) osmo)
HOLDER_VESTING=$(addr holder-vesting)
HOLDER_VESTING_OSMO=$(to_prefix $(addr holder-vesting) osmo)
USER1_STRIDE=$(addr user1)
USER1_OSMO=$(to_prefix $(addr user1) osmo)
USER1_COSMOS=$(to_prefix $(addr user1) cosmos)
ATOM_ON_STRIDE=$(ibc_denom transfer/channel-0/uatom)
OSMO_ON_STRIDE=$(ibc_denom transfer/channel-1/uosmo)
EOF
rm -rf "$HOME_TMP"
cat "$OUT"
```

- [ ] **Step 2: Run it and check the derived ATOM denom**

Run: `cd integration-tests/rehearsal && bash gen_keys.sh`
Expected: `ATOM_ON_STRIDE=ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2` and `VAULT_MS_OSMO` starts with `osmo1`. If the ATOM hash differs, `ibc_denom` is wrong; fix before continuing.

- [ ] **Step 3: Add three validator keys to the network keys file**

For each of `val6`, `val7`, `val8`: `mnemonic=$(strided keys mnemonic)` and append `{"name": "valN", "mnemonic": "<mnemonic>"}` to `.validators` in `integration-tests/network/configs/keys.json` with `jq`. Verify `jq '.validators | length' integration-tests/network/configs/keys.json` prints `8`.

- [ ] **Step 4: Commit**

```bash
git add integration-tests/rehearsal integration-tests/network/configs/keys.json
git commit -m "rehearsal: keys and derived addresses"
```

### Task 2: v35 rehearsal constants

**Files:**
- Modify: `utils/admins.go`
- Modify: `x/stakeibc/types/wind_down.go`
- Modify: `x/staketia/types/celestia.go`

**Interfaces:**
- Consumes: `integration-tests/rehearsal/addresses.env` (Task 1).
- Produces: a v35 binary whose wind-down code targets the k8s network.
- Review: yes (these are the values every later phase trusts)

- [ ] **Step 1: Admins**

`utils/admins.go` keeps the F5 line (build.sh rewrites it to the keys.json admin) and adds the multisig:

```go
var Admins = map[string]bool{
	"stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh": true, // F5 (rewritten to the k8s admin key by build.sh)
	"<ADMIN_MS_STRIDE>":                            true, // rehearsal 2-of-3 admin multisig
	"stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl": true, // gov module
}
```

- [ ] **Step 2: Wind-down constants**

In `x/stakeibc/types/wind_down.go`:

```go
var (
	SweepOperatorAddress = "<SWEEP_OPERATOR>"
	OsmosisVaultAddress  = "<VAULT_MS_OSMO>"
)

const (
	OsmosisChainId                   = "osmosis-test-1"
	OsmosisBech32Prefix              = "osmo"
	StrideToOsmosisTransferChannelId = "channel-1"
	WindDownTransferTimeout          = 15 * time.Minute
)

var HostToOsmosisTransferChannel = map[string]string{
	"cosmoshub-test-1": "channel-1",
	"osmosis-test-1":   "",
}

var SweepUnwindChannels = map[string]string{
	"channel-0": "cosmos",
	"channel-1": "osmo",
}
```

and the staketia copies: `StaketiaDepositAddress`, `StaketiaRedemptionAddress`, `StaketiaClaimAddress`, `StaketiaSafeAddress`, `StaketiaOperatorAddress` set to the `STAKETIA_*` values. Leave the stakedym constants.

- [ ] **Step 3: Staketia constants**

In `x/staketia/types/celestia.go`:

```go
CelestiaChainId                   = "cosmoshub-test-1"
StrideToCelestiaTransferChannelId = "channel-0"
CelestiaNativeTokenDenom          = "uatom"
CelestiaNativeTokenIBCDenom       = "<ATOM_ON_STRIDE>"
DelegationAddressOnCelestia       = "<HUB_MS_COSMOS>"
RewardAddressOnCelestia           = "<HUB_REWARD>"
DepositAddress                    = "<STAKETIA_DEPOSIT>"
RedemptionAddress                 = "<STAKETIA_REDEMPTION>"
ClaimAddress                      = "<STAKETIA_CLAIM>"
SafeAddressOnStride               = "<STAKETIA_SAFE>"
OperatorAddressOnStride           = "<STAKETIA_OPERATOR_STRIDE>"
CelestiaUnbondingPeriodSeconds    = uint64(240)
CelestiaBechPrefix                = "cosmos"
```

and `CelestiaConnectionId = "connection-0"`.

- [ ] **Step 4: Build and run the constant tests that still apply**

Run: `go build ./... && go test ./x/stakeibc/types/ -run 'WindDown|Sweep' -count=1`
Expected: build OK; the address-format tests pass. `app/upgrades/v35` mainnet-export tests are expected to fail on this branch and are not run.

- [ ] **Step 5: Commit**

```bash
git commit -am "rehearsal: point wind-down and staketia constants at the k8s network"
```

### Task 3: The v34 branch

**Files:**
- Create branch `wind-down-rehearsal-v34` from `v34.1.0`; modify `x/staketia/types/celestia.go` there.

**Interfaces:**
- Produces: a branch name `build.sh` can check out as `UPGRADE_OLD_VERSION`.
- Review: no

- [ ] **Step 1: Create the branch in a scratch worktree**

```bash
git worktree add /tmp/rehearsal-v34 -b wind-down-rehearsal-v34 v34.1.0
```

(A raw worktree is fine here: it is deleted at the end of this task and never navigated to.)

- [ ] **Step 2: Patch the same staketia constants as Task 2 step 3** in `/tmp/rehearsal-v34/x/staketia/types/celestia.go` (identical values). v34 reads `CelestiaChainId` for every stakeibc lookup, so this is what lets staketia run against the Hub before the upgrade.

- [ ] **Step 3: Build, commit, remove the scratch worktree**

```bash
(cd /tmp/rehearsal-v34 && go build ./... && git commit -am "rehearsal: staketia constants on v34")
git worktree remove /tmp/rehearsal-v34
git log --oneline -1 wind-down-rehearsal-v34
```

### Task 4: Network configuration

**Files:**
- Modify: `integration-tests/network/values.yaml`
- Modify: `integration-tests/network/configs/relayer.yaml`
- Modify: `integration-tests/network/scripts/config.sh`
- Modify: `integration-tests/network/scripts/init-chain.sh`
- Modify: `integration-tests/network/scripts/upgrade.sh`

**Interfaces:**
- Consumes: `addresses.env`, `integration-tests/rehearsal/keys.json` (Task 1).
- Produces: a helm chart that starts the rehearsal topology with staketia, slashing, wasm and cosmwasmpool genesis set.
- Review: yes (genesis mistakes cost a full restart)

- [ ] **Step 1: values.yaml**

```yaml
activeChains: [stride, cosmoshub, osmosis]
relayers:
  - {name: stride-cosmoshub, type: relayer, chainA: stride, chainB: cosmoshub}
  - {name: stride-osmosis,   type: relayer, chainA: stride, chainB: osmosis}
  - {name: cosmoshub-osmosis, type: relayer, chainA: cosmoshub, chainB: osmosis}
```

No hermes entries. `numValidators`: stride 4, cosmoshub 8, osmosis 3. Keep the image versions as they are.

- [ ] **Step 2: relayer.yaml** — add a path:

```yaml
  cosmoshub-osmosis:
    src:
      chain-id: cosmoshub-test-1
    dst:
      chain-id: osmosis-test-1
```

- [ ] **Step 3: config.sh** — add `STRIDE_HOUR_EPOCH_DURATION="30s"` and the rehearsal keys path `REHEARSAL_KEYS_FILE=${CONFIG_DIR}/rehearsal-keys.json`. Copy `integration-tests/rehearsal/keys.json` to `integration-tests/network/configs/rehearsal-keys.json` and `addresses.env` to `integration-tests/network/configs/rehearsal-addresses.env` (the configs dir is the configmap the pods see).

- [ ] **Step 4: init-chain.sh, Stride genesis** — in `update_stride_genesis` add:

```bash
    source configs/rehearsal-addresses.env
    jq_inplace '(.app_state.epochs.epochs[] | select(.identifier=="hour") ).duration |= "'$STRIDE_HOUR_EPOCH_DURATION'"' $genesis_json

    # staketia against the Hub (the Hub plays celestia). 50 ATOM already "delegated" by the Hub multisig.
    jq_inplace '.app_state.staketia.host_zone.chain_id = "cosmoshub-test-1"
      | .app_state.staketia.host_zone.native_token_denom = "uatom"
      | .app_state.staketia.host_zone.native_token_ibc_denom = "'$ATOM_ON_STRIDE'"
      | .app_state.staketia.host_zone.transfer_channel_id = "channel-0"
      | .app_state.staketia.host_zone.unbonding_period_seconds = "240"
      | .app_state.staketia.host_zone.delegation_address = "'$HUB_MS_COSMOS'"
      | .app_state.staketia.host_zone.reward_address = "'$HUB_REWARD'"
      | .app_state.staketia.host_zone.deposit_address = "'$STAKETIA_DEPOSIT'"
      | .app_state.staketia.host_zone.redemption_address = "'$STAKETIA_REDEMPTION'"
      | .app_state.staketia.host_zone.claim_address = "'$STAKETIA_CLAIM'"
      | .app_state.staketia.host_zone.safe_address_on_stride = "'$STAKETIA_SAFE'"
      | .app_state.staketia.host_zone.operator_address_on_stride = "'$STAKETIA_OPERATOR_STRIDE'"
      | .app_state.staketia.host_zone.remaining_delegated_balance = "50000000"' $genesis_json
```

and in `add_accounts` add the rehearsal accounts to the keyring and genesis on **every** chain (so each pod can sign with them): every entry of `.multisig_members[]`, `.hub_multisig_members[]`, `.sweep_operator`, `.staketia.operator`, `.staketia.reward`, `.holders.base` with `${USER_BALANCE}${DENOM}`; `.staketia.deposit`, `.staketia.redemption`, `.staketia.claim`, `.staketia.safe` and `.holders.vesting` with keyring only (`keys add --recover`, no genesis balance; the vesting account is created by tx in the seed). Read them from `$REHEARSAL_KEYS_FILE`.

- [ ] **Step 5: init-chain.sh, host genesis** — in `update_host_genesis`:

```bash
    # Downtime slashing fast and visible: 1% after ~30 missed blocks, on both hosts
    jq_inplace '.app_state.slashing.params.signed_blocks_window = "60"
      | .app_state.slashing.params.min_signed_per_window = "0.500000000000000000"
      | .app_state.slashing.params.downtime_jail_duration = "60s"
      | .app_state.slashing.params.slash_fraction_downtime = "0.010000000000000000"' $genesis_json

    if [[ "$CHAIN_NAME" == "osmosis" ]]; then
        jq_inplace '.app_state.wasm.params.code_upload_access.permission = "Everybody"
          | .app_state.wasm.params.instantiate_default_permission = "Everybody"
          | .app_state.cosmwasmpool.params.code_id_whitelist = ["1"]
          | .app_state.cosmwasmpool.params.pool_migration_limit = "20"
          | .app_state.poolmanager.params.pool_creation_fee = [{"denom":"uosmo","amount":"1000000"}]' $genesis_json
    fi
```

Keep the existing concentrated-liquidity line.

- [ ] **Step 6: upgrade.sh** — add a fourth voter `STRIDED3` (`stride-validator-3`, key `val4`) and accept an optional `UPGRADE_HEIGHT` env override in place of `latest+45` (the seed script sets the height to land 90s after a day epoch).

- [ ] **Step 7: Lint and commit**

Run: `cd integration-tests && helm lint network && bash -n network/scripts/init-chain.sh network/scripts/upgrade.sh`
Expected: lint OK, no syntax errors.

```bash
git add integration-tests/network
git commit -m "rehearsal: network topology, staketia/slashing/wasm genesis, hour epoch"
```

### Task 5: Transmuter bytecode

**Files:**
- Create: `integration-tests/rehearsal/transmuter_v3.2.0.wasm`

**Interfaces:**
- Produces: the exact mainnet bytecode of code id 996, uploaded in phase 7 as code id 1 on the test Osmosis.
- Review: no

- [ ] **Step 1: Download and verify the hash**

```bash
curl -s -A "curl/8.0" https://osmosis-api.polkachu.com/cosmwasm/wasm/v1/code/996 > /tmp/code996.json
jq -r '.data' /tmp/code996.json | base64 -d > integration-tests/rehearsal/transmuter_v3.2.0.wasm
echo "expected $(jq -r '.code_info.data_hash' /tmp/code996.json | tr A-F a-f)"
echo "actual   $(shasum -a 256 integration-tests/rehearsal/transmuter_v3.2.0.wasm | cut -d' ' -f1)"
```

Expected: the two hashes are equal. Commit the file (`git commit -m "rehearsal: transmuter 3.2.0 bytecode (mainnet code 996)"`).

### Task 6: Ops script tables for the k8s network

**Files:**
- Modify: `scripts/wind-down/measure_delegation_drift.py`
- Modify: `scripts/wind-down/build_sweep_batches.py`
- Modify: `scripts/wind-down/coverage_check.py`
- Modify: `scripts/wind-down/check_transmuter_pool.py`

**Interfaces:**
- Consumes: `addresses.env`. REST hosts: `https://stride-api.internal.stridenet.co`, `https://cosmoshub-api.internal.stridenet.co`, `https://osmosis-api.internal.stridenet.co` (verify each answers `/cosmos/base/tendermint/v1beta1/node_info` once the network is up; phase 0 does this).
- Produces: scripts that run against the rehearsal network with their logic unchanged.
- Review: no (tables only; a wrong table shows up in phase 0/7/8 immediately)

- [ ] **Step 1: drift script** — `STRIDE_REST` → the k8s Stride host; `ZONES` → `{"cosmoshub-test-1": {"registry": None, "decimals": 6}, "osmosis-test-1": {"registry": None, "decimals": 6}}`; `EXTRA_ENDPOINTS` → the two k8s host REST URLs; make the chain-registry fetch skip when `registry` is `None`. Do not change the comparison logic.
- [ ] **Step 2: sweep builder** — `UNWIND_CHANNELS` → `{"channel-0": "cosmos", "channel-1": "osmo"}`; the protocol address list → the `STAKETIA_*` rehearsal addresses (keep the stakedym constants); `DEFAULT_SWEEP_OPERATOR` → `SWEEP_OPERATOR`. Run `python3 -m pytest scripts/wind-down/test_build_sweep_batches.py -q`; fix only the tests that assert the swapped tables.
- [ ] **Step 3: coverage check** — `REQUIRED_ROUTES` → `{"stuatom": frozenset({"channel-0"}), "stuosmo": frozenset()}`; `OSMOSIS_REST_DEFAULT` → the k8s Osmosis host. Run its tests; fix only the tests that assert the policy.
- [ ] **Step 4: pool check** — constants block: both REST defaults, `TRANSMUTER_CODE_ID = "1"`, `COSMWASMPOOL_MODULE` → the module account printed by `osmosisd q auth module-account cosmwasmpool` (fill in during phase 7 if unknown now; leave a `REHEARSAL_FILL_IN` marker the phase 7 script greps for and refuses to run past), `STRIDE_TO_OSMOSIS_CHANNEL_ON_OSMOSIS = "channel-0"`, `ADMIN = MODERATOR = VAULT_MS_OSMO`, `HOST_TO_OSMOSIS_CHANNEL` → `{"cosmoshub-test-1": "channel-1"}`, `POOLS` → empty (phase 7 appends the three pool specs once the ids exist).
- [ ] **Step 5: Commit** — `git commit -am "rehearsal: ops script tables for the k8s network"`.

### Task 7: The driver library

**Files:**
- Create: `integration-tests/rehearsal/lib.sh`
- Create: `docs/wind-down/rehearsal-log.md` (header only)

**Interfaces:**
- Produces, for every phase script: `KX` (kubectl prefix), `strided_old`, `strided_new`, `strided`, `gaiad`, `osmosisd` (exec wrappers; each takes CLI args), `STRIDE_TX`, `HUB_TX`, `OSMO_TX` (common tx flags), `log <text>`, `log_cmd <label> <cmd...>` (runs, prints and appends a fenced block), `wait_tx <chain> <hash>` (polls `q tx`, prints `code` and `raw_log`, returns non-zero on code ≠ 0), `wait_until <timeout-s> <description> <cmd...>` (polls every 5s until the command exits 0), `day_epoch_next_start` (unix seconds, from `q stakeibc show-epoch-tracker day`), `sleep_until <unix-seconds>`, `ms_tx <chain> <multisig-name> <members-csv> -- <tx args...>` (generate-only → sign with the first two members in amino-json → multisign → broadcast → `wait_tx`), `checkpoint <name> <cmd...>` (runs an assertion command, logs PASS/FAIL, exits on FAIL unless `CHECKPOINT_SOFT=1`), `rate_of <chain-id>` (prints `HostZone.RedemptionRate`), `assert_rate_unchanged`.
- Review: yes (every phase trusts it)

- [ ] **Step 1: Write lib.sh**

```bash
#!/bin/bash
# Shared driver for the wind-down rehearsal (spec docs/superpowers/specs/2026-10-02-wind-down-rehearsal-design.md).
set -euo pipefail
REHEARSAL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO="$(cd "$REHEARSAL_DIR/../.." && pwd)"
LOG=$REPO/docs/wind-down/rehearsal-log.md
source "$REHEARSAL_DIR/addresses.env"

[[ "$(kubectl config current-context)" == "integration" ]] || { echo "kube context must be 'integration'"; exit 1; }
KX="kubectl -n integration"
NEW_BIN=/home/validator/.stride/cosmovisor/upgrades/v35/bin/strided

strided_old() { $KX exec stride-validator-0 -c validator -- strided "$@"; }
strided_new() { $KX exec stride-validator-0 -c validator -- $NEW_BIN "$@"; }
strided() { if strided_old version 2>/dev/null | grep -q '^v34'; then strided_old "$@"; else strided_new "$@"; fi; }
strided_pod() { local pod=$1; shift; $KX exec $pod -c validator -- strided "$@"; }
gaiad()    { $KX exec cosmoshub-validator-0 -c validator -- gaiad "$@"; }
osmosisd() { $KX exec osmosis-validator-0 -c validator -- osmosisd "$@"; }

STRIDE_TX="--keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json"
HUB_TX="--keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json"
OSMO_TX="--keyring-backend test --chain-id osmosis-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 0.04uosmo -y -o json"

log() { printf '%s %s\n' "$(date -u +%FT%TZ)" "$*" | tee -a "$LOG"; }
log_cmd() { # label, then the command
  local label=$1; shift
  log "### $label"; printf '```\n$ %s\n' "$*" >> "$LOG"
  local out; out=$("$@" 2>&1) || true
  printf '%s\n```\n' "$out" >> "$LOG"; printf '%s\n' "$out"
}
tx_hash() { jq -r '.txhash'; }
wait_tx() { # chain hash
  local chain=$1 hash=$2 res
  for _ in $(seq 1 30); do
    res=$($chain q tx "$hash" -o json 2>/dev/null) && break; sleep 2
  done
  [[ -n "${res:-}" ]] || { log "tx $hash not found on $chain"; return 1; }
  local code; code=$(jq -r '.code' <<<"$res")
  log "tx $hash code=$code $(jq -r '.raw_log' <<<"$res" | head -c 300)"
  [[ "$code" == "0" ]]
}
wait_until() { # timeout-s description cmd...
  local timeout=$1 desc=$2; shift 2; local start=$SECONDS
  until "$@" >/dev/null 2>&1; do
    (( SECONDS - start < timeout )) || { log "TIMEOUT waiting for: $desc"; return 1; }; sleep 5
  done; log "ready: $desc"
}
day_epoch_next_start() { strided q stakeibc show-epoch-tracker day -o json | jq -r '.epoch_tracker.next_epoch_start_time' | cut -c1-10; }
# next_epoch_start_time is nanoseconds; cut keeps the seconds
sleep_until() { local now; now=$(date +%s); (( $1 > now )) && sleep $(( $1 - now )) || true; }
ms_tx() { # chain multisig-name members-csv -- tx args...
  local chain=$1 ms=$2 members=$3; shift 4
  local m1=${members%%,*} rest=${members#*,} m2=${rest%%,*} chainid gasprice bin pod
  case $chain in
    strided_new)         chainid=stride-test-1;    gasprice=1ustrd;    bin=$NEW_BIN; pod=stride-validator-0;;
    strided_old|strided) chainid=stride-test-1;    gasprice=1ustrd;    bin=strided;  pod=stride-validator-0;;
    gaiad)               chainid=cosmoshub-test-1; gasprice=1uatom;    bin=gaiad;    pod=cosmoshub-validator-0;;
    osmosisd)            chainid=osmosis-test-1;   gasprice=0.04uosmo; bin=osmosisd; pod=osmosis-validator-0;;
  esac
  # The pod runs the whole generate / sign / multisign / broadcast pipeline so no file leaves the container
  local args; args=$(printf ' %q' "$@")
  local hash; hash=$($KX exec $pod -c validator -- sh -c "
    set -e
    $bin tx $args --from $ms --generate-only --keyring-backend test --chain-id $chainid --gas 600000 --gas-prices $gasprice > /tmp/unsigned.json
    $bin tx sign /tmp/unsigned.json --from $m1 --multisig $ms --sign-mode amino-json --keyring-backend test --chain-id $chainid --output-document /tmp/s1.json
    $bin tx sign /tmp/unsigned.json --from $m2 --multisig $ms --sign-mode amino-json --keyring-backend test --chain-id $chainid --output-document /tmp/s2.json
    $bin tx multisign /tmp/unsigned.json $ms /tmp/s1.json /tmp/s2.json --keyring-backend test --chain-id $chainid --output-document /tmp/signed.json
    $bin tx broadcast /tmp/signed.json --chain-id $chainid -o json" | tee -a "$LOG" | tx_hash)
  echo "$hash"; wait_tx $chain "$hash"
}
checkpoint() { # name cmd...
  local name=$1; shift
  if "$@"; then log "CHECKPOINT PASS: $name"; else log "CHECKPOINT FAIL: $name"; [[ "${CHECKPOINT_SOFT:-0}" == 1 ]] || exit 1; fi
}
rate_of() { strided q stakeibc show-host-zone "$1" -o json | jq -r '.host_zone.redemption_rate'; }
assert_rate_unchanged() { # chain-id expected
  [[ "$(rate_of "$1")" == "$2" ]]
}
```

`ms_tx` prints exactly the tx hash on stdout and returns `wait_tx`'s status. Add `export -f strided_old strided_new strided strided_pod gaiad osmosisd tx_hash wait_tx; export KX NEW_BIN LOG` at the end so `sh -c` callers see the wrappers.

- [ ] **Step 2: Log header**

`docs/wind-down/rehearsal-log.md`:

```markdown
# Wind-down rehearsal log

Network: k8s `integration` (stride-test-1 4 vals, cosmoshub-test-1 8 vals, osmosis-test-1 3 vals).
Branches: `wind-down-rehearsal` (v35), `wind-down-rehearsal-v34` (old binary).
Spec: docs/superpowers/specs/2026-10-02-wind-down-rehearsal-design.md. Every entry below is appended by the phase scripts.
```

- [ ] **Step 3: Syntax-check and smoke the wrappers against the live network once it is up** (Task 14 does the live check); for now `bash -n integration-tests/rehearsal/lib.sh` and commit:

```bash
git add integration-tests/rehearsal/lib.sh docs/wind-down/rehearsal-log.md
git commit -m "rehearsal: driver library and log"
```

## Parallel-safe tasks

Each writes one or two phase scripts under `integration-tests/rehearsal/`, sources `lib.sh`, and uses only the interfaces of Tasks 1 and 7. None depends on another parallel task. Each script starts with `source "$(dirname "$0")/lib.sh"` and `log "## Phase N: <name>"`.

### Task 8: `seed.sh` (pre-upgrade state on v34) and `upgrade.sh` (phase 1)

**Files:** Create `integration-tests/rehearsal/seed.sh`, `integration-tests/rehearsal/phase1_upgrade.sh`.
**Depends on:** Tasks 1-7. **Review:** yes (state machine timing)

- [ ] **Step 1: seed.sh part A, accounts and zones**

```bash
source "$(dirname "$0")/lib.sh"; log "## Seed (v34)"
# Multisigs into the pod keyrings (members were restored by init-chain)
strided_old keys add admin-ms --multisig m1,m2,m3 --multisig-threshold 2 --keyring-backend test || true
osmosisd    keys add vault-ms --multisig m1,m2,m3 --multisig-threshold 2 --keyring-backend test || true
gaiad       keys add hub-ms   --multisig d1,d2,d3 --multisig-threshold 2 --keyring-backend test || true
# Fund the multisigs and the sweep operator from the faucet
log_cmd "fund admin-ms" strided_old tx bank send faucet $ADMIN_MS_STRIDE 1000000000ustrd $STRIDE_TX
log_cmd "fund vault-ms" osmosisd tx bank send faucet $VAULT_MS_OSMO 100000000uosmo $OSMO_TX
log_cmd "fund hub-ms"   gaiad tx bank send faucet $HUB_MS_COSMOS 100000000uatom $HUB_TX
sleep 6
# Validators of each host, ordered val1..valN by moniker
HUB_VALS=$(gaiad q staking validators -o json | jq -r '.validators | sort_by(.description.moniker) | .[].operator_address')
OSMO_VALS=$(osmosisd q staking validators -o json | jq -r '.validators | sort_by(.description.moniker) | .[].operator_address')
checkpoint "8 hub validators"    test "$(wc -w <<<"$HUB_VALS")" -eq 8
checkpoint "3 osmosis validators" test "$(wc -w <<<"$OSMO_VALS")" -eq 3
# Register the two zones (1-day cadence; 3 messages per ICA on the Hub so 8 validators make 3 batches)
log_cmd "register hub zone" strided_old tx stakeibc register-host-zone connection-0 uatom cosmos $ATOM_ON_STRIDE channel-0 1 false --max-messages-per-ica-tx 3 --from admin $STRIDE_TX
log_cmd "register osmo zone" strided_old tx stakeibc register-host-zone connection-1 uosmo osmo $OSMO_ON_STRIDE channel-1 1 false --from admin $STRIDE_TX
vals_file() { jq -n --arg list "$1" '{validators: ($list | split(" ") | to_entries | map({name: ("val"+((.key+1)|tostring)), address: .value, weight: 10}))}'; }
$KX exec stride-validator-0 -c validator -- sh -c "cat > /tmp/hub_vals.json" <<<"$(vals_file "$(echo $HUB_VALS)")"
$KX exec stride-validator-0 -c validator -- sh -c "cat > /tmp/osmo_vals.json" <<<"$(vals_file "$(echo $OSMO_VALS)")"
log_cmd "add hub validators"  strided_old tx stakeibc add-validators cosmoshub-test-1 /tmp/hub_vals.json --from admin $STRIDE_TX
log_cmd "add osmo validators" strided_old tx stakeibc add-validators osmosis-test-1 /tmp/osmo_vals.json --from admin $STRIDE_TX
wait_until 300 "hub delegation ICA" sh -c "strided_old q stakeibc show-host-zone cosmoshub-test-1 -o json | jq -e '.host_zone.delegation_ica_address != \"\"'"
wait_until 300 "osmo delegation ICA" sh -c "strided_old q stakeibc show-host-zone osmosis-test-1 -o json | jq -e '.host_zone.delegation_ica_address != \"\"'"
```

(`wait_until` with `sh -c` needs the functions exported: add `export -f strided_old strided_new strided gaiad osmosisd; export KX NEW_BIN` at the end of `lib.sh`.)

- [ ] **Step 2: seed.sh part B, tokens on Stride, liquid stakes, holders**

```bash
log_cmd "atom to stride" gaiad tx ibc-transfer transfer transfer channel-0 $USER1_STRIDE 2000000000uatom --from user1 $HUB_TX
log_cmd "osmo to stride" osmosisd tx ibc-transfer transfer transfer channel-0 $USER1_STRIDE 1000000000uosmo --from user1 $OSMO_TX
wait_until 120 "atom on stride" sh -c "strided_old q bank balances $USER1_STRIDE -o json | jq -e '.balances[] | select(.denom==\"$ATOM_ON_STRIDE\")'"
log_cmd "liquid stake 1000 ATOM" strided_old tx stakeibc liquid-stake 1000000000 uatom --from user1 $STRIDE_TX
log_cmd "liquid stake 300 OSMO"  strided_old tx stakeibc liquid-stake 300000000 uosmo --from user1 $STRIDE_TX
wait_until 400 "hub delegated" sh -c "strided_old q stakeibc show-host-zone cosmoshub-test-1 -o json | jq -e '.host_zone.total_delegations|tonumber > 900000000'"
wait_until 400 "osmo delegated" sh -c "strided_old q stakeibc show-host-zone osmosis-test-1 -o json | jq -e '.host_zone.total_delegations|tonumber > 250000000'"
# Holders: base, vesting (delayed, 1 year), distribution module (fund-community-pool), escrow (via IBC out)
log_cmd "holder base"   strided_old tx bank send user1 $HOLDER_BASE 50000000stuatom,10000000ustrd,20000000$ATOM_ON_STRIDE $STRIDE_TX
log_cmd "vesting acct"  strided_old tx vesting create-vesting-account $HOLDER_VESTING 1000000ustrd $(( $(date +%s) + 31536000 )) --from faucet $STRIDE_TX
sleep 6
log_cmd "holder vesting" strided_old tx bank send user1 $HOLDER_VESTING 30000000stuatom,5000000ustrd $STRIDE_TX
log_cmd "distribution holds stATOM" strided_old tx distribution fund-community-pool 5000000stuatom --from user1 $STRIDE_TX
# stATOM on Osmosis (canonical), on the Hub, and Hub -> Osmosis (two-hop)
log_cmd "statom to osmosis" strided_old tx ibc-transfer transfer transfer channel-1 $USER1_OSMO 100000000stuatom --from user1 $STRIDE_TX
log_cmd "statom to hub"     strided_old tx ibc-transfer transfer transfer channel-0 $USER1_COSMOS 60000000stuatom --from user1 $STRIDE_TX
log_cmd "stosmo to osmosis" strided_old tx ibc-transfer transfer transfer channel-1 $USER1_OSMO 50000000stuosmo --from user1 $STRIDE_TX
STATOM_ON_HUB=ibc/$(printf 'transfer/channel-0/stuatom' | shasum -a 256 | cut -d' ' -f1 | tr a-f A-F)
wait_until 120 "statom on hub" sh -c "gaiad q bank balances $USER1_COSMOS -o json | jq -e '.balances[] | select(.denom==\"$STATOM_ON_HUB\")'"
log_cmd "statom hub->osmosis (two-hop)" gaiad tx ibc-transfer transfer transfer channel-1 $USER1_OSMO 30000000$STATOM_ON_HUB --from user1 $HUB_TX
# Balances the transfer tx will move later
HUB_ZONE=$(strided_old q stakeibc show-host-zone cosmoshub-test-1 -o json)
log_cmd "fund hub fee ICA"        gaiad tx bank send user1 $(jq -r .host_zone.fee_ica_address <<<"$HUB_ZONE") 3000000uatom $HUB_TX
log_cmd "fund hub withdrawal ICA" gaiad tx bank send user1 $(jq -r .host_zone.withdrawal_ica_address <<<"$HUB_ZONE") 4000000uatom $HUB_TX
OSMO_ZONE=$(strided_old q stakeibc show-host-zone osmosis-test-1 -o json)
log_cmd "fund osmo fee ICA"        osmosisd tx bank send user1 $(jq -r .host_zone.fee_ica_address <<<"$OSMO_ZONE") 3000000uosmo $OSMO_TX
log_cmd "fund osmo withdrawal ICA" osmosisd tx bank send user1 $(jq -r .host_zone.withdrawal_ica_address <<<"$OSMO_ZONE") 4000000uosmo $OSMO_TX
```

- [ ] **Step 3: seed.sh part C, staketia and the rate limit**

```bash
# The Hub multisig delegates the 50 ATOM genesis says it holds, and grants the operator the mainnet authz set
HUB_VAL1=$(head -1 <<<"$HUB_VALS")
ms_tx gaiad hub-ms d1,d2,d3 -- staking delegate $HUB_VAL1 50000000uatom
ms_tx gaiad hub-ms d1,d2,d3 -- authz grant $STAKETIA_OPERATOR_HUB unbond --allowed-validators $(echo $HUB_VALS | tr ' ' ,)
ms_tx gaiad hub-ms d1,d2,d3 -- authz grant $STAKETIA_OPERATOR_HUB delegate --allowed-validators $(echo $HUB_VALS | tr ' ' ,)
ms_tx gaiad hub-ms d1,d2,d3 -- authz grant $STAKETIA_OPERATOR_HUB generic --msg-type /cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward
cat > /tmp/transfer_grant.json <<EOF
{"body":{"messages":[{"@type":"/cosmos.authz.v1beta1.MsgGrant","granter":"$HUB_MS_COSMOS","grantee":"$STAKETIA_OPERATOR_HUB",
 "grant":{"authorization":{"@type":"/ibc.applications.transfer.v1.TransferAuthorization","allocations":[{"source_port":"transfer","source_channel":"channel-0",
 "spend_limit":[{"denom":"uatom","amount":"1000000000000"}],"allow_list":["$STAKETIA_CLAIM"],"allowed_packet_data":[]}]},"expiration":null}}],
 "memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},
 "auth_info":{"signer_infos":[],"fee":{"amount":[{"denom":"uatom","amount":"600000"}],"gas_limit":"600000","payer":"","granter":""}},"signatures":[]}
EOF
$KX cp /tmp/transfer_grant.json cosmoshub-validator-0:/tmp/unsigned.json -c validator
$KX exec cosmoshub-validator-0 -c validator -- sh -c "set -e
  gaiad tx sign /tmp/unsigned.json --from d1 --multisig hub-ms --sign-mode amino-json --keyring-backend test --chain-id cosmoshub-test-1 --output-document /tmp/s1.json
  gaiad tx sign /tmp/unsigned.json --from d2 --multisig hub-ms --sign-mode amino-json --keyring-backend test --chain-id cosmoshub-test-1 --output-document /tmp/s2.json
  gaiad tx multisign /tmp/unsigned.json hub-ms /tmp/s1.json /tmp/s2.json --keyring-backend test --chain-id cosmoshub-test-1 --output-document /tmp/signed.json
  gaiad tx broadcast /tmp/signed.json -o json" | tee -a "$LOG"
checkpoint "transfer grant present" sh -c "gaiad q authz grants $HUB_MS_COSMOS $STAKETIA_OPERATOR_HUB -o json | jq -e '.grants[] | select(.authorization[\"@type\"]==\"/ibc.applications.transfer.v1.TransferAuthorization\")'"
# Rate limit on stATOM over the Osmosis channel: gov-only, so a proposal voted by the four validators
cat > /tmp/ratelimit.json <<EOF
{"title":"stATOM rate limit","summary":"rehearsal","metadata":"","deposit":"2000000000ustrd","messages":[{"@type":"/ibc.applications.rate_limiting.v1.MsgAddRateLimit",
 "signer":"stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl","denom":"stuatom","channel_or_client_id":"channel-1","max_percent_send":"10","max_percent_recv":"10","duration_hours":"24"}]}
EOF
$KX cp /tmp/ratelimit.json stride-validator-0:/tmp/ratelimit.json -c validator
log_cmd "rate limit proposal" strided_old tx gov submit-proposal /tmp/ratelimit.json --from val1 $STRIDE_TX
sleep 4; PROP=$(strided_old q gov proposals -o json | jq -r '.proposals | max_by(.id|tonumber).id')
for i in 0 1 2 3; do strided_pod stride-validator-$i tx gov vote $PROP yes --from val$((i+1)) $STRIDE_TX >/dev/null; done
wait_until 90 "rate limit live" sh -c "strided_old q ratelimit list-rate-limits -o json | jq -e '.rate_limits | length > 0'"
```

If the ratelimit msg type URL differs on this ibc-go, read it from `strided q ratelimit --help` / the proto under `deps/` and fix the JSON; log the correction.

- [ ] **Step 4: seed.sh part D, the redemption timeline**

Record epoch boundaries `D` from `day_epoch_next_start` and schedule relative to them. `U` (upgrade time) is `D_n + 90s`, with `n` chosen so that the whole timeline fits: the script computes `D0 = day_epoch_next_start` and sets `D1=D0+180, D2=D0+360, D3=D0+540, D4=D0+720` (`D4` is `D_n`; `U=D4+90`).

```bash
D0=$(day_epoch_next_start); D1=$((D0+180)); D2=$((D0+360)); D3=$((D0+540)); D4=$((D0+720)); U=$((D4+90))
log "day epochs: D0=$D0 D1=$D1 D2=$D2 D3=$D3 D4=$D4 upgrade target U=$U"
# Osmosis zone: val3 -> weight 0, then slashed by downtime, then a redemption that must land in UNBONDING_RETRY_QUEUE
OSMO_VAL3=$(sed -n 3p <<<"$OSMO_VALS")
log_cmd "osmo val3 weight 0" strided_old tx stakeibc change-validator-weight osmosis-test-1 $OSMO_VAL3 0 --from admin $STRIDE_TX
log_cmd "stop signing osmosis-validator-2" $KX exec osmosis-validator-2 -c validator -- sh -c 'kill -STOP $(pidof osmosisd)'
wait_until 240 "osmo val3 jailed" sh -c "osmosisd q staking validator $OSMO_VAL3 -o json | jq -e '.validator.jailed == true'"
log_cmd "resume osmosis-validator-2" $KX exec osmosis-validator-2 -c validator -- sh -c 'kill -CONT $(pidof osmosisd)'
log_cmd "osmo RE redeem (retry)" strided_old tx stakeibc redeem-stake 20000000 osmosis-test-1 $USER1_OSMO --from user1 $STRIDE_TX
# Osmosis zone RA': a clean record from before the slash is impossible now, so the osmo CLAIMABLE case is skipped; the Hub covers it.
# Hub: RA -> CLAIMABLE (submitted at D0 or D1), RB -> best-effort EXIT_TRANSFER_QUEUE (submitted D3), RC -> IN_PROGRESS (submitted D4), RD -> QUEUE (after D4)
log_cmd "hub RA redeem" strided_old tx stakeibc redeem-stake 30000000 cosmoshub-test-1 $USER1_COSMOS --from user1 $STRIDE_TX
sleep_until $((D2+10))
log_cmd "hub RB redeem" strided_old tx stakeibc redeem-stake 20000000 cosmoshub-test-1 $USER1_COSMOS --from user1 $STRIDE_TX
sleep_until $((D3+10))
log_cmd "hub RC redeem" strided_old tx stakeibc redeem-stake 20000000 cosmoshub-test-1 $USER1_COSMOS --from user1 $STRIDE_TX
# Staketia: two redemptions and a spillover, confirmed per the v34 operator flow (PrepareUndelegation runs every 4th day epoch)
log_cmd "staketia R1" strided_old tx staketia redeem-stake 20000000 $USER1_STRIDE --from user1 $STRIDE_TX
sleep_until $((D4+10))
log_cmd "hub RD redeem (queue)" strided_old tx stakeibc redeem-stake 10000000 cosmoshub-test-1 $USER1_COSMOS --from user1 $STRIDE_TX
log_cmd "staketia R2" strided_old tx staketia redeem-stake 20000000 $USER1_STRIDE --from user1 $STRIDE_TX
log_cmd "staketia R3 spillover" strided_old tx staketia redeem-stake 40000000 $USER1_STRIDE --from user1 $STRIDE_TX
# Pre-upgrade liquid stake so a deposit record is in flight at U
sleep_until $((U-40))
log_cmd "deposit in flight" strided_old tx stakeibc liquid-stake 10000000 uatom --from user1 $STRIDE_TX
HIST_TX=$(strided_old q txs --query "message.action='/stride.stakeibc.MsgLiquidStake'" -o json 2>/dev/null | jq -r '.txs[0].txhash' || true)
echo "HIST_TX=$HIST_TX" >> "$REHEARSAL_DIR/state.env"
log "pre-upgrade rates: hub=$(rate_of cosmoshub-test-1) osmo=$(rate_of osmosis-test-1)"
printf 'RATE_HUB=%s\nRATE_OSMO=%s\nU=%s\n' "$(rate_of cosmoshub-test-1)" "$(rate_of osmosis-test-1)" "$U" >> "$REHEARSAL_DIR/state.env"
log_cmd "records at upgrade" strided_old q records list-epoch-unbonding-record -o json
log_cmd "staketia records at upgrade" strided_old q staketia unbonding-records -o json
```

The staketia `UNBONDING_IN_PROGRESS` record needs the operator to have confirmed an undelegation before `U`; because `PrepareUndelegation` runs only every 4th day epoch (12 minutes), the seed waits for the first prepare after R1 (`wait_until 800 ... status UNBONDING_QUEUE`), then runs the v34 operator flow: `gaiad tx authz exec` of a `MsgUndelegate` for the record's `native_amount` from `hub-ms` to `$HUB_VAL1` signed by `st-operator`, then `strided_old tx staketia confirm-undelegation <id> <hub-tx-hash> --from st-operator`. Put this between R1 and the `sleep_until $((D4+10))` and shift `D4`/`U` later by 720s if R1's prepare has not happened by `D3`.

- [ ] **Step 5: phase1_upgrade.sh**

```bash
source "$(dirname "$0")/lib.sh"; source "$REHEARSAL_DIR/state.env"; log "## Phase 1: upgrade to v35"
height=$(strided_old status | jq -r '.sync_info.latest_block_height // .SyncInfo.latest_block_height')
now=$(date +%s); UPGRADE_HEIGHT=$(( height + (U - now) ))
log "upgrade height $UPGRADE_HEIGHT (now $height, target time $U)"
UPGRADE_HEIGHT=$UPGRADE_HEIGHT bash "$REPO/integration-tests/network/scripts/upgrade.sh" | tee -a "$LOG"
wait_until 300 "v35 running" sh -c "strided_new status >/dev/null 2>&1 && strided_new q upgrade applied v35 -o json | jq -e '.height'"
log_cmd "handler log lines" $KX logs stride-validator-0 -c validator --since=10m | grep -E 'v35|wind-down|Upgrade v35'
checkpoint "no handler error" sh -c "! $KX logs stride-validator-0 -c validator --since=10m | grep -E 'v35.*(error|ERR|panic)'"
checkpoint "liquid-stake cannot route"  sh -c "strided_new tx stakeibc liquid-stake 1000 uatom --from user1 $STRIDE_TX 2>&1 | grep -qi \"can't route\\|unknown command\\|not found\""
checkpoint "autopilot stakeibc off"    sh -c "strided_new q autopilot params -o json | jq -e '.params.stakeibc_active == false'"
checkpoint "rate limits removed"       sh -c "strided_new q ratelimit list-rate-limits -o json | jq -e '.rate_limits | length == 0'"
checkpoint "wasm upload gov-only"      sh -c "strided_new q wasm params -o json | jq -e '.code_upload_access.permission == \"AnyOfAddresses\" and (.code_upload_access.addresses == [\"stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl\"])'"
checkpoint "ica host allow-list trimmed" sh -c "strided_new q interchain-accounts host params -o json | jq -e '.params.allow_messages | index(\"/stride.stakeibc.MsgLiquidStake\") == null'"
checkpoint "historical tx decodes"     sh -c "strided_new q tx $HIST_TX -o json | jq -e '.txhash'"
checkpoint "hub rate frozen"  assert_rate_unchanged cosmoshub-test-1 $RATE_HUB
checkpoint "osmo rate frozen" assert_rate_unchanged osmosis-test-1 $RATE_OSMO
log_cmd "records after upgrade" strided_new q records list-epoch-unbonding-record -o json
```

The autopilot "through autopilot" and "through the ICA host" route checks are an ICS-20 memo transfer from the Hub with a liquid-stake memo (expect refund) and are added to phase 1 as `CHECKPOINT_SOFT=1` checks; the exact memo shape is in `x/autopilot/types/parser.go`.

- [ ] **Step 6: Syntax-check and commit** — `bash -n` both files; `git add integration-tests/rehearsal && git commit -m "rehearsal: seed and upgrade phase scripts"`.

### Task 9: `phase2_day0.sh`, `phase3_redemptions.sh`, `phase4_drain.sh`

**Files:** Create the three scripts. **Depends on:** Tasks 1-7. **Review:** yes (money paths)

- [ ] **Step 1: phase2_day0.sh**

```bash
source "$(dirname "$0")/lib.sh"; source "$REHEARSAL_DIR/state.env"; log "## Phase 2: day 0"
HUB_VALS=$(strided_new q stakeibc show-validators cosmoshub-test-1 -o json | jq -r '.validators[].address')
OSMO_VALS=$(strided_new q stakeibc show-validators osmosis-test-1 -o json | jq -r '.validators[].address')
checkpoint "non-admin refresh rejected" sh -c "strided_new tx stakeibc update-delegation cosmoshub-test-1 $(head -1 <<<"$HUB_VALS") --from user1 $STRIDE_TX 2>&1 | grep -qi 'invalid admin\\|unauthorized\\|not an admin'"
for v in $HUB_VALS;  do ms_tx strided_new admin-ms m1,m2,m3 -- stakeibc update-delegation cosmoshub-test-1 $v; done
for v in $OSMO_VALS; do ms_tx strided_new admin-ms m1,m2,m3 -- stakeibc update-delegation osmosis-test-1 $v; done
wait_until 300 "no slash query in flight" sh -c "strided_new q stakeibc show-validators osmosis-test-1 -o json | jq -e '[.validators[].slash_query_in_progress] | all(. == false)'"
log_cmd "osmo validators after refresh" strided_new q stakeibc show-validators osmosis-test-1 -o json
checkpoint "osmo rate still frozen" assert_rate_unchanged osmosis-test-1 $RATE_OSMO
checkpoint "drain refused while records queued (hub)" sh -c "strided_new tx stakeibc undelegate-from-validators cosmoshub-test-1 --all --from admin $STRIDE_TX | jq -e '.code != 0'"
checkpoint "drain refused while retry record (osmo)"  sh -c "strided_new tx stakeibc undelegate-from-validators osmosis-test-1 --all --from admin $STRIDE_TX | jq -e '.code != 0'"
log_cmd "drift" python3 "$REPO/scripts/wind-down/measure_delegation_drift.py" --chain-id cosmoshub-test-1 --chain-id osmosis-test-1
checkpoint "zero over-recorded" sh -c "jq -e '[.[] | select(.over_recorded == true)] | length == 0' $REPO/scripts/wind-down/drift/drift.json"
# Staketia day 0: the operator undelegates the whole multisig delegation via authz and confirms the queued record(s)
HUB_VAL1=$(head -1 <<<"$HUB_VALS"); DEL=$(gaiad q staking delegation $HUB_MS_COSMOS $HUB_VAL1 -o json | jq -r '.delegation_response.balance.amount')
cat > /tmp/unbond.json <<EOF
{"body":{"messages":[{"@type":"/cosmos.staking.v1beta1.MsgUndelegate","delegator_address":"$HUB_MS_COSMOS","validator_address":"$HUB_VAL1","amount":{"denom":"uatom","amount":"$DEL"}}]}}
EOF
$KX cp /tmp/unbond.json cosmoshub-validator-0:/tmp/unbond.json -c validator
H=$(gaiad tx authz exec /tmp/unbond.json --from st-operator $HUB_TX | tx_hash); wait_tx gaiad $H
for id in $(strided_new q staketia unbonding-records -o json | jq -r '.unbonding_records[] | select(.status=="UNBONDING_QUEUE") | .id'); do
  log_cmd "confirm-undelegation $id" strided_new tx staketia confirm-undelegation $id $H --from st-operator $STRIDE_TX
done
```

The drift JSON field name for "over-recorded" is whatever `measure_delegation_drift.py` writes; read it from the script's `report` section and use that key.

- [ ] **Step 2: phase3_redemptions.sh**

```bash
source "$(dirname "$0")/lib.sh"; source "$REHEARSAL_DIR/state.env"; log "## Phase 3: last redemption cycle"
not_terminal() { strided_new q records list-epoch-unbonding-record -o json | jq -e '[.epoch_unbonding_record[].host_zone_unbondings[] | select((.native_token_amount|tonumber) > 0 and .status != "CLAIMABLE")] | length == 0'; }
wait_until 900 "every hub/osmo unbonding CLAIMABLE" not_terminal
claim_all() {
  strided_new q records list-user-redemption-record -o json | jq -r '.user_redemption_record[] | select(.claim_is_pending==false) | "\(.host_zone_id) \(.epoch_number) \(.receiver)"' | while read z e r; do
    log_cmd "claim $z $e" strided_new tx stakeibc claim-undelegated-tokens $z $e $r --from user1 $STRIDE_TX
  done
}
claim_all; sleep 60; claim_all
wait_until 300 "zero user redemption records" sh -c "strided_new q records list-user-redemption-record -o json | jq -e '.user_redemption_record | length == 0'"
log_cmd "deposit records" strided_new q records list-deposit-record -o json
checkpoint "stranded deposit never staked" sh -c "strided_new q records list-deposit-record -o json | jq -e '[.deposit_record[] | select(.status==\"DELEGATION_QUEUE\")] | length >= 1'"
sleep 180   # four stride epochs
checkpoint "hub rate frozen"  assert_rate_unchanged cosmoshub-test-1 $RATE_HUB
checkpoint "osmo rate frozen" assert_rate_unchanged osmosis-test-1 $RATE_OSMO
checkpoint "no new epoch unbonding records" sh -c "strided_new q records list-epoch-unbonding-record -o json | jq -e '[.epoch_unbonding_record[] | select(.epoch_number > '$(strided_new q stakeibc show-epoch-tracker day -o json | jq -r .epoch_tracker.epoch_number)' - 10 and (.host_zone_unbondings|length)==0)] | length == 0'"
checkpoint "no reinvest/claim-rewards ICA" sh -c "! $KX logs stride-validator-0 -c validator --since=3m | grep -Ei 'reinvest|ClaimAccruedStakingRewards|withdrawal balance'"
```

- [ ] **Step 3: phase4_drain.sh**

```bash
source "$(dirname "$0")/lib.sh"; source "$REHEARSAL_DIR/state.env"; log "## Phase 4: admin drain"
no_flags() { strided_new q stakeibc show-validators $1 -o json | jq -e '[.validators[].delegation_changes_in_progress] | all(. == 0)'; }
wait_until 300 "hub flags clear"  no_flags cosmoshub-test-1
wait_until 300 "osmo flags clear" no_flags osmosis-test-1
submit_window() { local n; n=$(day_epoch_next_start); local now; now=$(date +%s); (( n - now < 50 )) && sleep_until $((n+5)); }
# Live test: the smallest hub validator in full (vals have equal weight, so pick val8 which also drained nothing yet)
HUB_VALS=$(strided_new q stakeibc show-validators cosmoshub-test-1 -o json | jq -r '.validators[].address'); SMALL=$(tail -1 <<<"$HUB_VALS")
echo "[{\"address\":\"$SMALL\",\"offset\":\"0\"}]" | $KX exec -i stride-validator-0 -c validator -- sh -c 'cat > /tmp/one.json'
submit_window; ms_tx strided_new admin-ms m1,m2,m3 -- stakeibc undelegate-from-validators cosmoshub-test-1 /tmp/one.json
wait_until 240 "live-test ack" sh -c "strided_new q stakeibc show-validators cosmoshub-test-1 -o json | jq -e '.validators[] | select(.address==\"$SMALL\") | .delegation == \"0\" and .delegation_changes_in_progress == 0'"
checkpoint "nothing burned" sh -c "strided_new q bank total --denom stuatom -o json | jq -e '.amount.amount == \"$(cat $REHEARSAL_DIR/state.env | grep STATOM_SUPPLY | cut -d= -f2)\"'"
# Injection 1: slash hub val7 after the refresh, then --all: val7's batch fails, the others succeed
HUB_VAL7=$(sed -n 7p <<<"$HUB_VALS")
$KX exec cosmoshub-validator-6 -c validator -- sh -c 'kill -STOP $(pidof gaiad)'
wait_until 240 "hub val7 jailed" sh -c "gaiad q staking validator $HUB_VAL7 -o json | jq -e '.validator.jailed == true'"
$KX exec cosmoshub-validator-6 -c validator -- sh -c 'kill -CONT $(pidof gaiad)'
submit_window; ms_tx strided_new admin-ms m1,m2,m3 -- stakeibc undelegate-from-validators cosmoshub-test-1 --all
wait_until 240 "drain acks" no_flags cosmoshub-test-1
log_cmd "hub validators after --all" strided_new q stakeibc show-validators cosmoshub-test-1 -o json
checkpoint "val7 batch failed, others drained" sh -c "strided_new q stakeibc show-validators cosmoshub-test-1 -o json | jq -e '[.validators[] | select(.address != \"$HUB_VAL7\") | .delegation|tonumber] | max < 1000000 and ([.validators[] | select(.address == \"$HUB_VAL7\") | .delegation|tonumber] | .[0] > 1000000)'"
ms_tx strided_new admin-ms m1,m2,m3 -- stakeibc update-delegation cosmoshub-test-1 $HUB_VAL7
wait_until 120 "val7 refreshed" sh -c "strided_new q stakeibc show-validators cosmoshub-test-1 -o json | jq -e '.validators[] | select(.address==\"$HUB_VAL7\") | .slash_query_in_progress == false'"
echo "[{\"address\":\"$HUB_VAL7\",\"offset\":\"0\"}]" | $KX exec -i stride-validator-0 -c validator -- sh -c 'cat > /tmp/v7.json'
submit_window; ms_tx strided_new admin-ms m1,m2,m3 -- stakeibc undelegate-from-validators cosmoshub-test-1 /tmp/v7.json
wait_until 240 "val7 drained" sh -c "strided_new q stakeibc show-validators cosmoshub-test-1 -o json | jq -e '.validators[] | select(.address==\"$HUB_VAL7\") | (.delegation|tonumber) < 1000000'"
# Injection 2 (osmosis zone): ICA timeout -> channel closes -> restore -> resubmit. Pause the stride-osmosis relayer past the day-epoch timeout.
OSMO_VALS=$(strided_new q stakeibc show-validators osmosis-test-1 -o json | jq -r '.validators[].address'); OV1=$(head -1 <<<"$OSMO_VALS")
echo "[{\"address\":\"$OV1\",\"offset\":\"0\"}]" | $KX exec -i stride-validator-0 -c validator -- sh -c 'cat > /tmp/ov1.json'
$KX scale deployment relayer-stride-osmosis --replicas=0
submit_window; ms_tx strided_new admin-ms m1,m2,m3 -- stakeibc undelegate-from-validators osmosis-test-1 /tmp/ov1.json
sleep_until $(( $(day_epoch_next_start) + 20 )); $KX scale deployment relayer-stride-osmosis --replicas=1
OSMO_CONN=connection-1
wait_until 300 "osmo delegation channel closed" sh -c "strided_new q stakeibc show-host-zone osmosis-test-1 -o json >/dev/null && strided_new q ibc channel channels -o json | jq -e '[.channels[] | select(.port_id | startswith(\"icacontroller-osmosis-test-1.DELEGATION\")) | .state] | index(\"STATE_CLOSED\") != null'"
ms_tx strided_new admin-ms m1,m2,m3 -- stakeibc restore-interchain-account osmosis-test-1 $OSMO_CONN osmosis-test-1.DELEGATION
wait_until 300 "osmo delegation channel reopened" sh -c "strided_new q ibc channel channels -o json | jq -e '[.channels[] | select(.port_id | startswith(\"icacontroller-osmosis-test-1.DELEGATION\")) | .state] | index(\"STATE_OPEN\") != null'"
checkpoint "flags reset by restore" no_flags osmosis-test-1
# Injection 3: offset drain of osmo val2, then --all for the rest; the offset stays recorded and is calibrated away
OV2=$(sed -n 2p <<<"$OSMO_VALS")
echo "[{\"address\":\"$OV2\",\"offset\":\"1000000\"}]" | $KX exec -i stride-validator-0 -c validator -- sh -c 'cat > /tmp/ov2.json'
submit_window; ms_tx strided_new admin-ms m1,m2,m3 -- stakeibc undelegate-from-validators osmosis-test-1 /tmp/ov2.json
wait_until 240 "offset ack" sh -c "strided_new q stakeibc show-validators osmosis-test-1 -o json | jq -e '.validators[] | select(.address==\"$OV2\") | .delegation == \"1000000\"'"
submit_window; ms_tx strided_new admin-ms m1,m2,m3 -- stakeibc undelegate-from-validators osmosis-test-1 --all
wait_until 240 "osmo drained" no_flags osmosis-test-1
# Injection 4: the dead window: submit in the last fifth of the day epoch and expect a failed send, nothing flagged
sleep_until $(( $(day_epoch_next_start) - 25 ))
CHECKPOINT_SOFT=1 checkpoint "dead-window send fails" sh -c "ms_tx strided_new admin-ms m1,m2,m3 -- stakeibc undelegate-from-validators cosmoshub-test-1 --all 2>&1 | grep -qi 'timeout\\|fail\\|code=[1-9]'"
checkpoint "no flags after dead window" no_flags cosmoshub-test-1
checkpoint "hub total at dust"  sh -c "strided_new q stakeibc show-host-zone cosmoshub-test-1 -o json | jq -e '(.host_zone.total_delegations|tonumber) < 2000000'"
checkpoint "osmo total at dust" sh -c "strided_new q stakeibc show-host-zone osmosis-test-1 -o json | jq -e '(.host_zone.total_delegations|tonumber) < 2000000'"
checkpoint "hub rate frozen"  assert_rate_unchanged cosmoshub-test-1 $RATE_HUB
checkpoint "osmo rate frozen" assert_rate_unchanged osmosis-test-1 $RATE_OSMO
```

The "lost ack" injection (batch executes, ack lost, `CalibrateDelegation` empty-response path) is run by hand by the orchestrator after injection 2 if the closed channel left an executed batch unacknowledged; the script logs the validator delegations on the host (`gaiad q staking delegations <delegation ICA>`) right after the restore so the orchestrator can see whether it happened. Record `STATOM_SUPPLY` (bank total of `stuatom`) into `state.env` at the start of phase 4.

- [ ] **Step 4: Syntax-check and commit** — `bash -n` all three; `git add integration-tests/rehearsal && git commit -m "rehearsal: day-0, redemption and drain phase scripts"`.

### Task 10: `phase5_transfers.sh` and `phase6_staketia.sh`

**Files:** Create both. **Depends on:** Tasks 1-7. **Review:** yes

- [ ] **Step 1: phase5_transfers.sh**

```bash
source "$(dirname "$0")/lib.sh"; source "$REHEARSAL_DIR/state.env"; log "## Phase 5: transfers to Osmosis"
ATOM_ON_OSMO=ibc/$(printf 'transfer/channel-1/uatom' | shasum -a 256 | cut -d' ' -f1 | tr a-f A-F)
vault_bal() { osmosisd q bank balance $VAULT_MS_OSMO $1 -o json | jq -r '.balance.amount'; }
ica_bal() { # chain zone type
  local addr; addr=$(strided_new q stakeibc show-host-zone $2 -o json | jq -r ".host_zone.$(tr A-Z a-z <<<$3)_ica_address"); $1 q bank balances $addr -o json | jq -r '.balances[0].amount // "0"'; }
pre_checklist() { # zone: every record CLAIMABLE or zero, no user records, no claim pending
  strided_new q records list-epoch-unbonding-record -o json | jq -e --arg z $1 '[.epoch_unbonding_record[].host_zone_unbondings[] | select(.host_zone_id==$z and (.native_token_amount|tonumber)>0 and .status!="CLAIMABLE")] | length == 0' &&
  strided_new q records list-user-redemption-record -o json | jq -e --arg z $1 '[.user_redemption_record[] | select(.host_zone_id==$z)] | length == 0'; }
B0=$(vault_bal $ATOM_ON_OSMO); O0=$(vault_bal uosmo)
# Live test: WITHDRAWAL and FEE first, both zones
for t in WITHDRAWAL FEE; do
  amt=$(ica_bal gaiad cosmoshub-test-1 $t);    ms_tx strided_new admin-ms m1,m2,m3 -- stakeibc transfer-from-ica cosmoshub-test-1 $t ${amt}uatom
  amt=$(ica_bal osmosisd osmosis-test-1 $t);   ms_tx strided_new admin-ms m1,m2,m3 -- stakeibc transfer-from-ica osmosis-test-1 $t ${amt}uosmo
done
wait_until 300 "hub tokens landed in vault as ATOM-on-Osmosis" sh -c "[ \$(vault_bal $ATOM_ON_OSMO) -gt $B0 ]"
wait_until 120 "osmo bank-send form landed" sh -c "[ \$(vault_bal uosmo) -gt $O0 ]"
# Injection: a transfer that times out and refunds (relayer paused for longer than the 15-minute timeout)
$KX scale deployment relayer-cosmoshub-osmosis --replicas=0
W=$(ica_bal gaiad cosmoshub-test-1 WITHDRAWAL); [[ "$W" == 0 ]] && { gaiad tx bank send user1 $(strided_new q stakeibc show-host-zone cosmoshub-test-1 -o json | jq -r .host_zone.withdrawal_ica_address) 1000000uatom $HUB_TX >/dev/null; sleep 6; W=1000000; }
ms_tx strided_new admin-ms m1,m2,m3 -- stakeibc transfer-from-ica cosmoshub-test-1 WITHDRAWAL ${W}uatom
log "sleeping past the 15-minute transfer timeout"; sleep 960; $KX scale deployment relayer-cosmoshub-osmosis --replicas=1
wait_until 300 "timed-out transfer refunded to the withdrawal ICA" sh -c "[ \$(ica_bal gaiad cosmoshub-test-1 WITHDRAWAL) -ge $W ]"
ms_tx strided_new admin-ms m1,m2,m3 -- stakeibc transfer-from-ica cosmoshub-test-1 WITHDRAWAL ${W}uatom
wait_until 300 "resubmission landed" sh -c "[ \$(ica_bal gaiad cosmoshub-test-1 WITHDRAWAL) -lt $W ]"
# Full balances once the pre-transfer checklist passes
checkpoint "hub pre-transfer checklist"  pre_checklist cosmoshub-test-1
checkpoint "osmo pre-transfer checklist" pre_checklist osmosis-test-1
for z in cosmoshub-test-1:gaiad:uatom osmosis-test-1:osmosisd:uosmo; do IFS=: read zone chain denom <<<"$z"
  for t in DELEGATION WITHDRAWAL REDEMPTION; do amt=$(ica_bal $chain $zone $t); [[ "$amt" -gt 0 ]] && ms_tx strided_new admin-ms m1,m2,m3 -- stakeibc transfer-from-ica $zone $t ${amt}${denom}; done
done
for z in cosmoshub-test-1:gaiad osmosis-test-1:osmosisd; do IFS=: read zone chain <<<"$z"
  for t in DELEGATION WITHDRAWAL FEE REDEMPTION; do wait_until 300 "$zone $t at dust" sh -c "[ \$(ica_bal $chain $zone $t) -lt 1000 ]"; done
done
log_cmd "vault balances" osmosisd q bank balances $VAULT_MS_OSMO -o json
```

`ica_bal` must read the host denom's balance, not `.balances[0]`; use `jq -r --arg d <denom> '.balances[] | select(.denom==$d) | .amount // "0"'` with the denom passed in.

- [ ] **Step 2: phase6_staketia.sh**

```bash
source "$(dirname "$0")/lib.sh"; source "$REHEARSAL_DIR/state.env"; log "## Phase 6: staketia"
wait_until 400 "hub multisig unbonding matured" sh -c "gaiad q staking unbonding-delegations $HUB_MS_COSMOS -o json | jq -e '.unbonding_responses | length == 0'"
BAL=$(gaiad q bank balance $HUB_MS_COSMOS uatom -o json | jq -r .balance.amount)
exec_transfer() { # receiver memo amount -> tx hash
  cat > /tmp/xfer.json <<EOF
{"body":{"messages":[{"@type":"/ibc.applications.transfer.v1.MsgTransfer","source_port":"transfer","source_channel":"channel-0","token":{"denom":"uatom","amount":"$3"},
 "sender":"$HUB_MS_COSMOS","receiver":"$1","timeout_height":{"revision_number":"0","revision_height":"0"},"timeout_timestamp":"$(( ($(date +%s)+900) * 1000000000 ))","memo":"$2"}]}}
EOF
  $KX cp /tmp/xfer.json cosmoshub-validator-0:/tmp/xfer.json -c validator
  gaiad tx authz exec /tmp/xfer.json --from st-operator $HUB_TX | tx_hash; }
checkpoint "transfer with memo rejected by grant" sh -c "h=\$(exec_transfer $STAKETIA_CLAIM hello 1000); ! wait_tx gaiad \$h"
checkpoint "transfer to other receiver rejected"  sh -c "h=\$(exec_transfer $USER1_STRIDE '' 1000); ! wait_tx gaiad \$h"
H=$(exec_transfer $STAKETIA_CLAIM '' $BAL); wait_tx gaiad $H
wait_until 120 "claim address funded" sh -c "strided_new q bank balance $STAKETIA_CLAIM $ATOM_ON_STRIDE -o json | jq -e '(.balance.amount|tonumber) >= $BAL'"
for id in $(strided_new q staketia unbonding-records -o json | jq -r '.unbonding_records[] | select(.status=="UNBONDED") | .id'); do
  log_cmd "confirm-sweep $id" strided_new tx staketia confirm-sweep $id $H --from st-operator $STRIDE_TX
done
wait_until 120 "redeemers paid (hour epoch)" sh -c "strided_new q staketia redemption-records -o json | jq -e '.redemption_records | length == 0'"
checkpoint "every unbonding record CLAIMED or empty" sh -c "strided_new q staketia unbonding-records -o json | jq -e '[.unbonding_records[] | select(.status != \"CLAIMED\" and (.native_amount|tonumber) > 0)] | length == 0'"
DEL_ICA=$(strided_new q stakeibc show-host-zone cosmoshub-test-1 -o json | jq -r .host_zone.delegation_ica_address)
D0=$(gaiad q bank balance $DEL_ICA uatom -o json | jq -r .balance.amount)
ms_tx strided_new admin-ms m1,m2,m3 -- stakeibc transfer-staketia-claim-balance 1000000
wait_until 120 "1 ATOM landed on the delegation ICA" sh -c "[ \$(gaiad q bank balance $DEL_ICA uatom -o json | jq -r .balance.amount) -ge $((D0+1000000)) ]"
ms_tx strided_new admin-ms m1,m2,m3 -- stakeibc transfer-staketia-claim-balance 0
wait_until 120 "claim address empty" sh -c "strided_new q bank balance $STAKETIA_CLAIM $ATOM_ON_STRIDE -o json | jq -e '(.balance.amount|tonumber) == 0'"
ms_tx strided_new admin-ms m1,m2,m3 -- stakeibc transfer-from-ica cosmoshub-test-1 DELEGATION $(gaiad q bank balance $DEL_ICA uatom -o json | jq -r .balance.amount)uatom
wait_until 300 "staketia ATOM in the vault" sh -c "[ \$(gaiad q bank balance $DEL_ICA uatom -o json | jq -r .balance.amount) -lt 1000 ]"
```

Phase 6 runs before phase 5's final `DELEGATION` transfer on the Hub when the staketia unbonding matures first; the orchestrator runs 5 and 6 in whichever order the timing allows and both are safe to rerun (each transfer sends the current balance).

- [ ] **Step 3: Syntax-check and commit** — `bash -n`; `git commit -m "rehearsal: transfer and staketia phase scripts"`.

### Task 11: `phase7_pools.sh`

**Files:** Create `integration-tests/rehearsal/phase7_pools.sh`, `integration-tests/rehearsal/instantiate_pool.sh`. **Depends on:** Tasks 1-7. **Review:** yes

- [ ] **Step 1: instantiate_pool.sh** — prints the instantiate JSON for a pool:

```bash
#!/bin/bash
# usage: instantiate_pool.sh <sttoken-denom-on-osmosis> <native-denom> <rate 18dp> <subdenom> <admin>
rate_factor=$(python3 -c "from decimal import Decimal; print(int(Decimal('$3') * 10**18))")
jq -n --arg st "$1" --arg nat "$2" --arg f "$rate_factor" --arg sub "$4" --arg adm "$5" \
 '{pool_asset_configs:[{denom:$st, normalization_factor:"1000000000000000000"},{denom:$nat, normalization_factor:$f}],
   alloyed_asset_subdenom:$sub, alloyed_asset_normalization_factor:$f, admin:$adm, moderator:$adm}'
```

- [ ] **Step 2: phase7_pools.sh**

```bash
source "$(dirname "$0")/lib.sh"; source "$REHEARSAL_DIR/state.env"; log "## Phase 7: pools"
STATOM_CANON=ibc/$(printf 'transfer/channel-0/stuatom' | shasum -a 256 | cut -d' ' -f1 | tr a-f A-F)
STOSMO_CANON=ibc/$(printf 'transfer/channel-0/stuosmo' | shasum -a 256 | cut -d' ' -f1 | tr a-f A-F)
STATOM_HUBROUTE=ibc/$(printf 'transfer/channel-1/transfer/channel-0/stuatom' | shasum -a 256 | cut -d' ' -f1 | tr a-f A-F)
ATOM_ON_OSMO=ibc/$(printf 'transfer/channel-1/uatom' | shasum -a 256 | cut -d' ' -f1 | tr a-f A-F)
# Upload the mainnet bytecode (code id 1 is whitelisted at genesis)
$KX cp "$REHEARSAL_DIR/transmuter_v3.2.0.wasm" osmosis-validator-0:/tmp/transmuter.wasm -c validator
if ! osmosisd q wasm code-info 1 >/dev/null 2>&1; then log_cmd "store code" osmosisd tx wasm store /tmp/transmuter.wasm --from val1 $OSMO_TX --gas 5000000; sleep 6; fi
checkpoint "code 1 stored" osmosisd q wasm code-info 1
create_pool() { # name sttoken native rate subdenom -> pool id
  local inst; inst=$(bash "$REHEARSAL_DIR/instantiate_pool.sh" "$2" "$3" "$4" "$5" "$VAULT_MS_OSMO")
  $KX exec osmosis-validator-0 -c validator -- sh -c "cat > /tmp/inst_$1.json" <<<"$inst"
  ms_tx osmosisd vault-ms m1,m2,m3 -- cosmwasmpool create-pool 1 "$(cat <<<"$inst")" >/dev/null
  osmosisd q poolmanager all-pools -o json | jq -r '.pools | last | .pool_id'; }
POOL_ATOM=$(create_pool atom $STATOM_CANON $ATOM_ON_OSMO $RATE_HUB stATOMr)
POOL_OSMO=$(create_pool osmo $STOSMO_CANON uosmo $RATE_OSMO stOSMOr)
POOL_ATOM_HUB=$(create_pool atomhub $STATOM_HUBROUTE $ATOM_ON_OSMO $RATE_HUB stATOMhub)
printf 'POOL_ATOM=%s\nPOOL_OSMO=%s\nPOOL_ATOM_HUB=%s\n' $POOL_ATOM $POOL_OSMO $POOL_ATOM_HUB >> "$REHEARSAL_DIR/state.env"
contract() { osmosisd q poolmanager pool $1 -o json | jq -r '.pool.contract_address'; }
# Allocation: route pool = escrow(channel-0 on Stride) x rate; canonical = the rest. Coverage check gates the funding.
STRIDE_EXPORT=$REHEARSAL_DIR/export_pre_pools.json
bash "$REHEARSAL_DIR/export.sh" "$STRIDE_EXPORT"
cat > /tmp/pools.json <<EOF
[{"chain_id":"cosmoshub-test-1","pool_id":"$POOL_ATOM","sttoken_denom":"$STATOM_CANON","route_channel":null},
 {"chain_id":"cosmoshub-test-1","pool_id":"$POOL_ATOM_HUB","sttoken_denom":"$STATOM_HUBROUTE","route_channel":"channel-0"},
 {"chain_id":"osmosis-test-1","pool_id":"$POOL_OSMO","sttoken_denom":"$STOSMO_CANON","route_channel":null}]
EOF
CHECKPOINT_SOFT=1 checkpoint "coverage before funding (expected: shortfall)" python3 "$REPO/scripts/wind-down/coverage_check.py" --export $STRIDE_EXPORT --pools /tmp/pools.json --vault $VAULT_MS_OSMO
ESCROW_HUB=$(strided_new q ibc-transfer escrow-address transfer channel-0 2>/dev/null || strided_new q ibc-transfer escrow-address channel-0)
ROUTE_ESCROW=$(strided_new q bank balance $ESCROW_HUB stuatom -o json | jq -r .balance.amount)
ROUTE_ALLOC=$(python3 -c "from decimal import Decimal; import math; print(math.ceil(Decimal('$ROUTE_ESCROW')*Decimal('$RATE_HUB')))")
ATOM_TOTAL=$(osmosisd q bank balance $VAULT_MS_OSMO $ATOM_ON_OSMO -o json | jq -r .balance.amount)
fund() { # pool amount denom
  ms_tx osmosisd vault-ms m1,m2,m3 -- wasm execute $(contract $1) '{"join_pool":{}}' --amount ${2}${3}
  ms_tx osmosisd vault-ms m1,m2,m3 -- wasm execute $(contract $1) "{\"mark_corrupted_assets\":{\"denoms\":[\"$3\"]}}"; }
# Test deposit, then the remainder, then mark (unmark/join/re-mark is the "complete an allocation" path)
ms_tx osmosisd vault-ms m1,m2,m3 -- wasm execute $(contract $POOL_ATOM_HUB) '{"join_pool":{}}' --amount 1000000$ATOM_ON_OSMO
ms_tx osmosisd vault-ms m1,m2,m3 -- wasm execute $(contract $POOL_ATOM_HUB) "{\"mark_corrupted_assets\":{\"denoms\":[\"$ATOM_ON_OSMO\"]}}"
checkpoint "join blocked while marked" sh -c "! ms_tx osmosisd vault-ms m1,m2,m3 -- wasm execute $(contract $POOL_ATOM_HUB) '{\"join_pool\":{}}' --amount 1$ATOM_ON_OSMO"
ms_tx osmosisd vault-ms m1,m2,m3 -- wasm execute $(contract $POOL_ATOM_HUB) "{\"unmark_corrupted_assets\":{\"denoms\":[\"$ATOM_ON_OSMO\"]}}"
fund $POOL_ATOM_HUB $((ROUTE_ALLOC - 1000000)) $ATOM_ON_OSMO
fund $POOL_ATOM $((ATOM_TOTAL - ROUTE_ALLOC)) $ATOM_ON_OSMO
fund $POOL_OSMO $(osmosisd q bank balance $VAULT_MS_OSMO uosmo -o json | jq -r '.balance.amount') uosmo
# Outsider joins the canonical pool with stTokens: shows as a remaining claim in the coverage check
log_cmd "outsider join" osmosisd tx wasm execute $(contract $POOL_ATOM) '{"join_pool":{}}' --amount 1000000$STATOM_CANON --from user1 $OSMO_TX
# Swaps: exact rate, rounded down; native -> stToken refused
Q=$(osmosisd q wasm contract-state smart $(contract $POOL_ATOM) "{\"calc_out_amt_given_in\":{\"token_in\":{\"denom\":\"$STATOM_CANON\",\"amount\":\"1000000\"},\"token_out_denom\":\"$ATOM_ON_OSMO\",\"swap_fee\":\"0\"}}" -o json | jq -r '.data.token_out.amount')
EXP=$(python3 -c "from decimal import Decimal; print(int(Decimal('1000000')*Decimal('$RATE_HUB')))")
checkpoint "quote = floor(1e6 x rate)" test "$Q" = "$EXP"
log_cmd "swap statom->atom" osmosisd tx poolmanager swap-exact-amount-in 1000000$STATOM_CANON 1 --swap-route-pool-ids $POOL_ATOM --swap-route-denoms $ATOM_ON_OSMO --from user1 $OSMO_TX
checkpoint "atom->statom refused" sh -c "osmosisd tx poolmanager swap-exact-amount-in 1000$ATOM_ON_OSMO 1 --swap-route-pool-ids $POOL_ATOM --swap-route-denoms $STATOM_CANON --from user1 $OSMO_TX | jq -e '.code != 0'"
log_cmd "route-pool swap (two-hop statom)" osmosisd tx poolmanager swap-exact-amount-in 1000000$STATOM_HUBROUTE 1 --swap-route-pool-ids $POOL_ATOM_HUB --swap-route-denoms $ATOM_ON_OSMO --from user1 $OSMO_TX
bash "$REHEARSAL_DIR/export.sh" "$REHEARSAL_DIR/export_post_funding.json"
checkpoint "coverage after funding" python3 "$REPO/scripts/wind-down/coverage_check.py" --export "$REHEARSAL_DIR/export_post_funding.json" --pools /tmp/pools.json --vault $VAULT_MS_OSMO
checkpoint "pool gate" python3 "$REPO/scripts/wind-down/check_transmuter_pool.py"
```

Before `check_transmuter_pool.py` runs, append the three `PoolSpec` entries to its `POOLS` list and fill `COSMWASMPOOL_MODULE` from `osmosisd q auth module-account cosmwasmpool`; the script refuses to run while the `REHEARSAL_FILL_IN` marker is present. The `pools.json` keys must match what `coverage_check.py` reads; copy its documented shape from the PR 6 plan (`docs/superpowers/plans/2026-09-29-wind-down-pr6-release-gate.md`).

- [ ] **Step 3: export.sh** (used by phases 7 and 8):

```bash
#!/bin/bash
# usage: export.sh <out.json>. Stops stride-validator-3, exports, restarts it.
source "$(dirname "$0")/lib.sh"
$KX exec stride-validator-3 -c validator -- sh -c 'kill -STOP $(pidof strided)'
sleep 3
$KX exec stride-validator-3 -c validator -- sh -c "$(strided_old version 2>/dev/null | grep -q ^v34 && echo strided || echo $NEW_BIN) export --home /home/validator/.stride 2>/dev/null" > "$1"
$KX exec stride-validator-3 -c validator -- sh -c 'kill -CONT $(pidof strided)'
jq -e '.app_state.bank.supply' "$1" >/dev/null && log "export written: $1 ($(wc -c <"$1") bytes)"
```

If `export` refuses while the process is only paused (db lock), use `$KX exec ... -- sh -c 'kill $(pidof strided)'` and let the container restart instead.

- [ ] **Step 4: Syntax-check and commit** — `bash -n`; `git commit -m "rehearsal: pool phase scripts"`.

### Task 12: `phase8_sweep.sh` and `phase9_halt.sh`

**Files:** Create both. **Depends on:** Tasks 1-7. **Review:** yes (the sweep moves user balances)

- [ ] **Step 1: phase8_sweep.sh**

```bash
source "$(dirname "$0")/lib.sh"; source "$REHEARSAL_DIR/state.env"; log "## Phase 8: sweep off Stride"
EXPORT=$REHEARSAL_DIR/export_pre_sweep.json; bash "$REHEARSAL_DIR/export.sh" "$EXPORT"
cat > /tmp/prices.json <<EOF
{"stuatom":{"usd_per_token":4,"decimals":6},"stuosmo":{"usd_per_token":0.5,"decimals":6},"ustrd":{"usd_per_token":0.1,"decimals":6},"$ATOM_ON_STRIDE":{"usd_per_token":4,"decimals":6}}
EOF
ASOF=$(strided_new q block --type=height $(jq -r '.initial_height' $EXPORT) -o json 2>/dev/null | jq -r '.header.time' || date -u +%FT%TZ)
rm -rf /tmp/batches; python3 "$REPO/scripts/wind-down/build_sweep_batches.py" --export $EXPORT --prices /tmp/prices.json --floor-usd 1 \
  --denoms stuatom,stuosmo,ustrd --extra-denom "$ATOM_ON_STRIDE=4:6" --out-dir /tmp/batches --as-of "$ASOF" --batch-size 50 | tee -a "$LOG"
checkpoint "admin cannot sweep" sh -c "f=\$(ls /tmp/batches/*.json | head -1); $KX cp \$f stride-validator-0:/tmp/b.json -c validator; strided_new tx stakeibc sweep-tokens-off-stride stuatom /tmp/b.json --from admin $STRIDE_TX 2>&1 | grep -qi 'sweep operator\\|unauthorized\\|invalid'"
for f in /tmp/batches/*.json; do
  $KX cp $f stride-validator-0:/tmp/batch.json -c validator
  out=$(strided_new tx stakeibc sweep-tokens-off-stride stuatom,stuosmo,ustrd,$ATOM_ON_STRIDE /tmp/batch.json --from sweep-operator $STRIDE_TX); h=$(tx_hash <<<"$out")
  wait_tx strided_new $h; log "gas used: $(strided_new q tx $h -o json | jq -r '.gas_used') for $(jq length $f) addresses"
  checkpoint "builder batch skipped nothing ($f)" sh -c "strided_new q tx $h -o json | jq -e '[.events[] | select(.type==\"sweep_tokens_off_stride\") | .attributes[] | select(.key==\"num_skipped\") | .value] | all(. == \"0\")'"
done
# Hand-built batch with the distribution module and the escrow account: both skipped with a reason
DIST=$(strided_new q auth module-account distribution -o json | jq -r '.account.value.address // .account.base_account.address')
ESCROW=$(strided_new q ibc-transfer escrow-address transfer channel-0 2>/dev/null || strided_new q ibc-transfer escrow-address channel-0)
echo "[\"$DIST\",\"$ESCROW\",\"$USER1_STRIDE\"]" | $KX exec -i stride-validator-0 -c validator -- sh -c 'cat > /tmp/skip.json'
h=$(strided_new tx stakeibc sweep-tokens-off-stride stuatom /tmp/skip.json --from sweep-operator $STRIDE_TX | tx_hash); wait_tx strided_new $h
checkpoint "two skipped with reasons" sh -c "strided_new q tx $h -o json | jq -e '[.events[] | select(.type==\"sweep_tokens_off_stride\") | .attributes[] | select(.key==\"num_skipped\") | .value] | index(\"2\") != null'"
wait_until 300 "stATOM landed on osmosis for holder-base" sh -c "osmosisd q bank balance $HOLDER_BASE_OSMO $(printf 'ibc/%s' $(printf 'transfer/channel-0/stuatom' | shasum -a 256 | cut -d' ' -f1 | tr a-f A-F)) -o json | jq -e '(.balance.amount|tonumber) > 0'"
wait_until 300 "ATOM unwound to the hub for holder-base" sh -c "gaiad q bank balance $(python3 -c \"import sys; sys.path.insert(0,'$REPO/scripts/wind-down'); import bech32_ref; print(bech32_ref.encode('cosmos', bech32_ref.decode('$HOLDER_BASE')[1]))\") uatom -o json | jq -e '(.balance.amount|tonumber) > 0'"
checkpoint "holder-base empty on stride" sh -c "strided_new q bank balances $HOLDER_BASE -o json | jq -e '[.balances[] | select(.denom==\"stuatom\" or .denom==\"ustrd\")] | length == 0'"
```

The timeout-refund case for the sweep: pause `relayer-stride-osmosis` before sending one batch of one address, wait 16 minutes, confirm the holder's balance is back, resume the relayer and resend that batch. Add this as the last block of the script with the same pattern as phase 5's injection.

- [ ] **Step 2: phase9_halt.sh**

```bash
source "$(dirname "$0")/lib.sh"; source "$REHEARSAL_DIR/state.env"; log "## Phase 9: halt"
checkpoint "zero user redemption records" sh -c "strided_new q records list-user-redemption-record -o json | jq -e '.user_redemption_record | length == 0'"
checkpoint "no flags" sh -c "for z in cosmoshub-test-1 osmosis-test-1; do strided_new q stakeibc show-validators \$z -o json | jq -e '[.validators[].delegation_changes_in_progress] | all(. == 0)' || exit 1; done"
checkpoint "staketia clean" sh -c "strided_new q staketia redemption-records -o json | jq -e '.redemption_records | length == 0'"
checkpoint "channels clear both ways" sh -c "for c in channel-0 channel-1; do strided_new q ibc channel packet-commitments transfer \$c -o json | jq -e '.commitments | length == 0' || exit 1; done; gaiad q ibc channel packet-commitments transfer channel-0 -o json | jq -e '.commitments | length == 0'; osmosisd q ibc channel packet-commitments transfer channel-0 -o json | jq -e '.commitments | length == 0'"
bash "$REHEARSAL_DIR/export.sh" "$REHEARSAL_DIR/export_pre_halt.json"
checkpoint "coverage before halt" python3 "$REPO/scripts/wind-down/coverage_check.py" --export "$REHEARSAL_DIR/export_pre_halt.json" --pools /tmp/pools.json --vault $VAULT_MS_OSMO
H=$(( $(strided_new status | jq -r '.sync_info.latest_block_height // .SyncInfo.latest_block_height') + 30 ))
for i in 0 1 2 3; do $KX exec stride-validator-$i -c validator -- sh -c "sed -i 's/^halt-height = .*/halt-height = $H/' /home/validator/.stride/config/app.toml && kill \$(pidof strided)"; done
wait_until 300 "stride halted at $H" sh -c "! strided_new status >/dev/null 2>&1 || [ \$(strided_new status | jq -r '.sync_info.latest_block_height // .SyncInfo.latest_block_height') -ge $H ]"
log "Stride halted; swapping every holder's balance on Osmosis"
# Final invariant: every stToken holder on Osmosis swaps everything; every swap succeeds; every pool keeps a non-negative native surplus
contract() { osmosisd q poolmanager pool $1 -o json | jq -r '.pool.contract_address'; }
STATOM_CANON=ibc/$(printf 'transfer/channel-0/stuatom' | shasum -a 256 | cut -d' ' -f1 | tr a-f A-F); ATOM_ON_OSMO=ibc/$(printf 'transfer/channel-1/uatom' | shasum -a 256 | cut -d' ' -f1 | tr a-f A-F)
for holder in user1 holder-base holder-vesting; do
  bal=$(osmosisd q bank balance $(osmosisd keys show $holder -a --keyring-backend test) $STATOM_CANON -o json | jq -r .balance.amount)
  [[ "$bal" -gt 0 ]] && checkpoint "$holder swaps $bal" sh -c "osmosisd tx poolmanager swap-exact-amount-in ${bal}$STATOM_CANON 1 --swap-route-pool-ids $POOL_ATOM --swap-route-denoms $ATOM_ON_OSMO --from $holder $OSMO_TX | jq -e '.code == 0'"
done
log_cmd "final pool liquidity" osmosisd q wasm contract-state smart $(contract $POOL_ATOM) '{"get_total_pool_liquidity":{}}' -o json
checkpoint "canonical pool surplus >= 0" sh -c "osmosisd q wasm contract-state smart $(contract $POOL_ATOM) '{\"get_total_pool_liquidity\":{}}' -o json | jq -e '[.data.total_pool_liquidity[] | select(.denom==\"$ATOM_ON_OSMO\") | .amount|tonumber] | .[0] >= 0'"
```

`holder-vesting` on Osmosis means the same key with the `osmo` prefix (the keyring entry exists on the Osmosis pod because init-chain restored every rehearsal key on every chain). The halt via `app.toml` requires the container to restart `strided`: the validator container's main process is `strided start` under cosmovisor, so killing it restarts the pod; if the pod's restart loses the edit (the config is on the state volume, so it should not), fall back to `kubectl scale statefulset stride-validator --replicas=0` as the halt.

- [ ] **Step 3: Syntax-check and commit** — `bash -n`; `git commit -m "rehearsal: sweep and halt phase scripts"`.

### Task 13: `phase0_preflight.sh`

**Files:** Create `integration-tests/rehearsal/phase0_preflight.sh`. **Depends on:** Tasks 1-7. **Review:** no

- [ ] **Step 1: Write it**

```bash
source "$(dirname "$0")/lib.sh"; log "## Phase 0: pre-flight"
checkpoint "stride channel-0 -> cosmoshub"  sh -c "strided_old q ibc channel end transfer channel-0 -o json | jq -e '.channel.state==\"STATE_OPEN\"' && strided_old q ibc channel client-state transfer channel-0 -o json | jq -e '.client_state.chain_id==\"cosmoshub-test-1\"'"
checkpoint "stride channel-1 -> osmosis"    sh -c "strided_old q ibc channel client-state transfer channel-1 -o json | jq -e '.client_state.chain_id==\"osmosis-test-1\"'"
checkpoint "hub channel-1 -> osmosis"       sh -c "gaiad q ibc channel client-state transfer channel-1 -o json | jq -e '.client_state.chain_id==\"osmosis-test-1\"'"
checkpoint "hub ica host allows MsgTransfer" sh -c "gaiad q interchain-accounts host params -o json | jq -e '.params.allow_messages | index(\"/ibc.applications.transfer.v1.MsgTransfer\") != null or index(\"*\") != null'"
checkpoint "osmo ica host allows MsgSend"    sh -c "osmosisd q interchain-accounts host params -o json | jq -e '.params.allow_messages | index(\"/cosmos.bank.v1beta1.MsgSend\") != null or index(\"*\") != null'"
for h in stride cosmoshub osmosis; do checkpoint "REST $h reachable" sh -c "curl -s -m 10 https://$h-api.internal.stridenet.co/cosmos/base/tendermint/v1beta1/node_info | jq -e '.default_node_info.network'"; done
# Prove the constants: a test transfer in and a signed spend out of the vault and the sweep operator
log_cmd "vault receives"  osmosisd tx bank send user1 $VAULT_MS_OSMO 1000000uosmo $OSMO_TX
sleep 6; ms_tx osmosisd vault-ms m1,m2,m3 -- bank send $VAULT_MS_OSMO $USER1_OSMO 1uosmo
log_cmd "sweep operator spends" strided_old tx bank send sweep-operator $USER1_STRIDE 1ustrd $STRIDE_TX
log_cmd "withdraw addresses" sh -c "for z in cosmoshub-test-1:gaiad osmosis-test-1:osmosisd; do IFS=: read zone chain <<<\$z; d=\$(strided_old q stakeibc show-host-zone \$zone -o json | jq -r .host_zone.delegation_ica_address); \$chain q distribution delegator-withdraw-addr \$d -o json; done"
log "gaia $(gaiad version 2>&1 | tail -1), osmosis $(osmosisd version 2>&1 | tail -1), strided $(strided_old version 2>&1 | tail -1)"
```

The host zones exist only after the seed, so the withdraw-address and multisig lines run at the end of the seed as well; phase 0 is run twice: once right after `make start` (channels, REST, versions) and once after the seed (the rest). Guard the zone-dependent lines with `strided_old q stakeibc show-host-zone cosmoshub-test-1 >/dev/null 2>&1 &&`.

- [ ] **Step 2: Commit** — `git commit -m "rehearsal: pre-flight script"`.

## Execution tasks (serial, run by the orchestrator)

### Task 14: Build the images and start the network

**Depends on:** Tasks 1-13. **Review:** no

- [ ] **Step 1:** `kubectl config use-context integration` and confirm `make -C integration-tests check-empty-namespace` passes (if pods exist, ask before `make stop`; a running network may belong to someone).
- [ ] **Step 2:** From the worktree root with a clean tree (`git status --short` empty; build.sh refuses otherwise): `cd integration-tests && UPGRADE_OLD_VERSION=wind-down-rehearsal-v34 make build-stride-upgrade 2>&1 | tee /tmp/build.log`. This checks out the v34 branch, rewrites the F5 admin to the keys.json admin on both branches, builds both binaries, and pushes `chains/stride:latest`. Expect 20-60 minutes under emulation. Afterwards `git status --short` must be empty and `git branch --show-current` must print `wind-down-rehearsal`.
- [ ] **Step 3:** `make start`; wait for readiness. Then `bash rehearsal/phase0_preflight.sh`. If the channel checkpoints fail because the relayers raced, `make stop`, wait for the namespace to empty, `make start` again.
- [ ] **Step 4:** Log the `kubectl get pods` output and commit the log.

### Task 15: Seed, upgrade, and run phases 2 to 9

**Depends on:** Task 14. **Review:** yes (the orchestrator reviews each checkpoint as it lands)

- [ ] **Step 1:** `bash rehearsal/seed.sh`; rerun phase 0's zone-dependent lines. Watch `state.env` for `U`.
- [ ] **Step 2:** `bash rehearsal/phase1_upgrade.sh` before `U - 60s`.
- [ ] **Step 3:** Phases 2, 3, 4 in order. Phases 5 and 6 as the unbondings allow. Then 7, 8, 9.
- [ ] **Step 4:** On any `CHECKPOINT FAIL`, diagnose before rerunning: read the pod logs (`kubectl logs stride-validator-0 -c validator --since=5m`), the record state, and the relayer logs. Fix the script if the script is wrong; record a finding in the log (`log "FINDING: ..."`) if the code or the runbook is wrong. Do not edit `x/` on this branch to make a checkpoint pass unless the fix is also filed as a finding for main.
- [ ] **Step 5:** After every phase, `git add -A integration-tests/rehearsal docs/wind-down/rehearsal-log.md && git commit -m "rehearsal: phase N log"`.

### Task 16: Summary

**Depends on:** Task 15. **Review:** no

- [ ] **Step 1:** Append a `## Summary` section to `docs/wind-down/rehearsal-log.md`: per phase, PASS/FAIL per checkpoint; the findings list, each tagged `code`, `runbook` or `script`, with the file and spec section it concerns; the gas measurement per sweep batch size; what was not covered and why.
- [ ] **Step 2:** `make stop` only after the user has seen the summary (they may want to poke at the network).
- [ ] **Step 3:** Report the summary to the user in the chat, with the path to the log.
