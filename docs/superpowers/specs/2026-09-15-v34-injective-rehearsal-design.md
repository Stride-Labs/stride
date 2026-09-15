# v34 Injective Reconciliation — k8s Upgrade Rehearsal

**Date:** 2026-09-15
**Status:** Approved
**Branch:** `v34-injective-rehearsal` (off `v34-injective-reconciliation`, PR #1526). **Throwaway — never merge.**
**Companion:** `2026-09-14-v34-injective-reconciliation-design.md` (the fix under rehearsal) and the
v33 precedent `5d4932e62` (`2026-04-30-poa-migration-rehearsal-design.md`).

## §1. Objective

Run the v33.0.0 → v34 upgrade on the k8s integration-tests network against a host zone that is in
the same broken state as `injective-1` on mainnet, and observe the fix end to end under real
consensus, a real relayer and a real ICA host:

1. The reconciliation applies the delta table and queues the excess.
2. The per-record sweep pays the records that fit at the first stride epoch.
3. The day-epoch hook submits the pending undelegation, the relayer delivers it, Gaia executes it,
   and the ack brings the ledger back to exactly the on-chain delegations.
4. After the host unbonding period the remaining record sweeps and the user claims.

The broken state is **reproduced, not seeded**: a delegate ICA executes on the host while its ack
is stranded by a channel closure, the deposit record is reset and re-delegated out of unbonded
redemption funds, and the bundled sweep fails every epoch. That is the mainnet incident.

Non-goals: the POA validator swap (separate rehearsal branch), the RR-offset design that was
rejected, unit tests on this branch (they go red by design), production hygiene of the edits.

## §2. Network

| | |
|---|---|
| Stride | 3 POA validators (`val1..val3`), chain id `stride-test-1`, image `chains/stride:latest` built by `UPGRADE_OLD_VERSION=v33.0.0 make build-stride-upgrade` (cosmovisor: v33.0.0 at genesis, branch v34 under `cosmovisor/upgrades/v34`) |
| Gaia | 3 validators (`numValidators: 3`), chain id `cosmoshub-test-1`, image `chains/cosmoshub:v22.1.0` (Gaia v25.1.0), unbonding 240s |
| Relayer | one go-relayer `stride-cosmoshub` (`relayer:v2.5.2`), links `connection-0`/`channel-0`, completes ICA handshakes, relays all ports |
| Dropped | osmosis, `stride-osmosis` relayer, both hermes pods (not needed; cluster has ~3.5 free cores) |
| Epochs | harness defaults: stride 45s, day 180s, mint 5s. ICA timeout = next epoch start − 20% (`DefaultBufferSize=5`) → ~36s relay window |
| Resources | 400m / 700M per chain pod |

Host zone registered on `connection-0` with **`unbonding-period 7`** so `GetUnbondingFrequency()`
is 2: regular unbondings on even day epochs, `SubmitPendingUndelegations` on odd ones. (With the
harness's usual `1`, frequency is 1 and the hook would defer every epoch.)

## §3. Source and harness edits (all marked `REHEARSAL ONLY — DO NOT MERGE`)

| File | Edit |
|---|---|
| `app/upgrades/v34/injective.go` | `InjectiveChainId = "cosmoshub-test-1"`; delta table replaced by the 3 Gaia validators with deltas measured in §5 (placeholder until then) |
| `x/stakeibc/types/host_zone.go` | `PerRecordSweepChainId = "cosmoshub-test-1"` |
| `app/upgrades/v34/upgrades.go` | `SwapPoaValidators` call removed (outgoing mainnet monikers are absent here; an error would halt the chain) |
| `integration-tests/network/values.yaml` | §2 topology |
| `integration-tests/network/scripts/start-relayer.sh` | `RELAYER_MANUAL=true` → restore keys, ensure path, then `sleep infinity` instead of `rly start`. Toggled with `kubectl set env deployment/relayer-stride-cosmoshub RELAYER_MANUAL=true|-` |
| `integration-tests/rehearsal/injective.sh` | phase driver (§4–§6) |

The driver runs from the laptop against context `integration` / namespace `integration`, doing
everything through `kubectl exec` with the `strided` / `gaiad` / `rly` CLIs in the pods. Keys
(`admin`, `user1`) are recovered into pod keyrings from `network/configs/keys.json`. No TS client.

Helpers: `stride_q`, `gaia_q` (JSON queries on pod 0 of each chain), `stride_tx`/`gaia_tx`
(broadcast + wait for inclusion + check code 0), `rly_manual` (exec into the relayer pod),
`wait_for` (poll a jq predicate with a timeout), `relayer_mode manual|daemon` (set env, wait
for the pod to be ready and, in daemon mode, for the path to report `STATE_OPEN`).

## §4. Pre-upgrade phases (on v33.0.0)

**setup**
1. Recover `admin`/`user1` keys on `stride-validator-0`, `user1` on `cosmoshub-validator-0`.
2. `register-host-zone connection-0 uatom cosmos <ibc-denom> channel-0 7 true --max-messages-per-ica-tx 2 --min-redemption-rate 0.9 --max-redemption-rate 1.5` (admin). Wait until all four ICA addresses are set on the host zone.
3. `add-validators cosmoshub-test-1 <file>` with the 3 Gaia valopers (from `gaiad q staking validators`), weight 10 each, names = monikers.
4. `gaiad tx ibc-transfer transfer transfer channel-0 <user1 stride addr> 3000000000uatom --from user1`; wait for the ibc/uatom balance on Stride.
5. `liquid-stake 3000000000 uatom` (user1). Wait until the deposit record is gone and
   `Σ validator.delegation == Σ gaiad delegations of the delegation ICA == 3000 ATOM`.

**redeem**
1. `redeem-stake 500000000 cosmoshub-test-1 <user1 cosmos addr>` (day epoch N).
2. Wait for the next day epoch; `redeem-stake 400000000 ...` (day epoch N+1) → two records.
3. Wait until both records are `EXIT_TRANSFER_QUEUE` (undelegate acked). Record their
   `native_token_amount` (`R1`, `R2`; ≈ 500/400 × RR). The bundled sweep must NOT run before §4
   drift completes: the records only become sweep-eligible 240s after the undelegation, and the
   drift phase closes the channel before that. If timing slips, redo redeem after the drift.

**drift** (the lost-ack theft)
1. `liquid-stake 300000000 uatom` (user1); wait for its deposit record to reach `DELEGATION_QUEUE`
   (transfer relayed).
2. `relayer_mode manual`.
3. Poll the delegation ICA channel's next-sequence-send; when the delegate packet appears at a
   stride epoch boundary, immediately `rly tx relay-packets stride-cosmoshub <ica-channel>` →
   Gaia executes the delegate (ICA delegations +300); the ack is written on Gaia but never relayed.
   Assert: on-chain +300, tracked unchanged, deposit record `DELEGATION_IN_PROGRESS`.
4. Admin `close-delegation-channel cosmoshub-test-1` → ICA packet with a 1ns timeout. Wait one
   block, `rly tx relay-packets` again → `MsgTimeout` on Stride → channel `STATE_CLOSED`. Assert
   channel closed; the delegate ack is now unrelayable.
5. `relayer_mode daemon`. Admin `restore-interchain-account cosmoshub-test-1 connection-0 <delegation owner>`; wait for a new OPEN delegation channel and a non-empty delegation ICA address. Assert the deposit record is back to `DELEGATION_QUEUE`.
6. Wait ≥240s after the undelegation so `R1+R2` is liquid, then let the next stride epoch run:
   the re-delegate consumes 300 of it and the bundled sweep for `R1+R2` error-acks. Assert the
   stuck state: on-chain == tracked + 300; ICA liquid == `R1+R2 − 300`; both records still
   `EXIT_TRANSFER_QUEUE` after two more sweeps.

## §5. Measure, build, swap the binary

**measure**: for each host zone validator, `delta = gaiad delegation of the ICA − tracked`;
print the Go table (`{Name, Address, Delta: mustInt("...")}`), `Σ delta` (expect 300 ATOM in
uatom, split by weight), and the liquid balance vs records. Paste into `injective.go`, commit.

**build**: `docker build -f Dockerfile --platform linux/amd64 -t core:stride .` (the root
Dockerfile, as `build.sh` does), extract `/usr/local/bin/strided`, and
`kubectl cp` it to `/home/validator/.stride/cosmovisor/upgrades/v34/bin/strided` on all three
Stride pods (`-c validator`), `chmod +x`, and check `strided version` inside the pod. Cosmovisor
only reads the upgrade binary at the upgrade height, so the network keeps running on v33.

## §6. Upgrade and verify

**upgrade**: `make upgrade-stride` (proposal at +45 blocks, votes val1–3, 30s voting). Then
`kubectl logs stride-validator-0 -c validator` must show `Starting upgrade v34`,
`TotalDelegations adjusted by <Σdelta>`, no `NOT applied`, `Upgrade v34 complete`, and blocks
continue on all three pods.

**verify** (each step polls with a timeout and prints the evidence):
1. First stride epoch: record 1 → `CLAIMABLE`, record 2 back to `EXIT_TRANSFER_QUEUE`
   (per-record sweep; one `MsgSend` each); redemption ICA balance == `R1`.
2. First odd day epoch: pod-0 log `Submitting pending undelegation of <Σdelta>`; ICA
   delegations on Gaia drop by Σdelta; after the ack `TotalDelegations` drops by Σdelta and
   **Σ tracked == Σ on-chain** (validator by validator).
3. ~240s later: record 2 → `CLAIMABLE`; redemption ICA balance == `R1+R2`.
4. `claim-undelegated-tokens cosmoshub-test-1 <epoch> <user1 cosmos addr>` for both epochs;
   user1's uatom balance on Gaia rises by `R1+R2` (less nothing — claims are ICA sends).
5. Steady state over 3 more day epochs: no further "Submitting pending undelegation" lines,
   ledger still == chain, RR unchanged from its pre-upgrade value except normal reward accrual.

`status` prints host zone (validators, `TotalDelegations`, RR), unbonding records for the zone,
deposit records, delegation/redemption ICA balances, Gaia delegations of the ICA, and the ICA
channel states.

## §7. Risks

- The 36s manual-relay window: the driver polls every second and `relay-packets` takes a few
  seconds. A miss just times the delegate out (nothing executed on Gaia); reset the deposit
  record via restore if the channel closed, and retry at the next epoch.
- `rly` daemon mode after the closure will keep failing to relay the stranded ack (expected —
  closed channel) and will log errors; ignore them.
- `build.sh` for the upgrade image checks out `v33.0.0` in the working tree — commit everything
  first. It pushes the shared `chains/stride:latest` tag.
- ICQ oracle params still point at `osmosis-test-1`/`connection-1`, which don't exist: expect
  harmless ICQ errors in the logs.

---

## Build plan

### Chunk A — branch edits and driver
**Depends on:** none

- Source edits and harness edits per §3 (main agent, this checkout).
- `integration-tests/rehearsal/injective.sh` per §3–§6 (implementer, worktree): phases as
  subcommands, helpers, evidence printed per assertion, `set -euo pipefail`, no interactive
  prompts, all `kubectl` calls with `-n integration`. Amounts, epochs and addresses as constants
  at the top. `status` and `wait_for` are the building blocks.

### Chunk B — execution
**Depends on:** A

- Build the upgrade image, `make start`, run the phases in order, iterate the driver as the live
  network dictates, capture evidence into `integration-tests/rehearsal/injective-results.md`.
- Tear down with `make stop` when complete.
