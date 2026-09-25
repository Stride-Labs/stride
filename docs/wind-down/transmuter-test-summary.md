# Transmuter mainnet test: summary

Run on 2026-09-23 on osmosis-1 with real funds, against a throwaway stATOM/ATOM pool. Detailed
per-transaction record: `transmuter-test-log.md`. Contract and route reference: `transmuter.md`.

## What we built

- **Pool 3590**, a transmuter (alloyed pool, code id 996, contract v3.2.0) holding canonical stATOM and
  ATOM at the frozen Stride redemption rate `2.000174393066540432`, funded with 8 ATOM from the admin key.
- Later added the Hub, Agoric and Secret two-hop stATOM denoms so the pool had every foreign route that
  can reach Osmosis today. Ended with 5 pool assets.
- Two keys: an admin/moderator/LP (the "vault") and a stranger.
- Frozen at the end, not emptied. About $23 remains inside, nearly all shares held by the admin key.

## What we proved

Every number the plan predicted from the source matched the chain.

| Area | Result |
|---|---|
| Rate | Spot price exactly RR. 1,000,000 ustatom → 2,000,174 uatom; 2,000,000 uatom costs 999,913 ustatom. Rounding always favours the pool. |
| Funding | Vault joins with ATOM only; 1 alloyed share = 1 uatom of value. Foreign-route denoms swap 1:1 with canonical stATOM and at RR with ATOM. |
| User path A: router | Works through the poolmanager (what the app sends). Taker fee 0.02% on canonical stATOM/ATOM, 0.1% default on two-hop pairs. |
| User path B: contract | `join_pool` with stATOM then `exit_pool` with ATOM. No fee at all. |
| Foreign routes | Hub and Secret two-hop stATOM arrived with the predicted denom hashes and redeemed both ways. Agoric added by denom (supply already existed). |
| Safety | Dust can't leak (1 uatom → 0). Under-funded pool fails cleanly (`Insufficient pool asset`). Pool value is invariant under swaps, so a per-route cap is a hard ceiling. |
| Authority | Every admin and moderator message from the stranger: `Unauthorized`. Freeze blocks swaps, joins and even the admin's exits. Two-step admin hand-over, cancel and reject all work. |
| Levers | Per-route static cap blocks increases past the cap and never blocks the route leaving; the last limiter on a denom is permanent (widen to 1 to disable). Corrupted-asset marking blocks the reverse direction but also blocks the vault's own top-ups. Rescale is uniform only. |
| Alloyed asset | A plain bearer token: whoever holds it can exit. Custody it like the backing. |
| Nobody else showed up | 36+ transactions, zero third-party interaction, because the app's router ignores pools this small. |

## Adversarial pass (day 2)

We then tried to take value out that wasn't put in, or to break redemption. Nothing worked:

- 393 fuzzed quotes across every pair including the alloyed asset: never a unit in the trader's favour.
- Multi-leg routes through the pool (via the alloyed asset, via canonical stATOM) pay more fees and get less.
- Multi-asset joins and exits round each leg against the trader; a three-unit exit burns 7 shares.
- Foreign denom, same denom, alloyed as input, u128-max, over-pool exact-out: all rejected cleanly.
- Tokens bank-sent to the contract are invisible to the pool and unreachable by anyone.
- Joining before the vault funds gives exactly RR, nothing more.
- Admin fat-finger (a denom added with a wrong factor) is survivable: mark it corrupted before anyone
  swaps into it and it is blocked immediately and removed on the next swap.
- Caps cannot be bypassed by join, and they bind the vault's own native exits (widen first).
- Pool value never decreased across ~45 more transactions; it ended 14 uatom ahead of its shares.

## Per-route retest (day 3)

After the design moved to one pool per route, the tests were re-run on three fresh two-asset pools (canonical,
Hub route, Secret route) plus a deliberately inverted pool:

- The check script passed the three good pools and failed the inverted one on four checks before any funds
  went in.
- Isolation is total: a flavour cannot be swapped, joined or requested in any pool but its own.
- Every swap matched `floor((in − fee) × RR)`; the one-transaction join+exit (the hosted-page path) is atomic
  and fee-free, and reverts entirely if the exit asks for one unit more than the join minted.
- **Per-route pools alone don't stop de-hopping**: Hub-stATOM → ATOM → canonical stATOM works across two pools.
  **Marking the native token "corrupted" after funding makes every pool one-way**: ATOM can only leave, so
  stToken purchases and the de-hop are blocked while redemptions keep working. Two costs, both tested: the
  vault must unmark before a top-up, and when a pool's ATOM reaches zero the contract deletes ATOM from the
  pool; recovery is `add_new_assets` with the same factor plus a join, which restores the exact rate.
- Freeze is per pool; authority checks and the rounding fuzz came out as before.

## What we found that changes the plan

1. **Injective is blocked, by an Osmosis bug.** Osmosis's IBC rate-limiter contract compares channel ids
   as string prefixes without a trailing slash. Injective's channel to Osmosis is `channel-8` and to Stride
   `channel-89`, so every Stride-issued token from Injective (stATOM ~$108k, stINJ ~$53k) is rejected on
   arrival, at any amount. No other in-scope pair collides. Fix is one character in the contract; deploying
   it is Osmosis governance. Issue drafted for Osmosis. Fallbacks: Injective holders redeem via Stride in
   window 1, or route Injective → Hub → Osmosis after the halt.
2. **Stride ↔ Secret has no relayer in either direction.** Packets from 20 Sep were still pending. We
   relayed our own with hermes, one packet at a time, because the only public Secret RPC rate-limits hard.
   Window 1 needs us on that pair, ideally with a private Secret node.
3. **The Osmosis app cannot see foreign-route stATOM.** Its router (SQS) calls the two-hop denom "not a
   valid chain denom", so holders can't quote or swap it in the UI even though the chain handles it. They
   need a contract path: CLI instructions or a small page we host.
4. **Osmosis channel-476 to Secret is not a transfer channel** (it's Secret's SNIP-20 bridge contract).
   Bank-held stATOM on Secret has exactly one route, channel-1.
5. **Spec correction, caught before the test:** the normalization factors were inverted in the spec. The
   stToken takes `1e18`, the native token and the alloyed asset take `RR × 1e18`.

Smaller things worth knowing: the app's router needs roughly $40k in a stATOM/ATOM pool before it will
route through it; Osmosis pool creation costs 20 allUSDC, not OSMO; base gas price is 0.03 uosmo; hermes
1.13.2 relays Stride fine with `compat_mode = '0.38'` despite its SDK-version warning; `injectived` has no
macOS build (Docker works).

## Cost and accounting

Reconciled to 1 uatom: start plus inflows minus end equals taker fees plus 8 uatom of pool-favouring
rounding. Spent: 20 allUSDC pool fee, about $0.04 of gas, a few cents of taker fees. Leftovers: ~$21 in the
frozen pool, 5 allUSDC on the admin key, 1 stATOM back on Injective, 0.5 stATOM on Secret, 1 stATOM in
Stride's channel-40 escrow awaiting a timeout relay, 0.5 ATOM on the Hub.

## Recommendations carried into the spec

- One pool per stToken route (decided 2026-09-24 after the test): canonical + native, and a separate
  two-asset pool per foreign-route denom funded at its escrow share. No limiters.
- Mark the native token corrupted in every pool right after funding (one-way pools); unmark to top up;
  if a pool's native drains to zero, re-add it with `add_new_assets` at the same factor.
- Never use the join+exit page from the vault key (it holds funding shares).
- Split moderator (fast hot key) from admin (multisig); admin cannot be renounced.
- Destroy validator consensus keys after the halt (12-day Osmosis client window).
- File the Osmosis rate-limiter bug now; decide Injective holders' path.
- Plan a user-facing path for foreign-route holders (the app won't serve them).
