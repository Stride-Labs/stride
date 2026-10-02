# Wind-Down Dashboard: Ops tab

Status: approved in conversation 2026-10-02. Medium tier. Extends `2026-09-30-wind-down-dashboard-design.md`.

## Goal

A first tab, **Ops**, that is the team's home base for the wind-down: open the dashboard on any
day and see what has to be done right now, what is coming, and what must not be done. Every
step of the ops window from the protocol wind-down spec (§9 with its three checklists, §9c, §8's
Osmosis procedure, and the actionable open items of §12) is on it, dated, as a checklist whose
ticks are committed to the repo so the whole team sees the same state.

## Content: `scripts/wind-down/dashboard/ops/plan.json`

The plan is data, reviewed like the spec. Shape:

```json
{
  "anchors": { "upgrade": "2026-10-12T12:00:00Z", "day_epoch_1500": "2026-10-12T19:00:00Z" },
  "days": [
    {
      "date": "2026-10-12",            // ISO date; "until" (optional) makes a multi-day block
      "until": "2026-10-16",
      "title": "Upgrade day",
      "note": "one line of context",
      "estimated": false,              // true when the date follows from unbonding times rather than a decision
      "avoid": ["things that must not happen in this block"],
      "windows": [
        {
          "label": "12:00 UTC · upgrade height",   // optional; a block with one window may omit it
          "steps": [
            {
              "id": "haqq-close-29",               // stable, unique; the status file keys on it
              "text": "one line, imperative",
              "detail": "optional longer text: how to verify, which command, which REST path",
              "ref": "§9 step 1b",
              "zones": ["celestia", "cosmoshub-4"] // optional: renders one sub-tick per zone, ids `<id>:<zone>`
              "conditional": "only if the vote slips"  // optional: shown as a tag
            }
          ]
        }
      ]
    }
  ]
}
```

Dates come from: the proposal on Monday 2026-10-05 (5-day voting period, so it passes 10-10);
the upgrade height at about 12:00 UTC (8am EDT) on 2026-10-12; day epoch 1500 at 19:00 UTC that
day, which every zone's unbonding frequency (3, 4, 5) divides, so all queued redemptions submit
then and the drains follow in the 19:00–08:00 UTC window; and each host's real unbonding time
for the completion blocks: haqq 7 days, osmosis and celestia (and the staketia multisig, which is
on Celestia) 14, the Hub, injective, band, terra, saga and dYdX 21, juno and sommelier 28. Real
pools are instantiated after the upgrade (the rate freezes there and the factors are fixed at
instantiation); the route denoms are seeded into the vault while the holder-chain relayers are
still up, 10-05 to 10-11.

Completeness is the requirement: every numbered step and sub-step, every checklist bullet and
every "do not" of §9/§9c, the §8 pool procedure, and the §12 items that need an action, appear
once, on the day they apply, with per-zone sub-ticks where the step runs per zone.

## Status: `scripts/wind-down/dashboard/ops/status.json`

`{ "<step id or step id:zone>": { "done": true, "at": "<iso>", "by": "<name>" } }`, written by
the server on `POST /api/ops/check` with body `{ "id", "done", "by" }` (atomic write, keys
sorted, so the diff is one line per tick). Committed by whoever ticks, like any file.
`GET /api/ops` returns `{ "plan", "status", "today" }` with no refresh thread (it is read from
disk on each request, so an edit to the plan shows on reload).

## Page

- The tab is first and the default. Header line: today's date, the current block's title, and
  the next dated block with its countdown.
- One section per block, in date order. Past blocks collapsed with `done/total`; the current
  block open and marked; the next block open; later blocks collapsed. Estimated dates carry a
  `~` and a tooltip.
- Inside a block: the `avoid` list as a tinted "Do not" box at the top, then each window with
  its label and checklist. A step is a checkbox, its text, optional `conditional` tag, the
  `ref`, and a `details` toggle for the long text. Per-zone steps show their sub-ticks indented
  and the parent tick is derived (all zones done).
- Ticking writes through `POST /api/ops/check` with the name from a small "you are" field kept
  in localStorage; the box re-renders from the server's answer. Ticked steps show `by · when`.
- No styling beyond the existing stylesheet's vocabulary.

## Testing

Unit tests: the plan file loads, every id is unique (including zone-expanded ids), every day has
a valid date, every step has `id`/`text`, and the status writer round-trips and sorts keys. The
live check: the tab renders the plan, a tick lands in `status.json` and survives a reload.

## Build plan

One chunk: `ops.py` (load plan, load/write status, today), the two routes in `server.py`,
`static/ops.js`, the tab in `index.html`/`app.js` (first and default), `test_ops.py`, README
section. The plan content is written by the main agent in parallel; the implementer tests
against a two-day sample plan if the real one is not there yet.
