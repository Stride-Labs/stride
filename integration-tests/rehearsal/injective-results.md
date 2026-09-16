# v34 Injective reconciliation — k8s rehearsal results

Branch `v34-injective-rehearsal` (throwaway). Network: v33.0.0 → v34 (branch), 3 Stride POA
validators + 3 Gaia validators, rly daemon + hermes 1.13.1 CLI pod. Driver:
`integration-tests/rehearsal/injective.sh`; phase logs in the session scratchpad.

Two runs: run 1 (2026-09-15/16) against the PR's original delete-on-submit design found the
stranded-undelegation gap; run 2 (2026-09-16, rehearsal branch rebased onto #1526 @ `543ef9955`,
the in-flight design) is the result for the final code.

## Run 2 (final code): every step verified end to end

| Step | Evidence |
|---|---|
| Incident reproduced (not seeded) | 300 ATOM delegate ICA executed on Gaia with its ack stranded (rly denylist + hermes receive-or-timeout), channel-1 closed by the 1ns-timeout ICA, restored on channel-7; the reset deposit record re-delegated every stride epoch (insufficient funds) until two redemptions (500 / 400 stATOM → 501,010,934 + 400,808,747 uatom) matured, then consumed 300 of the 901.8 unbonded ATOM (Gaia h=1763). Bundled sweep for 901,819,681 failed every stride epoch against 601.06 liquid; on-chain − tracked = **300,760,825** (val3 +100,253,608, cosmoshub1 +100,253,609, val2 +100,253,608), stable across the 10 min until the upgrade |
| Upgrade (expedited gov, h=1798, day epoch 18) | `v34: reconciled val3 delegation 803513329 -> 903766937 (100253608)` (×3), `v34: cosmoshub-test-1 TotalDelegations adjusted by 300760825 to 2711300823`, `Upgrade v34 complete`; RR 1.004412 → **1.129824** (+12.5%, inside 0.9–1.5) |
| Per-record sweep (first stride epoch, 5:09PM) | two separate redemption callbacks `[10]` and `[11]`: record 10 (501,010,934) → `CLAIMABLE`; record 11 (400,808,747) error-acked and back to `EXIT_TRANSFER_QUEUE`, retried alone every epoch — no all-or-nothing |
| Pending undelegation (day epoch 19 = first non-unbonding epoch, 5:10PM) | `Submitting pending undelegation of 300760825uatom` once; `UNDELEGATE ICACALLBACK Starting undelegate callback for Epoch Unbonding Records: []`, per-validator `Delegation Change: -100253617 / -100253604 / -100253604`, `Total Burned from Batch 0stuatom` (channel-7 seq 414, SUCCESS) |
| Ledger and RR after the ack | `TotalDelegations` 2,711,300,823 → 2,413,112,179 (the pre-incident value plus accrual); RR **1.004793** at the next epoch vs 1.004412 four epochs earlier = the normal ~0.0095%/epoch accrual line, i.e. the reconciliation's RR jump is fully reversed by the ack |
| Second sweep and claims | after the 240s host unbonding record 11 → `CLAIMABLE` (5:15PM); redemption ICA = **901,819,681** = R1+R2 exactly; both `claim-undelegated-tokens` succeeded; user1's Gaia balance rose by exactly 901,819,681 |
| Steady state | `claims` phase: val3 / cosmoshub1 / val2 tracked == on-chain to the uatom (804,977,795 / 804,977,810 / 804,977,795), `TotalDelegations == Σ on-chain` = 2,414,933,400; pending-undelegation submissions still 1 at day epoch 23 (four day epochs after the submission, including two unbonding epochs) |

Not exercised in run 2 (the channel stayed healthy behind the undelegate ICA): the in-flight
counter's FAILURE/TIMEOUT release, the restore clearing it, and the stale-channel guard. Those are
covered by the unit tests in #1526 (`pending_undelegation_test.go`, `icacallbacks_undelegate_test.go`,
`msg_server_test.go`).

Rounding: this time Gaia returned the full 300,760,825 for the undelegation (Gaia unbond amounts
300,609,101 / 300,609,091 / 300,601,489 at h=1542 for the redemptions; delegation ICA dust of
~0.2 ATOM from reinvests covered any share-conversion shortfall). Run 1 saw a 1-uatom shortfall —
see finding 2.

## Run 1 (original design): incident reproduced, one gap found

Same choreography with 300 ATOM / 500 + 400 stATOM. Drift measured **309,643,420** (val3
+104,334,860, cosmoshub1 +104,334,870, val2 +100,973,690); reconciliation booked it (RR 1.0146 →
1.1007); per-record sweep paid record 40 and retried record 41; the undelegation was submitted the
next day epoch — and then channel-12 closed on an earlier packet's timeout with the undelegate ICA
in flight (see finding 1). After the fix (then a restore-time scan, since replaced) the hook
resubmitted 309,643,420, the ack brought `TotalDelegations` to exactly the pre-incident value with
the RR identical to its pre-upgrade value, record 41 swept after the host unbonding, and both claims
paid 910,396,762. Only end difference: val2 −13,160,312 (1.0%) = a downtime slash on the test
validator, which the slash-query machinery owns.

## Findings for the real PR (#1526)

1. **Stranded pending undelegation is lost when the channel dies behind it (fixed in #1526,
   `9708df412`/`543ef9955`).** On an ordered channel the later packets' timeouts are only
   processable in sequence, and a `rebalance` timeout ahead of them fails forever after the restore
   zeroed the in-progress counters. The undelegate TIMEOUT callback therefore never ran and, with
   the original delete-on-submit + re-queue-on-timeout design, the amount was gone. Final design:
   the amount stays stored until a SUCCESS ack consumes it; a per-chain in-flight batch counter
   blocks duplicates; FAILURE/TIMEOUT only release the counter; `RestoreInterchainAccount` clears
   it (one store delete — no scan of the icacallbacks store, restore gas unchanged); a record-less
   ack on a stale channel is ignored. Run 2 exercised the happy path of this design end to end.
2. **Rounding dust.** Run 1: Gaia returned 309,643,419 for a 309,643,420 undelegation (three
   truncating share conversions), leaving the last record short by 1 uatom. Mainnet has ~4.76 INJ
   of idle slack in the delegation ICA so it cannot bite there, but the release checklist should
   confirm that slack still exists on upgrade day.
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
  `hermes tx chan-open-try` — though in run 2 rly completed it on its own within a stride epoch, so
  the restored record can already be `DELEGATION_IN_PROGRESS` again when checked.
- 30s voting periods: submit and vote from the laptop's strided against the RPC ingress (0.2s per
  call); in-pod CLI calls take ~8s each under the 400m CPU limit, `--gas auto` under-estimates gov
  votes, and per-pod `kubectl exec` votes land after the window closes.
- Expired clients recover via `MsgRecoverClient` gov proposals as long as the substitute is created
  by the same relayer (rly) so trust level / clock drift match, and is fresh when the proposal
  executes.
- `kubectl logs | grep -q` under `pipefail` returns 141 (SIGPIPE) once the log is a few MB, which
  read as "line not found" for the upgrade wait in run 2; count matches instead (`grep -c`).
- The post-upgrade pipeline completes in ~7 minutes on this network (sweep at the first stride
  epoch, undelegation at the first odd day epoch, second sweep 240s later), so `verify` must start
  right after `upgrade`; `claims` covers the remaining steps when it did not.
