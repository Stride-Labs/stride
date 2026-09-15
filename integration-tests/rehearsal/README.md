# v34 Injective reconciliation rehearsal (REHEARSAL ONLY — DO NOT MERGE)

Driver: `injective.sh`. Spec: `docs/superpowers/specs/2026-09-15-v34-injective-rehearsal-design.md`.
Runs from the laptop against context/namespace `integration` (`kctx`), everything through
`kubectl exec`. Needs `kubectl`, `jq`, `docker`; bash 3.2 is fine.

## Phase order

| # | Command | What it does / waits for |
|---|---|---|
| 0 | `UPGRADE_OLD_VERSION=v33.0.0 make build-stride-upgrade && make start` (from `integration-tests/`) | network on v33.0.0 with the branch's v34 under cosmovisor |
| 1 | `rehearsal/injective.sh setup` | keys, `register-host-zone` (unbonding period 7 → frequency 2), `add-validators`, IBC-transfer 3000 ATOM, `liquid-stake`; waits until tracked == on-chain == 3000 ATOM |
| 2 | `rehearsal/injective.sh redeem` | `redeem-stake` 500 stATOM in an even day epoch and 400 in the next (odd) one; waits until both records are `EXIT_TRANSFER_QUEUE` and prints R1 / R2 |
| 3 | `rehearsal/injective.sh drift` | **run immediately after `redeem`** (the channel must be closed before the 240s host unbonding matures, or the bundled sweep succeeds and the incident is not reproduced). Stakes 300, takes the relayer manual, hand-relays the delegate so its ack is stranded, closes the channel via a 1ns-timeout ICA, restores the account, then waits for the stuck state: on-chain == tracked + 300, ICA liquid == R1+R2−300, both records unswept |
| 4 | `rehearsal/injective.sh measure` | prints the Go delta table (on-chain − tracked per validator), Σ delta, ICA liquid, R1 / R2 |
| — | **manual step** | paste the table into `app/upgrades/v34/injective.go` (`InjectiveDelegationDeltas`) and commit |
| 5 | `rehearsal/injective.sh build-and-swap` | `docker build` the v34 binary (keys.json admin baked in, like `build.sh`), `kubectl cp` it to `cosmovisor/upgrades/v34/bin/strided` on all three Stride pods, prints `strided version` from each |
| 6 | `rehearsal/injective.sh upgrade` | `make upgrade-stride`, then asserts the pod-0 log lines (`Starting upgrade v34`, `TotalDelegations adjusted by`, no `NOT applied`, `Upgrade v34 complete`) and that all three pods keep producing blocks. Needs the kube context's default namespace set to `integration` (upgrade.sh has one `kubectl exec` without `-n`) and a TTY (upgrade.sh uses `-it`) |
| 7 | `rehearsal/injective.sh verify` | spec §6 steps 1–5 with evidence: per-record sweep (record 1 `CLAIMABLE`, record 2 requeued, redemption ICA == R1), pending undelegation at the first odd day epoch (log line, Gaia delegations and `TotalDelegations` −Σdelta, ledger == chain per validator), record 2 `CLAIMABLE` after 240s, both claims land on user1's Gaia balance, three quiet day epochs |
| any | `rehearsal/injective.sh status` | host zone, records, ICA balances, Gaia delegations, ICA channel states |
| any | `rehearsal/injective.sh wait-day-epoch [even\|odd]` | block until the next day epoch (with that parity) starts |

Tear down with `make stop`.

## Notes

- Every phase is idempotent where cheap (existing host zone / keys / records are detected and
  skipped) and dies with the offending value otherwise. `POLL_INTERVAL` (default 3s) is
  overridable.
- `setup` and `drift` patch the relayer deployment to `Recreate` with a 1s grace period: the
  default rolling update keeps the old daemon alive for 30s+, which is longer than the window
  between a deposit reaching `DELEGATION_QUEUE` and the delegate ICA at the next stride epoch.
  If the daemon still delivers the delegate + ack before it stops, `drift` says so; re-run it
  (it stakes another 300 ATOM, the delta table is measured afterwards anyway).
- Parity: unbondings run at the start of even day epochs, `SubmitPendingUndelegations` on odd
  ones (unbonding period 7 → frequency 2). `redeem` lines the two redemptions up so both records
  unbond at the very next even epoch.
- The reconciliation bumps RR by Σdelta / TotalDelegations (~12% here vs. negligible on mainnet)
  until the pending undelegation acks. Inner RR bounds are unset at registration, so only the
  outer `[0.9, 1.5]` bounds apply and the zone does not halt.
- Manual-mode `rly` relays `MsgRecvPacket` / `MsgTimeout` only, never acks; the daemon after the
  closure logs errors for the stranded ack (expected). ICQ errors for the missing osmosis
  connection are harmless.
- Grep the driver for `VERIFY:` — CLI details that could not be checked against a live network.
