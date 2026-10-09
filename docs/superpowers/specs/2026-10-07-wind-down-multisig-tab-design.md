# Wind-down dashboard: Multisig tab, landed checks, time-gated live marks

Date: 2026-10-07. Extends `2026-10-02-wind-down-ops-tab-design.md`; the dashboard lives in
`scripts/wind-down/dashboard/`.

## Why

The drain steps on upgrade day are signed by the F5 protocol-admin multisig
(`stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh`, 2-of-3 legacy-amino, account 91; member keys
FS5 = Sam, FA5 = Aidan, FR5 = Riley). Today the ops step shows one templated command with
`<LIVE_TEST_VALOPER>` and `<PROTOCOL_ADMIN_KEY>` placeholders. On the day three people need the
exact generate / sign / multisign+broadcast commands per zone, each tagged with who runs it, and
the ops step should verify on-chain that the undelegation landed instead of relying on a tick.
Separately, live checks on future steps currently read `✓ ok` / `✗ fail` before the step's time
has come, which is noise; they should read `n/a` until the clock reaches the step's window.

## Decisions (from the 2026-10-07 discussion)

- Txs are done end to end one at a time, online: no pre-assigned sequences (account-sequence
  batching has misbehaved before). Each signer's `tx sign` looks the multisig's sequence up.
- The multisig key name in the keyring is `F5`. On cosmos-sdk v0.54.3 `tx sign --multisig
  <address>` resolves the address through the signer's keyring (`KeyByAddress`, then
  `getMultisigRecord`, then `isMultisigSigner`), so every signer needs the F5 multisig key in
  their keyring as well as their own key, or the command fails with "error getting account from
  keybase". The sign labels say so.
- Default signers: Sam (FS5) and Aidan (FA5); Aidan collects the signatures, multisigns and
  broadcasts. Riley's (FR5) sign command is rendered as the backup for either.
- Scope now: the two drain sets (live-test undelegate, full drain). The transfer-day sets
  (`MsgTransferFromIca` ×4 per zone, `MsgTransferStaketiaClaimBalance`) follow once this shape
  has been used.

## 1. Multisig tab

A fifth tab, after Funds flow, self-polling like Ops (no collector thread, no refresh button
semantics of its own): `GET /api/multisig` is composed by the server from the **validators
cache's current view** plus the plan, so it costs no chain calls of its own and is as fresh as
the Validators snapshot (300 s). Response: `{fetched_at: <validators fetched_at or null>,
data: {sets: [TxSet…]}}`.

### Module `multisig.py` (pure; unit-tested)

```python
MULTISIG_KEY = "F5"
MULTISIG_ADDRESS = config.PROTOCOL_ADMIN
NODE = "https://stride-strd-rpc.polkachu.com:443"
CHAIN_ID = "stride-1"

@dataclass(frozen=True)
class Signer:        tag: str; key: str            # ("Sam", "FS5"), ("Aidan", "FA5"), ("Riley", "FR5")
SIGNERS = (SAM, AIDAN, RILEY); DEFAULT_SIGNERS = (SAM, AIDAN); BROADCASTER = AIDAN

@dataclass(frozen=True)
class Command:       tag: str; label: str; text: str   # tag = who runs it ("Sam", "Aidan", "Riley", "anyone")

@dataclass(frozen=True)
class MultisigTx:
    chain_id: str
    title: str                   # "celestia · live test: mhventures (celestiavaloper1q2k…), 15,861,063 utia"
    ready: bool                  # False with `reason` when the inputs are not known yet
    reason: str | None
    commands: list[Command]      # generate, sign×3 (2 default + backup), multisign+broadcast
    files: list[str]             # the /tmp paths the commands share, for the "share these" note

@dataclass(frozen=True)
class TxSet:
    id: str                      # "live-test-undelegate", "full-drain"
    step_id: str                 # the ops step it belongs to
    title: str
    description: str             # one paragraph: what the tx does, what to watch, and the one-at-a-time rule
    txs: list[MultisigTx]        # one per zone, in config.ZONES order

def tx_sets(validators_data: dict[str, Any] | None) -> list[TxSet]
```

`validators_data` is the validators snapshot's `data` (or None when no snapshot yet: every tx
is `ready=False, reason="waiting for the Validators snapshot"`). Zone with `error` →
`ready=False, reason=<error>`. Live-test tx with `live_test_pick` null → `ready=False,
reason=live_test_reason`. Commands are still rendered for a not-ready tx, with `<LIVE_TEST_VALOPER>`
left as a placeholder, so the shape is visible before the day.

Command texts (live-test, celestia shown; `<F>` = `/tmp/wind-down/live-test-celestia`):

```
# anyone — write the validators file and the unsigned tx
mkdir -p /tmp/wind-down
echo '[{"address": "celestiavaloper1q2kaajedxm0r5xc0twdqz6atap96502d67yjyj", "offset": "0"}]' > <F>.json
strided tx stakeibc undelegate-from-validators celestia <F>.json --from stride1k8c2…azg7jlh --generate-only \
  --chain-id stride-1 --node <NODE> --gas 12000000 --fees 60000ustrd > <F>.unsigned.json

# Sam — sign (online: the multisig's account number and sequence are looked up)
strided tx sign <F>.unsigned.json --multisig stride1k8c2…azg7jlh --from FS5 --chain-id stride-1 --node <NODE> \
  --output-document <F>.FS5.json
# Aidan — sign
… --from FA5 … --output-document <F>.FA5.json
# Riley — backup signer (any two signatures suffice)
… --from FR5 … --output-document <F>.FR5.json

# Aidan — combine and broadcast
strided tx multisign <F>.unsigned.json F5 <F>.FS5.json <F>.FA5.json --chain-id stride-1 --node <NODE> > <F>.signed.json
strided tx broadcast <F>.signed.json --node <NODE> --broadcast-mode sync
```

Full-drain set: same shape with `undelegate-from-validators <chain_id> --all` and the per-zone
gas from the existing `drain-rest` expect text (cosmoshub-4 25M, osmosis-1 and ssc-1 15M,
others 13M), fees = gas × 0.005 ustrd. Files under `/tmp/wind-down/full-drain-<chain_id>`.
The unsigned file and the signature files travel between people (Slack); the tx card says so.

### `static/multisig.js`

`registerSelfPollingTab('multisig', start)`; polls `/api/multisig` every 30 s. Renders each
set as a `<details class="panel" id="set-<id>" open>` with the title, description, a "→ Ops
step" link (`#ops`), and one card per zone: title, ready/not-ready pill with the reason, then
the commands as `<pre class="copy">` blocks each preceded by a person pill (`Sam`, `Aidan`,
`Riley`, `anyone`) and the label. Clicking a block copies it (existing `copyOnClick`). When the
hash is `#multisig/<set-id>` the tab scrolls that set into view on first draw.

### Shell changes (`app.js`, `index.html`, `style.css`)

- Tab button and view for `multisig`; `TAB_NAMES` gains it.
- Hash routing accepts `#<tab>/<suffix>`: the tab is the part before `/`; a `hashchange`
  listener selects the tab so in-page links work without a click handler.
- Person pills: `.who` with a fixed colour per person (any three distinct accents).

## 2. Ops → Multisig link

Step schema gains `"multisig": "<set-id>"`. `stepHtml` renders, after the step text, a link
`→ Multisig tab: <set title>` (`href="#multisig/<set-id>"`); the title comes from
`/api/multisig` when loaded, else the set id. `ops.py` validates in a test that every `multisig`
reference names a set `multisig.tx_sets(None)` produces.

Plan edits:
- `drain-live-test`: text ends with "Commands per zone, with who runs what, are on the Multisig
  tab." `multisig: "live-test-undelegate"`; the templated `command` is dropped (the tab is
  the source); `expect` kept. `auto` added (§3).
- `drain-rest`: likewise, `multisig: "full-drain"`, `command` dropped, `auto` added.

## 3. Landed checks

Validators collector, per zone payload, gains `drained_count`: the number of validators whose
recorded delegation and host delegation are both under one whole token (`10**decimals`, the
threshold `pick_live_test` uses for "funded"), and which have at least one unbonding entry
from the delegation ICA **created after the upgrade**. "Created after the upgrade" ⇔
`entry.completion_time > UPGRADE_TIME + host unbonding_time`, where `UPGRADE_TIME` is a new
`config.UPGRADE_TIME = "2026-10-12T12:00:00Z"` (a test asserts it equals the plan's
`anchors.upgrade`) and the host's unbonding time comes from
`/cosmos/staking/v1beta1/params` (one more optional call per zone; when it or the entries
lookup fails, `drained_count` is `null`). Not exactly zero: a full drain of a validator whose
`SharesToTokensRate < 1` goes through `applySharesRoundingSafety`, which undelegates
`amount - max(1, amount/1e17)`, and the ack callback subtracts only that, so at least one base
unit stays recorded on Stride and as dust on the host. Caveat: a validator that held under a
whole token before the drain and that the day epoch happened to drain post-upgrade also counts.
The payload also gains `funded_count`, the number of registered validators with recorded at
least one whole token. `_unbonding_entries` returns per validator the count
and the latest completion time (a small dataclass) instead of a bare count.

Step checks:
- `drain-live-test`: `auto: {tab: "validators", path: "drained_count", at_least: 1}` — new
  comparator `at_least` in `autoValue` (number ≥ threshold → true, else false; `allow` applies
  as for `equals`).
- `drain-rest`: `auto: {tab: "validators", path: "funded_count", equals: 0}` (nothing left at a
  whole token; the buffer dust does not block it).

Both steps keep their single tick (auto steps render zone marks, not per-zone ticks).

## 4. Time-gated live marks

Window schema gains optional `"start"` (ISO UTC). A step's live check is *active* when
`Date.now() >= Date.parse(window.start)`, or, for a window without `start`, when the server's
`today` (ET date) `>= day.date`. Before that every zone mark renders
`<span class="ops-auto muted" title="applies from <window label or day date>; would read ✓ ok">n/a</span>`
and the summary renders `n/a` (muted) instead of `k/n ok`. Activation never ends: a check
stays live once its window has started. `drawOps` is already re-run on each poll, so the
transition needs no timer.

Windows that get `start` (all others inherit the day):

| day | window | start |
|---|---|---|
| 2026-10-06 | Thursday 2026-10-08 · after the day epoch (1496) | 2026-10-08T19:00:00Z |
| 2026-10-12 | ~8:00am ET · upgrade height | 2026-10-12T12:00:00Z |
| 2026-10-12 | 8:00am-3:00pm ET · the haqq sequence | 2026-10-12T12:00:00Z |
| 2026-10-12 | After the haqq sequence is under way · before 3:00pm ET | 2026-10-12T12:00:00Z |
| 2026-10-12 | 3:00pm ET · day epoch 1500 | 2026-10-12T19:00:00Z |
| 2026-10-12 | 3:00pm ET -> ~4:00am ET · the drain window | 2026-10-12T19:00:00Z |
| 2026-10-12 | ~4:11pm ET · stakedym | 2026-10-12T20:11:00Z |
| 2026-10-20 | staketia records 1472-1488 | 2026-10-19T22:45:00Z |
| 2026-10-27 | STRD · mass undelegation completes | 2026-10-26T12:00:00Z |

## Testing

- `test_multisig.py`: with a fake validators payload (one zone with a pick, one without, one
  with `error`), the live-test set has the exact command strings above (full string equality
  for celestia), the not-ready reasons, and the full-drain gas per zone; `tx_sets(None)` yields
  every zone not ready; set ids are unique.
- `test_validators.py`: `drained_count` counts a validator under a whole token (dust, 1/1) with a
  post-upgrade entry, ignores one with a pre-upgrade entry or a whole token or more, and is `null` when the params
  or entries lookup failed.
- `test_ops.py`: every `multisig` reference resolves to a set id; every `start` parses as an
  ISO timestamp; `at_least` is an int; `config.UPGRADE_TIME == anchors.upgrade`.
- Frontend: headless Chrome screenshots of the Multisig tab and the Ops drain window (marks
  read `n/a` today, since the drain window starts 2026-10-12).

## Build plan

**Chunk 1 — backend** (`multisig.py`, `test_multisig.py`, `validators.py`, `test_validators.py`,
`config.py`, `server.py`). Produces: `multisig.tx_sets`, the `/api/multisig` route (composed
from `CACHES["validators"].view()`), `drained_count` in the validators payload,
`config.UPGRADE_TIME`. Depends on: none.

**Chunk 2 — frontend and plan** (`static/multisig.js`, `static/index.html`, `static/app.js`,
`static/ops.js`, `static/style.css`, `ops/plan.json`, `test_ops.py`). Consumes the
`/api/multisig` shape and `drained_count` as specified here (the chunk can be built against the
spec before chunk 1 lands; the plan tests that touch `multisig.tx_sets` are written to import it
and will pass once merged). Depends on: none (merge after chunk 1).
