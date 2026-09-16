# v34 Injective reconciliation — k8s rehearsal results (2026-09-15/16)

Branch `v34-injective-rehearsal` (throwaway). Network: v33.0.0 → v34 (branch) → v35 (rehearsal-only
handler), 3 Stride POA validators + 3 Gaia validators, rly daemon + hermes 1.13.1 CLI pod. Driver:
`integration-tests/rehearsal/injective.sh`; phase logs in the session scratchpad.

## Outcome: every step of the fix verified end to end

| Step | Evidence |
|---|---|
| Incident reproduced (not seeded) | delegate ICA executed on Gaia with its ack stranded (rly denylist + hermes receive-or-timeout), channel closed by the 1ns-timeout ICA, restored; after two redemptions (500 / 400) matured the retrying delegate consumed 300 of the 911.9 unbonded ATOM. Bundled sweep for 910,396,762 failed every stride epoch; on-chain − tracked = **309,643,420** (val3 +104,334,860, cosmoshub1 +104,334,870, val2 +100,973,690) |
| Reconciliation | `v34: reconciled val3 1215184956 -> 1319519816 (104334860)` (×3), `TotalDelegations adjusted by 309643420 to 3955111740`, `Upgrade v34 complete`; RR 1.0146 → 1.1007 (inside 0.9–1.5) |
| Pending undelegation queued and submitted | `Submitting pending undelegation of 309643420uatom` one day epoch after the upgrade |
| Per-record sweep | record 40 (505,775,979) swept → `CLAIMABLE`, redemption ICA = 505,775,979 exactly; record 41 (404,620,783) error-acked against 94,977,363 liquid and kept retrying — no all-or-nothing |
| Undelegation ack books the ledger | `UNDELEGATE ICACALLBACK ICA SUCCESSFUL` (channel-13 seq 187), `Total Burned from Batch 0stuatom`; `TotalDelegations` 3955111740 → **3645468320** = the pre-incident value; val3 / cosmoshub1 tracked == on-chain to the uatom; RR back to **1.014568757872143026** (identical to the pre-upgrade value) |
| Remaining record sweeps and claims | after the 240s host unbonding record 41 → `CLAIMABLE`, redemption ICA = 910,396,762 = R1+R2; both `claim-undelegated-tokens` succeeded; user1's Gaia balance rose by exactly 910,396,762 |

Only ledger/chain difference at the end: val2 −13,160,312 (1.0%) = a downtime slash on the test
validator (jailed 23:04 UTC), which the slash-query machinery owns.

## Findings for the real PR (#1526)

1. **Stranded pending undelegation is lost when the channel dies behind it (fixed on this branch,
   proposed for the PR).** After v34 submitted the undelegation, channel-12 closed on an earlier
   packet's timeout. On an ordered channel the later packets' timeouts are only processable in
   sequence, and a `rebalance` timeout ahead of them fails forever after the restore zeroed the
   in-progress counters (`Unable to call registered ICACallback from OnTimeoutPacket`). The
   undelegate TIMEOUT callback — the re-queue path — therefore never ran and the 309.6 ATOM key was
   gone. `RestoreInterchainAccount` resets records and counters but did not know about it.
   Fix: `Keeper.RequeueStrandedPendingUndelegations` (x/stakeibc/keeper/pending_undelegation.go),
   called from the restore's delegation branch: re-queues every record-less undelegate callback on a
   non-open delegation channel from its stored splits and removes the callback data. Exercised here
   via the rehearsal-only v35 handler: the hook resubmitted 309,643,420 and the flow completed.
2. **Rounding dust.** Gaia returned 309,643,419 for a 309,643,420 undelegation (three truncating
   share conversions). With amounts chosen for exact equality the last record was short by 1 uatom.
   Mainnet has ~4.76 INJ of idle slack in the delegation ICA so it cannot bite there, but the
   release checklist should confirm that slack still exists on upgrade day.
3. **`UpdateGovParams` in the handler** sets the 5-day voting period on whatever chain runs it; any
   post-upgrade governance on a test network must be expedited (29s window).

## Harness lessons (all encoded in the driver / branch)

- rly v2.5.2 `tx relay-packets` is deprecated and crashes; batched `MsgTimeout`s on an ordered
  channel fail proof verification (`next_sequence_recv`); `--max-msgs 1` cannot keep up with the ICA
  packet rate. Timeouts are best driven one per tx from hermes (`max_msg_num = 1`).
- The daemon must never stop: light clients expire 204s after the last update (85% of the 240s
  unbonding). Use `rly start --time-threshold 90s`, and denylist only the channel under study.
- hermes 1.9 cannot parse CometBFT 0.39 events; 1.13.1 with `compat_mode = '0.38'` and
  `max_gas = 4000000` works (an ICA receive + client update simulates at ~830k).
- The host never sees `chan-close-confirm` for a denylisted channel; restore is refused until it is
  sent by hand (`hermes tx chan-close-confirm`), and a stuck `INIT` handshake can be nudged with
  `hermes tx chan-open-try`.
- 30s voting periods: submit and vote from the laptop's strided against the RPC ingress (0.2s per
  call); in-pod CLI calls take ~8s each under the 400m CPU limit, `--gas auto` under-estimates gov
  votes, and per-pod `kubectl exec` votes land after the window closes.
- Expired clients recover via `MsgRecoverClient` gov proposals as long as the substitute is created
  by the same relayer (rly) so trust level / clock drift match, and is fresh when the proposal
  executes.
