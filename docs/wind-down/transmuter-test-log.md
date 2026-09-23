# Transmuter mainnet test log (stATOM)

Plan: `docs/superpowers/plans/2026-09-23-transmuter-mainnet-test.md`. Reference: `docs/wind-down/transmuter.md`.
Every tx below was signed from the `transmuter-test-*` keys in the local `test` keyring; amounts are real mainnet funds.

## Address book

| Role | Chain | Key | Address |
|---|---|---|---|
| `$KEY` admin / moderator / LP | osmosis-1 | transmuter-test-1 | osmo1v0694qqq6ztzxvzl807dgq7h3e857hdxvpmdlc |
| `$KEY2` stranger / admin candidate | osmosis-1 | transmuter-test-2 | osmo1yju8w83cc3z2jtqfkawus29yspw4dygecwmyfp |
| `$SKEY` seeder | stride-1 | transmuter-test-1 | stride1v0694qqq6ztzxvzl807dgq7h3e857hdx83gpax |
| Hub forwarder | cosmoshub-4 | transmuter-test-1 | cosmos1v0694qqq6ztzxvzl807dgq7h3e857hdxy6gaf2 |
| Injective forwarder | injective-1 | Keplr (same mnemonic) | to be recorded when used |
| Secret forwarder | secret-4 | Keplr | to be recorded when used |

## Task 1: snapshot (2026-09-23 20:00 UTC, osmosis-1 height 71,182,675)

| Input | Value |
|---|---|
| stATOM redemption rate (`stakeibc/host_zone/cosmoshub-4`) | `2.000174393066540432` |
| stATOM/ATOM taker fee | `0.0002` |
| Pool creation fee | `20000000 factory/osmo147h5x9pcj7lm0cttlaefx6sqq5vdfnmwfcqxkmjd7exqm9gc7grqhr75m0/alloyed/allUSDC` |
| Market, 10 stATOM → ATOM (SQS, pool 1283) | `19797431` uatom, 1.02% under RR |
| `$KEY` balances | 8.838973 ATOM, 25 USDC (noble, `ibc/498A…`), 5.037953 stATOM, 7.903662 OSMO |
| `$KEY2` balances | 1.007161 stATOM, 1 OSMO |
| `$SKEY` balances | 4 stATOM, 5 STRD |
| Hub balances | 0.5 ATOM |

Deviations from the plan's prerequisites: `$KEY` holds 8.84 ATOM instead of 10 (fund with what is
there); `$KEY2` has no ATOM yet (it will receive ATOM from its first swap); the fee token is
noble USDC, not allUSDC, so Task 2 gains a step: join the allUSDC pool with 25 USDC to mint allUSDC.

Overflow arithmetic for the real pool, factors `1e18` (stATOM) and `2000174393066540432` (ATOM):
gcd 16, lcm `1.25e35` (fits Uint128), normalized total for 1.3M stATOM plus its ATOM `3.25e29` (fits),
alloyed minted for all stATOM `2.6e12` (fits). Expected swap math: 1,000,000 ustatom exact-in →
2,000,174 uatom; 2,000,000 uatom exact-out ← 999,913 ustatom.

Decisions: subdenom `stATOMtest`; at the end the pool is frozen (`set_active_status false`) and
not emptied.

## Tx log

| # | Task | What | Tx hash | Height | Result |
|---|---|---|---|---|---|
